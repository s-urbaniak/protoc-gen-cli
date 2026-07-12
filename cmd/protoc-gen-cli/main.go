package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// version is set via -ldflags at build time; "dev" otherwise.
var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-version" {
		name := filepath.Base(os.Args[0])
		if _, err := fmt.Fprintf(os.Stdout, "%v %v\n", name, version); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	var flags flag.FlagSet

	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}

	run := func(plug *protogen.Plugin) error {
		plug.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		return nil
	}

	opts.Run(run)
}
