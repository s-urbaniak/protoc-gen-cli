package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/braveokafor/proto-to-cli/internal/irbuild"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// Plugin version.
// set via -ldflags at build time; "dev" otherwise.
var version = "dev"

// Config holds an invocation's inputs.
type Config struct {
	// Plugin build version.
	Version string
	// Dump proto IR as JSON.
	DumpIR bool
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

	var flags flag.FlagSet
	flags.BoolVar(&cfg.DumpIR, "dump-ir", false, "dump each file's IR as JSON")

	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plug *protogen.Plugin) error {
		return run(plug, &cfg)
	})
}

// Run executes a plugin invocation.
func run(plug *protogen.Plugin, cfg *Config) error {
	plug.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
	plug.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024
	plug.SupportedFeatures = uint64(
		pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL,
	) | uint64(
		pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS,
	)

	for _, file := range plug.Files {
		if !file.Generate {
			continue
		}

		protoPath := file.Desc.Path()

		m, err := irbuild.Build(file, irbuild.Options{
			PluginVersion: cfg.Version,
			Files:         plug.FilesByPath,
		})
		if err != nil {
			return fmt.Errorf("IR build for %s: %w", protoPath, err)
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
	}

	return nil
}
