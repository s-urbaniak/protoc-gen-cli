# protoc-gen-cli

Generate an API's command-line interface from its `.proto`.

[![Build](https://github.com/braveokafor/protoc-gen-cli/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/braveokafor/protoc-gen-cli/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/braveokafor/protoc-gen-cli)](https://github.com/braveokafor/protoc-gen-cli/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/braveokafor/protoc-gen-cli.svg)](https://pkg.go.dev/github.com/braveokafor/protoc-gen-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/braveokafor/protoc-gen-cli/blob/main/LICENSE)

`grpcurl` calls any gRPC server. You write each request body as JSON. `protoc-gen-cli` generates a CLI for one API instead:

- Each service becomes a command.
- Each RPC becomes a subcommand.
- Each request field becomes its own flag.
- Your proto comments become the help text.

The plugin runs under `buf` and `protoc`. It accepts proto2, proto3, and editions through 2024. The `target` option selects the language and the CLI library. `gocobra` generates Go with [Cobra](https://github.com/spf13/cobra).

## proto in, CLI out

This service and message:

```protobuf
// ShelfService manages shelves.
service ShelfService {
  // Create adds a shelf to the store.
  rpc Create(CreateShelfRequest) returns (Shelf) {}
}

message CreateShelfRequest {
  // The shelf to create.
  Shelf shelf = 1;
}

message Shelf {
  // A unique shelf id.
  int64 id = 1;
  // The shelf theme, for example "fiction".
  string theme = 2;
}
```

become this command:

```console
$ bookstore shelf create --help
Create adds a shelf to the store.

Usage:
  bookstore shelf create [flags]

Flags:
  -h, --help                 help for create
      --shelf string         The shelf to create.
      --shelf.id int         A unique shelf id.
      --shelf.theme string   The shelf theme, for example "fiction".

Global Flags:
      --columns string         Table columns, as LABEL:path pairs into the JSON response.
                               Example: --columns 'ID:$.id,NAME:$.name'.
  -d, --data stringArray       A request body, inline.
      --dry-run                Print the assembled requests without sending them.
      --example                Print an example request body without sending it.
  -f, --filename stringArray   Request bodies from a file, or '-' for stdin.
  -o, --output string          Output format: json, jsonl, table, yaml.
                               Default: json on a terminal, jsonl when piped.
      --timeout duration       Per-call deadline (e.g. 30s, 2m); 0 means no deadline.
```

The command and flag descriptions above come from comments in the proto. The plugin writes only the global flags.

## When to use it

Use `protoc-gen-cli` when you want an easy-to-maintain CLI for one API. It suits an API that changes often and your team calls every day.

Use `grpcurl` or `buf curl` instead when you want to poke an unfamiliar server once. They need no build step. They call any server through reflection.

The generated CLI never exposes its RPC transport to users. It takes no `-H` flag and no target address. The caller owns the gRPC connection or Connect client, so transport, credentials, and interceptors stay there.

## Install

```sh
brew install --cask braveokafor/tap/protoc-gen-cli
```

Or `go install github.com/braveokafor/protoc-gen-cli/cmd/protoc-gen-cli@latest`. The [releases page](https://github.com/braveokafor/protoc-gen-cli/releases) also has archives for Linux, macOS, and Windows.

The gRPC quickstart below runs all three plugins, so `protoc-gen-go` and `protoc-gen-go-grpc` go on `PATH` too:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Quickstart

The generated file holds only the command tree. It names the message types and the gRPC client stubs directly. Run all three plugins into the same package:

```yaml
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
```

Run `buf generate`. Under `protoc`, the same options go to `--cli_out` and `--cli_opt`:

```sh
protoc -I proto \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  --cli_out=gen --cli_opt=paths=source_relative,target=gocobra \
  proto/bookstore/v1/bookstore.proto
```

Then mount the generated constructors on a root command:

```go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	bookstorev1 "example.com/quickstart/gen/bookstore/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	root := &cobra.Command{Use: "bookstore"}
	root.AddCommand(bookstorev1.NewShelfServiceCommand(conn))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := root.ExecuteContext(ctx); err != nil {
		var coded interface{ ExitCode() int }
		if errors.As(err, &coded) {
			os.Exit(coded.ExitCode())
		}
		os.Exit(1)
	}
}
```

`go build` gives you the CLI above. The plugin generates no `main` and no root command.

A service command takes the name of its service, minus a `-service` suffix. Pick a root name that differs from it, because a root named `shelf` here produces `shelf shelf create`.

You can also rename the command from the proto. See [Annotations](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/annotations.md).

## What your users get

- **Flags.** A scalar field becomes its own flag. A field of a message field becomes a dotted flag such as `--shelf.theme`, down to `request-expand-depth`. A repeated field repeats. A map field takes `key=value`.
- **Server-owned fields.** A field annotated `google.api.field_behavior = OUTPUT_ONLY` never becomes a flag or an `--example` value, including below a request resource. It remains legal in `-d` and `-f` bodies because protobuf accepts complete request messages; servers decide whether to ignore or reject it.
- **Whole-request input.** `-f/--filename` reads a file or `-` for stdin. `-d/--data` takes a request body inline. Both repeat, and both apply before the flags. JSON and YAML are the built-in formats. The caller adds more. A source holds one body or many: JSON Lines, a JSON array, or a YAML file with `---` separators. An rpc that sends one request merges them with `proto.Merge`.
- **`--example`.** Flags reach the top of a request. A field deeper than `request-expand-depth` has no flag of its own. A nested request otherwise needs a hand-written body. `--example` prints the whole shape, filled in and ready to edit. The server accepts it as a request. It round-trips:

  ```console
  $ bookstore shelf create --example | bookstore shelf create -f -
  ```

- **Output formats.** `-o/--output` selects a format. The built-ins are `json`, `jsonl`, `yaml`, and `table`. The caller adds one, removes one, and sets a different default. With no `-o`, output is `json` on a terminal and `jsonl` in a pipe. A table column reads a response field, or a field below one down to `response-expand-depth`. A repeated message field becomes its own titled sub-table. `--columns 'ID:$.id,TITLE:$.title'` names the table columns for one run, and selects `table` on its own.
- **Shell completion.** Cobra gives your root `completion bash|zsh|fish|powershell`. Enum flags complete their values.
- **Streaming.** A server-streaming command prints each response as it arrives. A client-streaming command sends one request for each body, as it reads it, so `-f catalogue.jsonl` and `producer | bookstore … -f -` both stream. Nothing reads stdin unless `-f -` names it.
- **Exit codes.** `0` success. `2` a wrong invocation. `124` a deadline. `130` an interrupt. `1` everything else, including a failed call. The CLI removes the `rpc error:` wrapper from the message.
- **`--timeout` and `--dry-run`.** `--timeout` sets a deadline for each call. `--dry-run` prints the assembled requests and sends nothing.

## Options

Pass these in `opt:`.

| Option | Default | Meaning |
| --- | --- | --- |
| `target` | *(required)* | The target to generate with. `gocobra` is Go with Cobra. |
| `client` | `grpc-go` | The generated Go client API: `grpc-go` or `connect-go`. This does not choose a wire protocol: a Connect client can speak Connect, gRPC, or gRPC-Web. `connect-go` requires `connectrpc.com/connect` and `protoc-gen-connect-go` v1.20.0. |
| `request-expand-depth` | `1` | How many message levels below a request field still get their own dotted flags. |
| `response-expand-depth` | `1` | How many message levels below a response field still get their own table columns. |
| `templates` | *(built-in)* | A directory of `*.tmpl` files. Each `{{define}}` block replaces the built-in fragment of the same name. |
| `dump-ir` | `false` | Also write `<file>.cli.ir.json` for each file. This dump is a debugging aid. |

## Configure from the proto

Import the `cli.v1` schema and annotate the proto. The CLI then needs nothing from the caller. The proto renames a command, deprecates it, adds a shorthand, or declares table columns.

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
deps:
  - buf.build/braveokafor/protoc-gen-cli
```

Run `buf dep update` to resolve it, then annotate:

```protobuf
import "cli/v1/cli.proto";

service BookstoreService {
  option (cli.v1.service).name = "catalog";

  rpc CreateBook(CreateBookRequest) returns (Book) {
    option (cli.v1.command).name = "add";
  }
}

message Book {
  int64 id = 1 [(cli.v1.param).skip = true];
  string author = 2 [(cli.v1.param).shorthand = "a"];
  string title = 3 [(cli.v1.param).shorthand = "t", (cli.v1.param).help = "Title to print on the spine."];
}
```

Under `protoc`, the release archive has `proto/cli/v1/cli.proto`. Add the archive's `proto/` directory to your `-I` path.

[`examples/bookstore-annotated`](https://github.com/braveokafor/protoc-gen-cli/tree/main/examples/bookstore-annotated) configures a whole CLI this way. [Annotations](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/annotations.md) documents every field.

## Next steps

**[Read the guide.](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/guide.md)** It takes the quickstart above. It finishes with a configured two-service CLI and a complete `main.go`.

Then go straight to the task you have:

| I want to | Read |
| --- | --- |
| Rename or deprecate a command, add a shorthand, override the help text | [Annotations](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/annotations.md) |
| Set the default output, add an output or input format, choose table columns | [The gocobra target](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/gocobra.md) |
| Know what text a timestamp, duration, `Any` or nested message takes | [Field types](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/field-types.md) |
| See what my users get: flags, merge order, completion, exit codes | [The generated CLI](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/cli-usage.md) |
| Fix a warning, an error, or a missing flag | [Troubleshooting](https://github.com/braveokafor/protoc-gen-cli/blob/main/docs/troubleshooting.md) |

## Examples

Each example holds a proto, the generated code, a stub server, and a `main.go`. Each one runs.

- [`bookstore`](https://github.com/braveokafor/protoc-gen-cli/tree/main/examples/bookstore). A plain proto. The caller does the configuration.
- [`bookstore-annotated`](https://github.com/braveokafor/protoc-gen-cli/tree/main/examples/bookstore-annotated). The same API, configured from the proto.
- [`kitchen-sink`](https://github.com/braveokafor/protoc-gen-cli/tree/main/examples/kitchen-sink). Every field type, every RPC shape, and the edge cases.

## Versioning and compatibility

Releases follow [Semantic Versioning](https://semver.org).

Every generated header records the plugin version. `protoc-gen-cli -version` prints it. A local `go build` reads `dev`.

## Contributing

See [CONTRIBUTING.md](https://github.com/braveokafor/protoc-gen-cli/blob/main/CONTRIBUTING.md). Run `make all` before you open a pull request. Write [Conventional Commits](https://www.conventionalcommits.org).

## License

[MIT](https://github.com/braveokafor/protoc-gen-cli/blob/main/LICENSE).
