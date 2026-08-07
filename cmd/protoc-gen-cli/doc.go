/*
Protoc-gen-cli generates an API's command-line interface from its .proto service definitions.

Each service becomes a command. Each rpc becomes a subcommand.
Each request field becomes its own flag. The comments in your proto become the help text.
The generated CLI also has whole-request input, shell completion, table output, streaming, and CLI-native exit codes.

# Install

	go install github.com/braveokafor/protoc-gen-cli/cmd/protoc-gen-cli@latest

# Use

The generated file holds only the command tree. It names the message types and the gRPC client stubs directly.
Run protoc-gen-go and protoc-gen-go-grpc into the same package:

	# buf.gen.yaml
	version: v2
	plugins:
	  - local: protoc-gen-go
	    out: gen
	    opt: [paths=source_relative]
	  - local: protoc-gen-go-grpc
	    out: gen
	    opt: [paths=source_relative]
	  - local: protoc-gen-cli
	    out: gen
	    opt: [paths=source_relative, target=gocobra]
	inputs:
	  - directory: proto

Under protoc, pass the same options with --cli_out and --cli_opt.

Mount the generated constructors on a root command of your own.
The plugin writes no main and no root command. The binary name, the connection, and the credentials stay yours.

# Options

Pass these in opt:.

  - target selects the target to generate with. Every run needs it.
    The gocobra target is Go with Cobra.
  - request-expand-depth sets how many message levels below a request field still get their own dotted flags, such as --book.title.
    The default is 1.
  - response-expand-depth sets how many message levels below a response field still get their own table columns, such as lot.book.title.
    The default is 1.
  - templates names a directory of *.tmpl files. Each {{define}} block in them replaces the built-in template fragment of the same name.
    Write them for the selected target.
  - dump-ir also writes <file>.cli.ir.json for each file. It shows the model that the target renders.
    The default is false.

# Documentation

The guide walks from a proto to a configured CLI:
https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/guide.md

The cli.v1 proto options configure the CLI from the schema:
https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/annotations.md
*/

package main
