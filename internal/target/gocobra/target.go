// Package gocobra is the Go + Cobra target for protoc-gen-cli. It emits one
// self-contained <GoPackageName>_cli.pb.go per proto package.
package gocobra

import (
	"embed"
	"fmt"
	"go/format"
	"path/filepath"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/braveokafor/proto-to-cli/internal/target"
	"golang.org/x/tools/imports"
)

//go:embed template.go.tmpl
var templateFS embed.FS

// Target implements target.Target for Go + Cobra.
type Target struct{}

// Generate returns one <GoPackageName>_cli.pb.go for the package.
func (t *Target) Generate(models []*ir.Model, _ target.Options) ([]target.File, error) {
	merged, sources, err := mergePackage(models)
	if err != nil {
		return nil, err
	}
	if merged == nil {
		return nil, nil
	}

	if err := checkFlagIdentifiers(merged); err != nil {
		return nil, err
	}

	rendered, err := target.RenderTemplate(
		templateFS,
		"template.go.tmpl",
		funcMap(merged, sources),
		merged,
	)
	if err != nil {
		return nil, fmt.Errorf("gocobra: %w", err)
	}

	formatted, err := format.Source(rendered)
	if err != nil {
		return nil, fmt.Errorf("gocobra: gofmt: %w\n--- generated ---\n%s", err, rendered)
	}

	name := filepath.Join(
		filepath.Dir(merged.ProtoFile),
		merged.FileOptions.GoPackageName+"_cli.pb.go",
	)

	// Best-effort import cleanup.
	pruned, err := imports.Process(name, formatted, &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	})
	if err == nil {
		formatted = pruned
	}

	return []target.File{{Name: name, Content: string(formatted)}}, nil
}

// mergePackage folds the package's service-bearing models into one, or nil if there are none.
func mergePackage(models []*ir.Model) (*ir.Model, []string, error) {
	var withServices []*ir.Model
	for _, m := range models {
		if len(m.Services) > 0 {
			withServices = append(withServices, m)
		}
	}
	if len(withServices) == 0 {
		return nil, nil, nil
	}

	base := withServices[0]
	if base.FileOptions.GoPackageName == "" {
		return nil, nil, fmt.Errorf(
			"%s has no go_package option, which the generated Go code needs; add one",
			base.ProtoFile,
		)
	}

	merged := &ir.Model{
		PluginVersion: base.PluginVersion,
		ProtoFile:     base.ProtoFile,
		ProtoPackage:  base.ProtoPackage,
		FileOptions:   base.FileOptions,
	}
	sources := make([]string, 0, len(withServices))
	for _, m := range withServices {
		if m.FileOptions.GoPackageName != base.FileOptions.GoPackageName {
			return nil, nil, fmt.Errorf(
				"%s and %s share proto package %s but set different go_package options; give them one go_package",
				base.ProtoFile,
				m.ProtoFile,
				base.ProtoPackage,
			)
		}
		merged.Services = append(merged.Services, m.Services...)
		sources = append(sources, m.ProtoFile)
	}
	return merged, sources, nil
}

func checkFlagIdentifiers(model *ir.Model) error {
	for _, svc := range model.Services {
		for _, cmd := range svc.Commands {
			seen := map[string]string{}
			for _, f := range cmd.Flags {
				id := goVarName(f)
				if prior, ok := seen[id]; ok {
					return fmt.Errorf(
						"gocobra: rpc %s.%s: fields %q and %q both derive the generated identifier %s; rename one of the fields",
						svc.ProtoName,
						cmd.ProtoName,
						prior,
						f.ProtoPath,
						id,
					)
				}
				seen[id] = f.ProtoPath
			}
		}
	}
	return nil
}
