package target

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"text/template"

	"github.com/braveokafor/proto-to-cli/internal/ir"
)

// RenderTemplate executes a template with model as its data. A non-empty
// override names a file to parse instead.
func RenderTemplate(
	templateFS fs.FS,
	name, override string,
	funcs template.FuncMap,
	model *ir.Model,
) ([]byte, error) {
	t := template.New(name).Funcs(funcs)

	source := name
	var err error
	if override != "" {
		source = override
		text, readErr := os.ReadFile(override)
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", override, readErr)
		}
		t, err = t.Parse(string(text))
	} else {
		t, err = t.ParseFS(templateFS, name)
	}
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", source, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, model); err != nil {
		return nil, fmt.Errorf("render %s: %w", source, err)
	}

	return buf.Bytes(), nil
}
