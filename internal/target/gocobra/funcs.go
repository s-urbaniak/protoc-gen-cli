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

func funcMap(model *ir.Model) template.FuncMap {
	ident := "cli" + strings.TrimPrefix(model.FileOptions.GoDescriptorName, "File")
	aliases := requestImports(model)
	anyCommand := func(pred func(*ir.Command) bool) bool {
		return slices.ContainsFunc(model.Services, func(svc *ir.Service) bool {
			return slices.ContainsFunc(svc.Commands, pred)
		})
	}
	return template.FuncMap{
		"fileIdent":    func() string { return ident },
		"goVarName":    goVarName,
		"goClientType": func(svc *ir.Service) string { return svc.GoName + "Client" },
		"goRequestType": func(cmd *ir.Command) string {
			if alias, foreign := aliases[cmd.Input.GoImportPath]; foreign {
				return alias + "." + cmd.Input.GoName
			}
			return cmd.Input.GoName
		},
		"goBinding": func(f *ir.Param) pflagBinding {
			switch {
			case f.Map:
				return pflagBinding{"StringArray", "[]string"}
			case f.Repeated:
				return repeatedBindings[f.Bind]
			default:
				return singularBindings[f.Bind]
			}
		},
		"imports": func() map[string]string {
			rows := maps.Clone(templateImports)
			for path, alias := range aliases {
				rows[alias] = path
			}
			return rows
		},
		"responseViews": func(svc *ir.Service) []*ir.View {
			byName := map[string]*ir.View{}
			for _, cmd := range svc.Commands {
				byName[cmd.View.FullName] = cmd.View
			}
			views := make([]*ir.View, 0, len(byName))
			for _, name := range slices.Sorted(maps.Keys(byName)) {
				views = append(views, byName[name])
			}
			return views
		},
		"hasRequestDocs": func(svc *ir.Service) bool {
			return slices.ContainsFunc(
				svc.Commands,
				func(c *ir.Command) bool { return !c.ClientStreaming },
			)
		},
		"anyRequestDocs": func() bool {
			return anyCommand(func(c *ir.Command) bool { return !c.ClientStreaming })
		},
		"anySingleResponse": func() bool {
			return anyCommand(func(c *ir.Command) bool { return !c.ServerStreaming })
		},
		"anyClientStreaming": func() bool {
			return anyCommand(func(c *ir.Command) bool { return c.ClientStreaming })
		},
		"anyBidi": func() bool {
			return anyCommand(
				func(c *ir.Command) bool { return c.ClientStreaming && c.ServerStreaming },
			)
		},
		"isJSONBind":   func(f *ir.Param) bool { return f.Bind == ir.BindJSON },
		"isStringBind": func(f *ir.Param) bool { return f.Bind == ir.BindString },
		"oneofGroups": func(cmd *ir.Command) [][]string {
			var order []string
			members := map[string][]string{}
			for _, f := range cmd.Params {
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
		"requiredParams": func(cmd *ir.Command) []*ir.Param {
			var req []*ir.Param
			for _, f := range cmd.Params {
				if f.Required {
					req = append(req, f)
				}
			}
			return req
		},
		"quoteJoin": func(names []string) string {
			quoted := make([]string, len(names))
			for i, n := range names {
				quoted[i] = strconv.Quote(n)
			}
			return strings.Join(quoted, ", ")
		},
		"flagUsage": flagUsage,
	}
}

func goVarName(f *ir.Param) string {
	name := "flag"
	for seg := range strings.SplitSeq(f.ProtoPath, ".") {
		name += strcase.UpperCamelCase(seg)
	}
	return name
}

func flagUsage(f *ir.Param) string {
	// pflag reads a back-quoted word as the value name.
	desc := strings.ReplaceAll(f.ShortHelp, "`", "")

	var hints []string
	if f.Required {
		hints = append(hints, "required")
	}
	if len(f.EnumValues) > 0 {
		hints = append(hints, "values: "+strings.Join(f.EnumValues, " | "))
	}
	if len(hints) == 0 {
		return desc
	}

	marker := "(" + strings.Join(hints, "; ") + ")"
	if desc == "" {
		return marker
	}
	return desc + " " + marker
}

// templateImports is the template's import block, from name to path.
// goimports drops the entries that a file does not use. Request aliases must
// not take these names.
var templateImports = map[string]string{
	"bytes":     "bytes",
	"context":   "context",
	"json":      "encoding/json",
	"errors":    "errors",
	"fmt":       "fmt",
	"io":        "io",
	"maps":      "maps",
	"os":        "os",
	"slices":    "slices",
	"strings":   "strings",
	"table":     "github.com/jedib0t/go-pretty/v6/table",
	"cobra":     "github.com/spf13/cobra",
	"pflag":     "github.com/spf13/pflag",
	"gjson":     "github.com/tidwall/gjson",
	"sjson":     "github.com/tidwall/sjson",
	"term":      "golang.org/x/term",
	"grpc":      "google.golang.org/grpc",
	"protojson": "google.golang.org/protobuf/encoding/protojson",
	"proto":     "google.golang.org/protobuf/proto",
	"yaml":      "sigs.k8s.io/yaml",
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
	for name := range templateImports {
		taken[name] = true
	}
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
