package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/braveokafor/proto-to-cli/internal/ir"
	"github.com/braveokafor/proto-to-cli/internal/irbuild"
	"github.com/braveokafor/proto-to-cli/internal/target"
	"github.com/braveokafor/proto-to-cli/internal/target/gocobra"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// version is set via -ldflags at build time; "dev" otherwise.
var version = "dev"

// Config holds one invocation's inputs: options parsed from opt= parameters,
// plus what main wires in.
type Config struct {
	// Parsed from opt=.
	Target              string // key into Targets
	DumpIR              bool   // also emit each file's IR beside it as <file>.cli.ir.json
	RequestExpandDepth  int    // 0 = only the request's own fields get flags
	ResponseExpandDepth int    // 0 = only the response's own fields become view fields

	// Wired in main.
	Version string                   // stamped into generated-file headers
	Targets map[string]target.Target // the selectable back-ends, e.g. "gocobra"
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-version" {
		name := filepath.Base(os.Args[0])
		if _, err := fmt.Fprintf(os.Stdout, "%v %v\n", name, version); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	var cfg Config
	cfg.Version = version
	cfg.Targets = map[string]target.Target{
		"gocobra": &gocobra.Target{},
	}

	var flags flag.FlagSet
	flags.BoolVar(&cfg.DumpIR, "dump-ir", false, "dump each file's IR as JSON")
	flags.StringVar(&cfg.Target, "target", "", "target to generate with")
	flags.IntVar(&cfg.RequestExpandDepth, "request-expand-depth", 1,
		"how many message levels down fields still get dotted flags")
	flags.IntVar(&cfg.ResponseExpandDepth, "response-expand-depth", 1,
		"how many message levels down response fields still become view fields")

	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plug *protogen.Plugin) error {
		return run(plug, &cfg)
	})
}

// run builds each generating file's IR and hands the models, grouped by
// proto package, to the selected target.
func run(plug *protogen.Plugin, cfg *Config) error {
	targetNames := slices.Sorted(maps.Keys(cfg.Targets))
	if cfg.Target == "" {
		return fmt.Errorf("opt=target=<name> is required (available: %v)", targetNames)
	}
	tgt := cfg.Targets[cfg.Target]
	if tgt == nil {
		return fmt.Errorf("unknown target %q (available: %v)", cfg.Target, targetNames)
	}
	if cfg.RequestExpandDepth < 0 {
		return fmt.Errorf(
			"opt=request-expand-depth=%d is negative; use 0 or more (0 expands no message fields)",
			cfg.RequestExpandDepth,
		)
	}
	if cfg.ResponseExpandDepth < 0 {
		return fmt.Errorf(
			"opt=response-expand-depth=%d is negative; use 0 or more (0 expands no message fields)",
			cfg.ResponseExpandDepth,
		)
	}

	plug.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
	plug.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024
	plug.SupportedFeatures = uint64(
		pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL,
	) | uint64(
		pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS,
	)

	// Models group per output directory and proto package, so a target
	// emits package-scoped files once.
	var keys []string
	groups := map[string][]*ir.Model{}

	for _, file := range plug.Files {
		if !file.Generate {
			continue
		}

		protoPath := file.Desc.Path()

		m, err := irbuild.Build(file, irbuild.Options{
			PluginVersion: cfg.Version,
			Files:         plug.FilesByPath,
			Warn: func(msg string) {
				fmt.Fprintf(os.Stderr, "protoc-gen-cli: %s: %s\n", protoPath, msg)
			},
			RequestExpandDepth:  cfg.RequestExpandDepth,
			ResponseExpandDepth: cfg.ResponseExpandDepth,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", protoPath, err)
		}

		if cfg.DumpIR {
			dump, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal IR for %s: %w", protoPath, err)
			}
			dump = append(dump, '\n')

			g := plug.NewGeneratedFile(
				file.GeneratedFilenamePrefix+".cli.ir.json",
				file.GoImportPath,
			)
			if _, err := g.Write(dump); err != nil {
				return fmt.Errorf("write IR for %s: %w", protoPath, err)
			}
		}

		key := path.Dir(protoPath) + ":" + m.ProtoPackage
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], m)
	}

	written := map[string]bool{}
	for _, key := range keys {
		files, err := tgt.Generate(groups[key], target.Options{})
		if err != nil {
			return err
		}

		for _, f := range files {
			if written[f.Name] {
				return fmt.Errorf("target %s: duplicate generated file %s", cfg.Target, f.Name)
			}
			written[f.Name] = true

			plug.NewGeneratedFile(f.Name, "").P(strings.TrimRight(f.Content, "\n"))
		}
	}

	return nil
}
