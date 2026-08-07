// Package irbuild makes a validated ir.Model from a protogen.File.
package irbuild

import (
	"cmp"
	"go/doc"
	"slices"
	"strings"

	"github.com/braveokafor/protoc-gen-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
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
			GoPackageName: string(file.GoPackageName),
			GoImportPath:  string(file.GoImportPath),
		},
	}

	for _, svc := range file.Services {
		name := strcase.KebabCase(string(svc.Desc.Name()))
		if s, ok := strings.CutSuffix(name, "-service"); ok {
			name = s
		}

		so := serviceOptions(svc.Desc)
		short, long := helpFrom(cmp.Or(so.GetHelp(), cleanComment(string(svc.Comments.Leading))))

		service := &ir.Service{
			FullName:   string(svc.Desc.FullName()),
			ProtoName:  string(svc.Desc.Name()),
			Name:       cmp.Or(so.GetName(), name),
			Aliases:    so.GetAliases(),
			Deprecated: svc.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated(),
			ShortHelp:  short,
			LongHelp:   long,
		}

		for _, m := range svc.Methods {
			request := &ir.MessageRef{
				FullName:     string(m.Input.Desc.FullName()),
				ProtoFile:    m.Input.Desc.ParentFile().Path(),
				ProtoPackage: string(m.Input.Desc.ParentFile().Package()),
				GoImportPath: string(m.Input.GoIdent.GoImportPath),
			}
			request.GoPackageName = string(opts.Files[request.ProtoFile].GoPackageName)

			shape := ir.ShapeUnary
			switch {
			case m.Desc.IsStreamingClient() && m.Desc.IsStreamingServer():
				shape = ir.ShapeBidi
			case m.Desc.IsStreamingClient():
				shape = ir.ShapeClientStream
			case m.Desc.IsStreamingServer():
				shape = ir.ShapeServerStream
			}

			co := commandOptions(m.Desc)
			doc := cmp.Or(co.GetHelp(), cleanComment(string(m.Comments.Leading)))
			// A well-known type's comment documents protobuf, not the caller's API.
			var notes []string
			if _, wellKnown := messageBinds[m.Input.Desc.FullName()]; !wellKnown {
				notes = append(notes, cleanComment(string(m.Input.Comments.Leading)))
			}
			if shape == ir.ShapeClientStream || shape == ir.ShapeBidi {
				notes = append(notes, "Each request body becomes one request.")
			}
			if shape == ir.ShapeServerStream || shape == ir.ShapeBidi {
				notes = append(
					notes,
					"The server can send multiple responses. Each prints as it arrives.",
				)
			}
			short, long := helpFrom(doc, notes...)

			cmd := &ir.Command{
				FullName:  string(m.Desc.FullName()),
				ProtoName: string(m.Desc.Name()),
				Name: cmp.Or(
					co.GetName(),
					strcase.KebabCase(string(m.Desc.Name())),
				),
				Aliases:    co.GetAliases(),
				Deprecated: m.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated(),
				Request:    request,
				Response:   string(m.Output.Desc.FullName()),
				Shape:      shape,
				ShortHelp:  short,
				LongHelp:   long,
			}

			cmd.Params = buildParams(m.Input.Desc, opts)
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
