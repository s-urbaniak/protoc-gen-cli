package gocobra

import (
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/stoewer/go-strcase"
)

// A goImport is one foreign package imported for a request type.
type goImport struct{ Alias, Path string }

// funcMap returns the helper-function map the template renders with.
func funcMap(model *ir.Model) template.FuncMap {
	return template.FuncMap{
		"goRPCName": func(svc *ir.Service, cmd *ir.Command) string {
			return svc.GoName + cmd.GoName
		},
		// The local variable backing f's flag.
		"goVarName": func(f *ir.Flag) string {
			name := "flag"
			for seg := range strings.SplitSeq(f.ProtoPath, ".") {
				name += strcase.UpperCamelCase(seg)
			}
			return name
		},
		"goClientType": func(svc *ir.Service) string { return svc.GoName + "Client" },
		"goBinding": func(f *ir.Flag) pflagBinding {
			switch {
			// pflag only accumulates a map flag's key=value entries;
			// the RunE arm parses them.
			case f.Map:
				return pflagBinding{"StringArray", "[]string"}
			case f.Repeated:
				return repeatedBindings[f.Bind]
			default:
				return singularBindings[f.Bind]
			}
		},
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
			aliases := map[string]string{}
			for _, svc := range model.Services {
				for _, cmd := range svc.Commands {
					if p := cmd.Input.GoImportPath; p != model.FileOptions.GoImportPath {
						aliases[p] = cmd.Input.GoPackageName
					}
				}
			}
			out := make([]goImport, 0, len(aliases))
			for _, p := range slices.Sorted(maps.Keys(aliases)) {
				out = append(out, goImport{Alias: aliases[p], Path: p})
			}
			return out
		},
		"isJSONBind":   func(f *ir.Flag) bool { return f.Bind == ir.BindJSON },
		"isStringBind": func(f *ir.Flag) bool { return f.Bind == ir.BindString },
		// Member flag names per oneof.
		"oneofGroups": func(cmd *ir.Command) [][]string {
			var order []string
			members := map[string][]string{}
			for _, f := range cmd.Flags {
				if f.Oneof == "" {
					continue
				}
				if _, seen := members[f.Oneof]; !seen {
					order = append(order, f.Oneof)
				}
				members[f.Oneof] = append(members[f.Oneof], f.Name)
			}
			var groups [][]string
			for _, k := range order {
				if len(members[k]) >= 2 {
					groups = append(groups, members[k])
				}
			}
			return groups
		},
		"quoteJoin": func(names []string) string {
			quoted := make([]string, len(names))
			for i, n := range names {
				quoted[i] = strconv.Quote(n)
			}
			return strings.Join(quoted, ", ")
		},
	}
}

// A pflagBinding holds the pflag setter stem and Go type for one bind.
type pflagBinding struct {
	Setter string
	GoType string
}

var singularBindings = map[ir.Bind]pflagBinding{
	ir.BindString: {"String", "string"},
	ir.BindBool:   {"Bool", "bool"},
	ir.BindInt:    {"Int64", "int64"},
	ir.BindUint:   {"Uint64", "uint64"},
	ir.BindFloat:  {"Float64", "float64"},
	ir.BindJSON:   {"String", "string"},
}

// StringSlice splits an argument on commas; pflag ships no Uint64Slice.
var repeatedBindings = map[ir.Bind]pflagBinding{
	ir.BindString: {"StringArray", "[]string"},
	ir.BindBool:   {"BoolSlice", "[]bool"},
	ir.BindInt:    {"Int64Slice", "[]int64"},
	ir.BindUint:   {"UintSlice", "[]uint"},
	ir.BindFloat:  {"Float64Slice", "[]float64"},
	ir.BindJSON:   {"StringArray", "[]string"},
}
