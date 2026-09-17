package gocobra

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/braveokafor/protoc-gen-cli/internal/ir"
	"google.golang.org/protobuf/compiler/protogen"
)

func funcMap(
	file *protogen.File,
	model *ir.Model,
	client string,
	files map[string]*protogen.File,
) template.FuncMap {
	// Two generated files can share a Go package.
	ident := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, model.ProtoFile)

	// Proto full names are unique across services, methods and messages.
	goNames := map[string]string{}
	types := map[string]protogen.GoIdent{}
	methods := map[string]*protogen.Method{}
	for _, svc := range file.Services {
		goNames[string(svc.Desc.FullName())] = svc.GoName
		for _, m := range svc.Methods {
			goNames[string(m.Desc.FullName())] = m.GoName
			goNames[string(m.Input.Desc.FullName())] = m.Input.GoIdent.GoName
			types[string(m.Input.Desc.FullName())] = m.Input.GoIdent
			types[string(m.Output.Desc.FullName())] = m.Output.GoIdent
			methods[string(m.Desc.FullName())] = m
		}
	}

	rows := maps.Clone(templateImports)
	foreign := map[string]string{}
	packageNames := map[string]string{}
	for _, candidate := range files {
		packageNames[string(candidate.GoImportPath)] = string(candidate.GoPackageName)
	}
	for _, svc := range model.Services {
		for _, cmd := range svc.Commands {
			for _, message := range []string{cmd.Request.FullName, cmd.Response} {
				ident := types[message]
				if importPath := string(
					ident.GoImportPath,
				); importPath != "" &&
					importPath != model.FileOptions.GoImportPath {
					if _, seen := foreign[importPath]; !seen {
						foreign[importPath] = goPackageName(importPath, cmd, packageNames)
					}
				}
			}
		}
	}
	aliases := map[string]string{}
	for _, path := range slices.Sorted(maps.Keys(foreign)) {
		alias := foreign[path]
		for i := 2; rows[alias] != ""; i++ {
			alias = foreign[path] + strconv.Itoa(i)
		}
		rows[alias] = path
		aliases[path] = alias
	}

	byMessage := map[string]*ir.View{}
	for _, svc := range model.Services {
		for _, cmd := range svc.Commands {
			byMessage[cmd.View.FullName] = cmd.View
		}
	}
	views := make([]*ir.View, 0, len(byMessage))
	for _, name := range slices.Sorted(maps.Keys(byMessage)) {
		views = append(views, byMessage[name])
	}

	return template.FuncMap{
		"goBinding":    goBinding,
		"goDoc":        goDoc,
		"flagUsage":    flagUsage,
		"oneofGroups":  oneofGroups,
		"prefix":       func() string { return "cli_" + ident + "_" },
		"export":       func() string { return "Cli_" + ident + "_" },
		"imports":      func() map[string]string { return rows },
		"messageViews": func() []*ir.View { return views },
		"goName":       func(fullName string) string { return goNames[fullName] },
		"goRequestType": func(cmd *ir.Command) string {
			return goType(types[cmd.Request.FullName], aliases)
		},
		"goResponseType": func(cmd *ir.Command) string {
			return goType(types[cmd.Response], aliases)
		},
		"clientIs": func(name string) bool { return client == name },
		"connectClientInterface": func(svc *ir.Service) string {
			return connectClientInterface(goNames[svc.FullName], svc, methods, aliases)
		},
		"quoteJoin": func(names []string) string {
			quoted := make([]string, len(names))
			for i, n := range names {
				quoted[i] = strconv.Quote(n)
			}
			return strings.Join(quoted, ", ")
		},
		"fileHasShape": func(shapes ...ir.Shape) bool {
			for _, svc := range model.Services {
				for _, cmd := range svc.Commands {
					if slices.Contains(shapes, cmd.Shape) {
						return true
					}
				}
			}
			return false
		},
	}
}

func goPackageName(importPath string, cmd *ir.Command, packageNames map[string]string) string {
	// The IR records the request package, including managed-mode rewrites. The
	// full descriptor set supplies the exact name for every other package, so
	// a response can live in a different Go package without guessing from its
	// import path.
	if importPath == cmd.Request.GoImportPath {
		return cmd.Request.GoPackageName
	}
	return packageNames[importPath]
}

func goType(ident protogen.GoIdent, aliases map[string]string) string {
	name := ident.GoName
	if alias, ok := aliases[string(ident.GoImportPath)]; ok {
		return alias + "." + name
	}
	return name
}

