package target

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"text/template"

	"github.com/braveokafor/protoc-gen-cli/internal/ir"
)

// The *.tmpl files in dir redefine same-named fragments.
func Render(
	builtin fs.FS,
	dir string,
	funcs template.FuncMap,
	model *ir.Model,
) ([]byte, error) {
	root, err := template.New("").Funcs(funcs).ParseFS(builtin, "*.tmpl")
	if err != nil {
		return nil, err
	}
	if dir != "" {
		if root, err = root.ParseGlob(filepath.Join(dir, "*.tmpl")); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := root.ExecuteTemplate(&buf, "file", model); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
