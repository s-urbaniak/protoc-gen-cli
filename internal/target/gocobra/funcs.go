package gocobra

import (
	"maps"
	"slices"
	"text/template"

	"github.com/braveokafor/proto-to-cli/internal/ir"
)

// A goImport is one foreign package imported for a request type.
type goImport struct{ Alias, Path string }

// funcMap returns the helper-function map the template renders with.
func funcMap(model *ir.Model) template.FuncMap {
	return template.FuncMap{
		"goRPCName": func(svc *ir.Service, cmd *ir.Command) string {
			return svc.GoName + cmd.GoName
		},
		"goClientType": func(svc *ir.Service) string { return svc.GoName + "Client" },
		// The qualifier must match the alias requestImports emits.
		"goRequestType": func(cmd *ir.Command) string {
			if cmd.Input.GoImportPath == model.FileOptions.GoImportPath {
				return cmd.Input.GoName
			}
			return cmd.Input.GoPackageName + "." + cmd.Input.GoName
		},
		// One aliased import per foreign request-type package, sorted by
		// path. The alias is protogen's package name for the defining file.
		"requestImports": func() []goImport {
			names := map[string]string{}
			for _, svc := range model.Services {
				for _, cmd := range svc.Commands {
					if p := cmd.Input.GoImportPath; p != model.FileOptions.GoImportPath {
						names[p] = cmd.Input.GoPackageName
					}
				}
			}
			out := make([]goImport, 0, len(names))
			for _, p := range slices.Sorted(maps.Keys(names)) {
				out = append(out, goImport{Alias: names[p], Path: p})
			}
			return out
		},
	}
}
