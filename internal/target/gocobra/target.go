// Package gocobra is the Go + Cobra target for protoc-gen-cli. It emits one
// self-contained <file>_cli.pb.go for each proto file.
package gocobra

import (
	"embed"
	"fmt"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/braveokafor/proto-to-cli/internal/target"
	"golang.org/x/tools/imports"
)

//go:embed template.go.tmpl
var templateFS embed.FS

// Generate returns one <file>_cli.pb.go for a proto file with services.
func Generate(model *ir.Model, _ target.Options) ([]target.File, error) {
	if len(model.Services) == 0 {
		return nil, nil
	}

	rendered, err := target.RenderTemplate(
		templateFS,
		"template.go.tmpl",
		funcMap(model),
		model,
	)
	if err != nil {
		return nil, fmt.Errorf("gocobra: %w", err)
	}

	name := model.GeneratedFilenamePrefix + "_cli.pb.go"

	// imports.Process formats the code and drops the imports that the
	// services do not reach.
	pruned, err := imports.Process(name, rendered, &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	})
	if err != nil {
		return nil, fmt.Errorf("gocobra: goimports: %w\n--- generated ---\n%s", err, rendered)
	}

	return []target.File{{Name: name, Content: string(pruned)}}, nil
}
