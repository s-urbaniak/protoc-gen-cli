// Package target defines the contract that a code-generation back-end
// implements.
//
// A target renders the IR of one proto file into source for one language
// and CLI library. Implementations are in subpackages with the name
// <language><library>. gocobra is Go + Cobra. cmd/protoc-gen-cli selects
// one back-end by its opt=target= name.
package target

import "github.com/braveokafor/proto-to-cli/internal/ir"

// A Target is one code-generation back-end.
type Target func(model *ir.Model, opts Options) ([]File, error)

// Options contains the per-invocation settings that every target receives.
type Options struct {
	// TemplateOverride renders with this file instead of the target's
	// built-in template. Empty uses the built-in.
	TemplateOverride string
}

// A File is one file to emit.
type File struct {
	Name    string // Name is relative to the plugin's output directory.
	Content string
}
