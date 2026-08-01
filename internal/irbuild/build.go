// Package irbuild makes a validated ir.Model from a protogen.File.
package irbuild

import (
	"cmp"
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
	PluginVersion string // An empty string means "dev".
	// Files contains every compilation file, with the proto path as the key.
	Files               map[string]*protogen.File
	Warn                func(string)
	RequestExpandDepth  int // At zero, only the request's own fields get params.
	ResponseExpandDepth int // At zero, only the response's own fields become view fields.
}

// Build makes the ir.Model for one proto file.
func Build(file *protogen.File, opts Options) (*ir.Model, error) {
	if opts.PluginVersion == "" {
		opts.PluginVersion = "dev"
	}

	// A request message that several RPCs share can repeat its warnings.
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

		so := serviceOptions(svc.Desc)
		service := &ir.Service{
			ProtoName: string(svc.Desc.Name()),
			GoName:    svc.GoName,
			Name:      cmp.Or(so.GetName(), name),
			Aliases:   so.GetAliases(),
			Hidden:    so.GetHidden(),
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

			co := commandOptions(m.Desc)
			cmd := &ir.Command{
				ProtoName: string(m.Desc.Name()),
				GoName:    m.GoName,
				Name: cmp.Or(
					co.GetName(),
					strcase.KebabCase(string(m.Desc.Name())),
				),
				Aliases:         co.GetAliases(),
				Hidden:          co.GetHidden(),
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
			cmd.ExampleJSON = buildExample(m.Input.Desc, opts)

			service.Commands = append(service.Commands, cmd)
		}

		model.Services = append(model.Services, service)
	}

	if err := model.Validate(); err != nil {
		return nil, err
	}

	return model, nil
}

// long is empty when it would only repeat short.
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

// isExpandable tells if the sub-fields of fd expand into their own entries.
func isExpandable(fd protoreflect.FieldDescriptor) bool {
	return isNested(fd) && !fd.IsList() && !fd.IsMap()
}

func isNested(fd protoreflect.FieldDescriptor) bool {
	if fd.Kind() != protoreflect.MessageKind {
		return false
	}
	_, wellKnown := messageBinds[fd.Message().FullName()]
	return !wellKnown
}
