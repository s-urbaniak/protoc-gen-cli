// Package irbuild lifts a protogen.File into a validated ir.Model.
package irbuild

import (
	"go/doc"
	"slices"
	"strings"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Options configures the IR builder.
type Options struct {
	PluginVersion string // "" means "dev"
	// Files indexes every compilation file by proto path.
	Files               map[string]*protogen.File
	Warn                func(string)
	RequestExpandDepth  int // 0 = only the request's own fields get params
	ResponseExpandDepth int // 0 = only the response's own fields become view fields
}

// Build produces the ir.Model for one proto file.
func Build(file *protogen.File, opts Options) (*ir.Model, error) {
	if opts.PluginVersion == "" {
		opts.PluginVersion = "dev"
	}

	// A request message shared by several RPCs would repeat its warnings.
	warned := map[string]bool{}
	warn := opts.Warn
	opts.Warn = func(msg string) {
		if warned[msg] {
			return
		}
		warned[msg] = true
		warn(msg)
	}

	model := &ir.Model{
		PluginVersion:           opts.PluginVersion,
		ProtoFile:               file.Desc.Path(),
		GeneratedFilenamePrefix: file.GeneratedFilenamePrefix,
		ProtoPackage:            string(file.Desc.Package()),
		FileOptions: ir.FileOptions{
			GoPackageName:    string(file.GoPackageName),
			GoImportPath:     string(file.GoImportPath),
			GoDescriptorName: file.GoDescriptorIdent.GoName,
		},
	}

	for _, svc := range file.Services {
		name := strcase.KebabCase(string(svc.Desc.Name()))
		if s, ok := strings.CutSuffix(name, "-service"); ok {
			name = s
		}

		comment := cleanComment(string(svc.Comments.Leading))
		short, long := helpFrom(comment)

		service := &ir.Service{
			ProtoName: string(svc.Desc.Name()),
			GoName:    svc.GoName,
			Name:      name,
			ShortHelp: short,
			LongHelp:  long,
		}

		for _, m := range svc.Methods {
			request := &ir.Request{
				FullName:     string(m.Input.Desc.FullName()),
				ProtoFile:    m.Input.Desc.ParentFile().Path(),
				ProtoPackage: string(m.Input.Desc.ParentFile().Package()),
				GoName:       m.Input.GoIdent.GoName,
				GoImportPath: string(m.Input.GoIdent.GoImportPath),
			}
			request.GoPackageName = string(opts.Files[request.ProtoFile].GoPackageName)

			clientStreaming := m.Desc.IsStreamingClient()
			serverStreaming := m.Desc.IsStreamingServer()

			doc := cleanComment(string(m.Comments.Leading))
			notes := []string{cleanComment(string(m.Input.Comments.Leading))}
			if clientStreaming {
				notes = append(notes, "Reads JSON requests from stdin, one after another.")
			}
			if serverStreaming {
				notes = append(
					notes,
					"The server may send multiple responses; each prints as it arrives.",
				)
			}
			short, long := helpFrom(doc, notes...)

			cmd := &ir.Command{
				ProtoName:       string(m.Desc.Name()),
				GoName:          m.GoName,
				Name:            strcase.KebabCase(string(m.Desc.Name())),
				Input:           request,
				Output:          string(m.Output.Desc.FullName()),
				ClientStreaming: clientStreaming,
				ServerStreaming: serverStreaming,
				ShortHelp:       short,
				LongHelp:        long,
			}

			if !clientStreaming {
				cmd.Params = buildParams(m.Input.Desc, opts)
			}
			cmd.View = buildView(m.Output.Desc, opts)

			service.Commands = append(service.Commands, cmd)
		}

		model.Services = append(model.Services, service)
	}

	if err := model.Validate(); err != nil {
		return nil, err
	}

	return model, nil
}

// short is primary's first sentence; long is all docs joined, "" when it adds nothing to short.
func helpFrom(primary string, extra ...string) (short, long string) {
	short = new(doc.Package).Synopsis(strings.TrimSpace(primary))
	parts := slices.DeleteFunc(
		append([]string{primary}, extra...),
		func(s string) bool { return s == "" },
	)
	if long = strings.Join(parts, "\n\n"); long == short {
		long = ""
	}
	return short, long
}

func cleanComment(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(strings.TrimPrefix(l, " "), " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// isExpandable reports whether fd's sub-fields expand into entries of their own.
func isExpandable(fd protoreflect.FieldDescriptor) bool {
	if fd.Kind() != protoreflect.MessageKind || fd.IsList() || fd.IsMap() {
		return false
	}
	_, wellKnown := messageBinds[fd.Message().FullName()]
	return !wellKnown
}
