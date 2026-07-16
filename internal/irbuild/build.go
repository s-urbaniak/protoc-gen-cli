// Package irbuild lifts a protogen.File into a validated ir.Model.
package irbuild

import (
	"go/doc"
	"strings"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

// Options configures the IR builder.
type Options struct {
	PluginVersion string
	// Files indexes every file in the invocation by path.
	Files map[string]*protogen.File
}

// Build produces an *ir.Model for a proto file.
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

		service := &ir.Service{
			ProtoName: string(svc.Desc.Name()),
			Name:      name,
			ShortHelp: shortDoc(comment),
		}
		if comment != service.ShortHelp {
			service.LongHelp = comment
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
			cmd := &ir.Command{
				ProtoName: string(m.Desc.Name()),
				Name:      strcase.KebabCase(string(m.Desc.Name())),
				Input:     request,
				Output:    string(m.Output.Desc.FullName()),
				ShortHelp: shortDoc(doc),
			}

			reqDoc := cleanComment(string(m.Input.Comments.Leading))
			if doc != cmd.ShortHelp || reqDoc != "" {
				parts := make([]string, 0, 2)
				for _, p := range []string{doc, reqDoc} {
					if p != "" {
						parts = append(parts, p)
					}
				}
				cmd.LongHelp = strings.Join(parts, "\n\n")
			}

			service.Commands = append(service.Commands, cmd)
		}

		model.Services = append(model.Services, service)
	}

	return model, nil
}

// shortDoc returns the first sentence of a comment.
func shortDoc(s string) string {
	return new(doc.Package).Synopsis(strings.TrimSpace(s))
}

func cleanComment(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(strings.TrimPrefix(l, " "), " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
