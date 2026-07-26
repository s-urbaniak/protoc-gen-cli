package target

import (
	"bytes"
	"fmt"
	"io/fs"
	"text/template"

	"github.com/braveokafor/proto-to-cli/internal/ir"
)

// RenderTemplate executes the named template from templateFS with model as
// its data.
func RenderTemplate(
	templateFS fs.FS,
	name string,
	funcs template.FuncMap,
	model *ir.Model,
) ([]byte, error) {
	t, err := template.New(name).Funcs(funcs).ParseFS(templateFS, name)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, model); err != nil {
		return nil, fmt.Errorf("render %s: %w", name, err)
	}

	return buf.Bytes(), nil
}
