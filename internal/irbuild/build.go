// Package irbuild lifts a protogen.File into a validated ir.Model.
package irbuild

import (
	"go/doc"
	"slices"
	"strings"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

// Options configures the IR builder.
type Options struct {
	PluginVersion string // "" means "dev"
	// Files, keyed by proto path, locates the file declaring a command's request message.
	Files map[string]*protogen.File
	// Warn receives generation warnings.
	Warn func(string)
	// RequestExpandDepth is how many message levels down fields still
	// get flags: a field of a message field derives a dotted flag like
	// --book.title. 0 stops at the request's own fields.
	RequestExpandDepth int
}

// Build produces the ir.Model for one proto file.
func Build(file *protogen.File, opts Options) (*ir.Model, error) {
	if opts.PluginVersion == "" {
		opts.PluginVersion = "dev"
	}

	model := &ir.Model{
		PluginVersion: opts.PluginVersion,
		ProtoFile:     file.Desc.Path(),
		ProtoPackage:  string(file.Desc.Package()),
		FileOptions: ir.FileOptions{
			GoPackageName: string(file.GoPackageName),
			GoImportPath:  string(file.GoImportPath),
		},
	}

	for _, svc := range file.Services {
		name := strcase.KebabCase(string(svc.Desc.Name()))
		if s, ok := strings.CutSuffix(name, "-service"); ok && s != "" {
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
			if f := opts.Files[request.ProtoFile]; f != nil {
				request.GoPackageName = string(f.GoPackageName)
			}

			doc := cleanComment(string(m.Comments.Leading))
			reqDoc := cleanComment(string(m.Input.Comments.Leading))
			short, long := helpFrom(doc, reqDoc)

			cmd := &ir.Command{
				ProtoName: string(m.Desc.Name()),
				GoName:    m.GoName,
				Name:      strcase.KebabCase(string(m.Desc.Name())),
				Input:     request,
				Output:    string(m.Output.Desc.FullName()),
				ShortHelp: short,
				LongHelp:  long,
			}

			cmd.Flags = buildFlags(m.Input.Desc, opts)

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
