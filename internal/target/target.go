// Package target defines the interface a code-generation back-end implements.
//
// A target renders the IR of one proto package into source for one language
// and CLI library. Implementations live in subpackages named
// <language><library>: gocobra is Go + Cobra, and cmd/protoc-gen-cli
// selects one by its opt=target= name.
package target

import "github.com/braveokafor/proto-to-cli/internal/ir"

// A Target generates source files from IR models.
type Target interface {
	// Generate turns one package's IR, one model per generating .proto
	// file, into the files to emit.
	Generate(models []*ir.Model, opts Options) ([]File, error)
}

// Options carries the per-invocation settings every target receives.
type Options struct{}

// A File is one file to emit.
type File struct {
	Name    string // path relative to the plugin's output directory
	Content string
}
