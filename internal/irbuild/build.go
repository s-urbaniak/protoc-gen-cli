// Package irbuild lifts a protogen.File into a validated ir.Model.
package irbuild

import (
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
			GoPackage:    string(file.GoPackageName),
			GoImportPath: string(file.GoImportPath),
		},
	}

	for _, svc := range file.Services {
		model.Services = append(model.Services, buildService(svc, &opts))
	}

	return model, nil
}

func buildService(svc *protogen.Service, opts *Options) *ir.Service {
	serviceName := strcase.KebabCase(string(svc.Desc.Name()))
	if s, ok := strings.CutSuffix(serviceName, "-service"); ok && s != "" {
		serviceName = s
	}

	out := &ir.Service{
		ProtoName:      string(svc.Desc.Name()),
		Name:           serviceName,
		LeadingComment: cleanComment(string(svc.Comments.Leading)),
	}

	for _, m := range svc.Methods {
		out.Commands = append(out.Commands, buildCommand(m, opts))
	}

	return out
}

func buildCommand(m *protogen.Method, opts *Options) *ir.Command {
	typeRef := &ir.TypeRef{
		FullName:     string(m.Input.Desc.FullName()),
		ProtoFile:    m.Input.Desc.ParentFile().Path(),
		ProtoPackage: string(m.Input.Desc.ParentFile().Package()),
		GoName:       m.Input.GoIdent.GoName,
		GoImportPath: string(m.Input.GoIdent.GoImportPath),
	}
	if f := opts.Files[typeRef.ProtoFile]; f != nil {
		typeRef.GoPackageName = string(f.GoPackageName)
	}

	cmd := &ir.Command{
		ProtoName:      string(m.Desc.Name()),
		Name:           strcase.KebabCase(string(m.Desc.Name())),
		LeadingComment: cleanComment(string(m.Comments.Leading)),
		Input:          typeRef,
		Output:         string(m.Output.Desc.FullName()),
	}

	return cmd
}

func cleanComment(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(strings.TrimPrefix(l, " "), " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
