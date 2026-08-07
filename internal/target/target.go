// Package target defines the code-generation back-end contract.
// Implementations live in subpackages named <language><library>.
package target

import (
	"github.com/braveokafor/protoc-gen-cli/internal/ir"
	"google.golang.org/protobuf/compiler/protogen"
)

// A Target is one code-generation back-end.
type Target func(file *protogen.File, model *ir.Model, opts Options) ([]File, error)

type Options struct {
	// TemplatesDir names a directory of *.tmpl files. Each {{define}} block in
	// them replaces the built-in template fragment of the same name. Empty
	// uses the built-ins.
	TemplatesDir string
}

type File struct {
	Name    string // Name is relative to the plugin's output directory.
	Content string
}
