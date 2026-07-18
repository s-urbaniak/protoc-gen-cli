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

// Generate returns one <GoPackageName>_cli.pb.go for the package. Models
// without services emit nothing.
func (t *Target) Generate(models []*ir.Model, _ target.Options) ([]target.File, error) {
	var model *ir.Model
	for _, m := range models {
		if len(m.Services) == 0 {
			continue
		}
		if model != nil {
			return nil, fmt.Errorf(
				"files %s and %s both declare services in package %s, which is not yet supported; keep a package's services in one file",
				model.ProtoFile,
				m.ProtoFile,
				m.ProtoPackage,
			)
		}
		model = m
	}
	if model == nil {
		return nil, nil
	}

	if model.FileOptions.GoPackageName == "" {
		return nil, fmt.Errorf(
			"%s has no go_package option, which the generated Go code needs; add one",
			model.ProtoFile,
		)
	}

	rendered, err := target.RenderTemplate(templateFS, "template.go.tmpl", nil, model)
	if err != nil {
		return nil, fmt.Errorf("gocobra: %w", err)
	}

	formatted, err := format.Source(rendered)
	if err != nil {
		return nil, fmt.Errorf("gocobra: gofmt: %w\n--- generated ---\n%s", err, rendered)
	}

	name := filepath.Join(
		filepath.Dir(model.ProtoFile),
		model.FileOptions.GoPackageName+"_cli.pb.go",
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
