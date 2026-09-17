package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/braveokafor/protoc-gen-cli/internal/irbuild"
	"github.com/braveokafor/protoc-gen-cli/internal/target"
	"github.com/braveokafor/protoc-gen-cli/internal/target/gocobra"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// -ldflags sets this at build time.
var version = "dev"

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

var targets = map[string]target.Target{
	"gocobra": gocobra.Generate,
}

type Config struct {
	// These come from opt=.
	Target              string
	DumpIR              bool
	RequestExpandDepth  int    // At zero, only the request's own fields get params.
	ResponseExpandDepth int    // At zero, only the response's own fields become view fields.
	TemplatesDir        string // Empty uses the built-in templates.
	// Client selects the generated Go client API. It does not select a wire
	// protocol: connect-go can itself speak Connect, gRPC, or gRPC-Web.
	Client string

	// main sets this.
	Version string
}

func main() {
	ver := resolveVersion()
	if len(os.Args) == 2 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		fmt.Printf("%v %v\n", filepath.Base(os.Args[0]), ver)
		os.Exit(0)
	}

	var cfg Config
	cfg.Version = ver

	var flags flag.FlagSet
	flags.BoolVar(&cfg.DumpIR, "dump-ir", false, "dump each file's IR as JSON")
	flags.StringVar(&cfg.Target, "target", "", "target to generate with")
	flags.IntVar(&cfg.RequestExpandDepth, "request-expand-depth", 1,
		"how many message levels down fields still get dotted flags")
	flags.IntVar(&cfg.ResponseExpandDepth, "response-expand-depth", 1,
		"how many message levels down response fields still become view fields")
	flags.StringVar(&cfg.TemplatesDir, "templates", "",
		"replace built-in template with the *.tmpl files in this directory")
	flags.StringVar(&cfg.Client, "client", "grpc-go",
		"generated client API: grpc-go or connect-go")

	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plug *protogen.Plugin) error {
		return run(plug, &cfg)
	})
}

// run builds the IR of each file and gives it to the selected target.
func run(plug *protogen.Plugin, cfg *Config) error {
	targetNames := slices.Sorted(maps.Keys(targets))
	if cfg.Target == "" {
		return fmt.Errorf("opt=target=<name> is required (available: %v)", targetNames)
	}
	tgt, ok := targets[cfg.Target]
	if !ok {
		return fmt.Errorf("unknown target %q (available: %v)", cfg.Target, targetNames)
	}
	if !validClient(cfg.Client) {
		return fmt.Errorf("unknown client %q (available: [connect-go grpc-go])", cfg.Client)
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

	written := map[string]string{}
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
			dump, _ := json.MarshalIndent(m, "", "  ")
			dump = append(dump, '\n')

			g := plug.NewGeneratedFile(
				file.GeneratedFilenamePrefix+".cli.ir.json",
				file.GoImportPath,
			)
			if _, err := g.Write(dump); err != nil {
				return fmt.Errorf("write IR for %s: %w", protoPath, err)
			}
		}

		files, err := tgt(file, m, target.Options{
			TemplatesDir: cfg.TemplatesDir,
			Client:       cfg.Client,
			Files:        plug.FilesByPath,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", protoPath, err)
		}

		for _, f := range files {
			if prior, taken := written[f.Name]; taken {
				return fmt.Errorf(
					"%s and %s both generate %s; rename one of the files or give them distinct go_package paths",
					prior,
					protoPath,
					f.Name,
				)
			}
			written[f.Name] = protoPath

			plug.NewGeneratedFile(f.Name, "").P(strings.TrimRight(f.Content, "\n"))
		}
	}

	return nil
}

func validClient(client string) bool {
	return client == "grpc-go" || client == "connect-go"
}