// connectClientInterface returns source for the per-service interface that a
// generated connect-go client satisfies. The CLI file shares the protobuf
// package with the messages, so importing its generated child connect package
// would create an import cycle.
func connectClientInterface(
	goName string,
	svc *ir.Service,
	methods map[string]*protogen.Method,
	aliases map[string]string,
) string {
	var b strings.Builder
	if goName == "" {
		return ""
	}
	fmt.Fprintf(
		&b,
		"// %sCLIClient is the subset of the generated connect-go client used by this command.\n",
		goName,
	)
	fmt.Fprintf(&b, "type %sCLIClient interface {\n", goName)
	for _, cmd := range svc.Commands {
		method := methods[cmd.FullName]
		if method == nil {
			continue
		}
		req := goType(method.Input.GoIdent, aliases)
		res := goType(method.Output.GoIdent, aliases)
		switch cmd.Shape {
		case ir.ShapeUnary:
			fmt.Fprintf(
				&b,
				"\t%s(context.Context, *connect.Request[%s]) (*connect.Response[%s], error)\n",
				method.GoName,
				req,
				res,
			)
		case ir.ShapeServerStream:
			fmt.Fprintf(
				&b,
				"\t%s(context.Context, *connect.Request[%s]) (*connect.ServerStreamForClient[%s], error)\n",
				method.GoName,
				req,
				res,
			)
		case ir.ShapeClientStream:
			fmt.Fprintf(
				&b,
				"\t%s(context.Context) *connect.ClientStreamForClient[%s, %s]\n",
				method.GoName,
				req,
				res,
			)
		case ir.ShapeBidi:
			fmt.Fprintf(
				&b,
				"\t%s(context.Context) *connect.BidiStreamForClient[%s, %s]\n",
				method.GoName,
				req,
				res,
			)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

type pflagBinding struct {
	// Method is the pflag method stem: String gives StringP and GetString.
	Method string
	Zero   string
	// Overlay is the command.flag arm that sets the value in the request JSON.
	Overlay string
}

var singularBindings = map[ir.Bind]pflagBinding{
	ir.BindString:    {"String", `""`, "value"},
	ir.BindBool:      {"Bool", "false", "value"},
	ir.BindInt:       {"Int64", "0", "value"},
	ir.BindUint:      {"Uint64", "0", "value"},
	ir.BindFloat:     {"Float64", "0", "value"},
	ir.BindJSON:      {"String", `""`, "json"},
	ir.BindList:      {"String", `""`, "json"},
	ir.BindAny:       {"String", `""`, "json"},
	ir.BindBytes:     {"String", `""`, "value"},
	ir.BindTimestamp: {"String", `""`, "value"},
	ir.BindDuration:  {"String", `""`, "value"},
	ir.BindFieldMask: {"String", `""`, "value"},
}

var repeatedBindings = map[ir.Bind]pflagBinding{
	ir.BindString:    {"StringArray", "nil", "value"},
	ir.BindBool:      {"BoolSlice", "nil", "value"},
	ir.BindInt:       {"Int64Slice", "nil", "value"},
	ir.BindUint:      {"UintSlice", "nil", "value"},
	ir.BindFloat:     {"Float64Slice", "nil", "value"},
	ir.BindJSON:      {"StringArray", "nil", "jsonArray"},
	ir.BindList:      {"StringArray", "nil", "jsonArray"},
	ir.BindAny:       {"StringArray", "nil", "jsonArray"},
	ir.BindBytes:     {"StringArray", "nil", "value"},
	ir.BindTimestamp: {"StringArray", "nil", "value"},
	ir.BindDuration:  {"StringArray", "nil", "value"},
	ir.BindFieldMask: {"StringArray", "nil", "value"},
}

func goBinding(param *ir.Param) pflagBinding {
	switch {
	case param.Map && isStringBind(param):
		return pflagBinding{"StringArray", "nil", "mapString"}
	case param.Map:
		return pflagBinding{"StringArray", "nil", "mapJSON"}
	case param.Repeated:
		return repeatedBindings[param.Bind]
	default:
		return singularBindings[param.Bind]
	}
}

// isStringBind reports whether the request takes the argument as a quoted JSON
// string.
func isStringBind(param *ir.Param) bool {
	switch param.Bind {
	case ir.BindString, ir.BindBytes, ir.BindTimestamp, ir.BindDuration, ir.BindFieldMask:
		return true
	}
	return false
}

func goDoc(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i, line := range lines {
		if line == "" {
			lines[i] = "//"
			continue
		}
		lines[i] = "// " + line
	}
	return strings.Join(lines, "\n")
}

func oneofGroups(cmd *ir.Command) [][]string {
	var order []string
	members := map[string][]string{}
	for _, param := range cmd.Params {
		if param.Oneof == "" {
			continue
		}
		if _, seen := members[param.Oneof]; !seen {
			order = append(order, param.Oneof)
		}
		members[param.Oneof] = append(members[param.Oneof], param.Name)
	}
	var groups [][]string
	for _, oneof := range order {
		if len(members[oneof]) >= 2 {
			groups = append(groups, members[oneof])
		}
	}
	return groups
}

func flagUsage(param *ir.Param) string {
	// pflag reads a back-quoted word as the value name.
	desc := strings.ReplaceAll(param.ShortHelp, "`", "")

	if len(param.EnumValues) == 0 {
		return desc
	}

	marker := "(values: " + strings.Join(param.EnumValues, " | ") + ")"
	if desc == "" {
		return marker
	}
	return desc + " " + marker
}

// Every package a built-in template references has a row, so goimports only
// prunes. A missing row makes it guess the path from the local machine.
var templateImports = map[string]string{
	"bytes":     "bytes",
	"context":   "context",
	"json":      "encoding/json",
	"errors":    "errors",
	"fmt":       "fmt",
	"io":        "io",
	"iter":      "iter",
	"maps":      "maps",
	"os":        "os",
	"slices":    "slices",
	"strings":   "strings",
	"time":      "time",
	"table":     "github.com/jedib0t/go-pretty/v6/table",
	"cobra":     "github.com/spf13/cobra",
	"jsonpath":  "github.com/theory/jsonpath",
	"sjson":     "github.com/tidwall/sjson",
	"term":      "golang.org/x/term",
	"grpc":      "google.golang.org/grpc",
	"status":    "google.golang.org/grpc/status",
	"connect":   "connectrpc.com/connect",
	"protojson": "google.golang.org/protobuf/encoding/protojson",
	"proto":     "google.golang.org/protobuf/proto",
	"yaml3":     "go.yaml.in/yaml/v3",
	"yaml":      "sigs.k8s.io/yaml",
}
