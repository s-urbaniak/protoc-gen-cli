// Package gocobra is the Go + Cobra target. It emits one <file>_cli.pb.go for each proto file.
package gocobra

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/braveokafor/protoc-gen-cli/internal/ir"
	"github.com/braveokafor/protoc-gen-cli/internal/target"
	"golang.org/x/tools/imports"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Generate returns one <file>_cli.pb.go for a proto file with services.
func Generate(file *protogen.File, model *ir.Model, opts target.Options) ([]target.File, error) {
	if len(model.Services) == 0 {
		return nil, nil
	}

	builtin, err := fs.Sub(templateFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("gocobra: %w", err)
	}
	rendered, err := target.Render(builtin, opts.TemplatesDir, funcMap(file, model), model)
	if err != nil {
		return nil, fmt.Errorf("gocobra: %w", err)
	}

	name := model.GeneratedFilenamePrefix + "_cli.pb.go"

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
