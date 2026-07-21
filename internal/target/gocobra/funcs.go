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

func funcMap(model *ir.Model, sources []string) template.FuncMap {
	imports := requestImports(model)
	return template.FuncMap{
		"goVarName": goVarName,
		"goRPCName": func(svc *ir.Service, cmd *ir.Command) string {
			return svc.GoName + cmd.GoName
		},
		"goClientType": func(svc *ir.Service) string { return svc.GoName + "Client" },
		"goRequestType": func(cmd *ir.Command) string {
			if alias, foreign := imports[cmd.Input.GoImportPath]; foreign {
				return alias + "." + cmd.Input.GoName
			}
			return cmd.Input.GoName
		},
		"goBinding": func(f *ir.Flag) pflagBinding {
			switch {
			case f.Map:
				return pflagBinding{"StringArray", "[]string"}
			case f.Repeated:
				return repeatedBindings[f.Bind]
			default:
				return singularBindings[f.Bind]
			}
		},
		"requestImports": func() map[string]string { return imports },
		"sourceHeader":   func() string { return strings.Join(sources, ", ") },
		"isJSONBind":     func(f *ir.Flag) bool { return f.Bind == ir.BindJSON },
		"isStringBind":   func(f *ir.Flag) bool { return f.Bind == ir.BindString },
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
			for _, oneof := range order {
				if len(members[oneof]) >= 2 {
					groups = append(groups, members[oneof])
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

func goVarName(f *ir.Flag) string {
	name := "flag"
	for seg := range strings.SplitSeq(f.ProtoPath, ".") {
		name += strcase.UpperCamelCase(seg)
	}
	return name
}

func requestImports(model *ir.Model) map[string]string {
	pkgName := map[string]string{}
	for _, svc := range model.Services {
		for _, cmd := range svc.Commands {
			if path := cmd.Input.GoImportPath; path != model.FileOptions.GoImportPath {
				pkgName[path] = cmd.Input.GoPackageName
			}
		}
	}
	aliases := map[string]string{}
	taken := map[string]bool{}
	for _, path := range slices.Sorted(maps.Keys(pkgName)) {
		alias := pkgName[path]
		for i := 2; taken[alias]; i++ {
			alias = pkgName[path] + strconv.Itoa(i)
		}
		taken[alias] = true
		aliases[path] = alias
	}
	return aliases
}

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

var repeatedBindings = map[ir.Bind]pflagBinding{
	ir.BindString: {"StringArray", "[]string"},
	ir.BindBool:   {"BoolSlice", "[]bool"},
	ir.BindInt:    {"Int64Slice", "[]int64"},
	ir.BindUint:   {"UintSlice", "[]uint"},
	ir.BindFloat:  {"Float64Slice", "[]float64"},
	ir.BindJSON:   {"StringArray", "[]string"},
}
