# Guide

This guide goes from a `.proto` to a CLI your users can run. It uses the [`gocobra` target](gocobra.md). It builds a bookstore client with two services. It controls that client from the proto. It configures it from the caller.

## 1. Define the services

Start with plain proto. No plugin options yet.

```protobuf
// proto/bookstore/v1/bookstore.proto
syntax = "proto3";

package bookstore.v1;

option go_package = "example.com/guide/gen/bookstore/v1;bookstorev1";

// ShelfService manages shelves.
service ShelfService {
  // Create adds a shelf to the store.
  rpc Create(CreateShelfRequest) returns (Shelf) {}
  // List returns every shelf.
  rpc List(ListShelvesRequest) returns (ListShelvesResponse) {}
}

// BookService manages books on a shelf.
service BookService {
  // Add puts a book on a shelf.
  rpc Add(AddBookRequest) returns (Book) {}
}

message CreateShelfRequest {
  // The shelf to create.
  Shelf shelf = 1;
}

message ListShelvesRequest {}

message ListShelvesResponse {
  // Every shelf in the store.
  repeated Shelf shelves = 1;
}

message AddBookRequest {
  // The shelf that receives the book.
  int64 shelf = 1;
  // The book to add.
  Book book = 2;
}

message Shelf {
  // A unique shelf id.
  int64 id = 1;
  // The shelf theme, for example "fiction".
  string theme = 2;
}

message Book {
  // A unique book id.
  int64 id = 1;
  // The author of the book.
  string author = 2;
  // The book title.
  string title = 3;
}
```

Write real comments. The plugin turns each one into help text. These comments are your CLI's documentation.

## 2. Generate

The generated file holds the command tree. It names the message types and the gRPC client stubs directly. All three plugins write into one package:

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

Run `buf generate`. You get `bookstore.pb.go`, `bookstore_grpc.pb.go` and `bookstore_cli.pb.go` in `gen/bookstore/v1`.

## 3. Mount it

The plugin writes no `main` and no root command. The caller owns both:

```go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	bookstorev1 "example.com/guide/gen/bookstore/v1"
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
	root.AddCommand(
		bookstorev1.NewShelfServiceCommand(conn),
		bookstorev1.NewBookServiceCommand(conn),
	)

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

`go build` gives you a working CLI:

```console
$ bookstore --help
Available Commands:
  book        BookService manages books on a shelf.
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  shelf       ShelfService manages shelves.
```

Pick a root name that differs from your service names. A root named `shelf` here gives `shelf shelf create`.

## 4. Control the CLI from the proto

The derived CLI works. But `books add -t` reads better than `book add --book.title`. A server-assigned id does not belong on a flag at all. The `cli.v1` options fix that once, for every caller.

Add the schema to your module:

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
deps:
  - buf.build/braveokafor/protoc-gen-cli
```

Run `buf dep update` to resolve it. Then annotate:

```protobuf
import "cli/v1/cli.proto";

service BookService {
  option (cli.v1.service).name = "books";

  rpc Add(AddBookRequest) returns (Book) {
    option (cli.v1.command).name = "add";
  }
}

message AddBookRequest {
  // The shelf that receives the book.
  int64 shelf = 1;
  // The book to add.
  Book book = 2 [(cli.v1.param).hoist = true];
}

message Book {
  // A unique book id.
  int64 id = 1 [(cli.v1.param).skip = true];
  // The author of the book.
  string author = 2 [(cli.v1.param).shorthand = "a"];
  // The book title.
  string title = 3 [(cli.v1.param).shorthand = "t", (cli.v1.param).help = "Title to print on the spine."];
}
```

Regenerate. The command changes:

```console
$ bookstore books add --help
Flags:
  -a, --author string   The author of the book.
      --book string     The book to add.
  -h, --help            help for add
      --shelf int       The shelf that receives the book.
  -t, --title string    Title to print on the spine.
```

`hoist` dropped the `--book.` prefix. `skip` removed `--book.id`. The shorthands work. `help` replaced the title's comment in its usage line.

An annotation adds one Go dependency. `protoc-gen-go` writes a blank import of the schema package into its own output. `go mod tidy` then adds `require github.com/braveokafor/protoc-gen-cli` to your `go.mod`. A proto with no annotations has no such dependency.

[Annotations](annotations.md) documents every option.

## 5. Configure from the caller

Some choices belong to the binary rather than the schema. Pass an options struct for each service.

This sets YAML as the default output. It adds a `line` format. It declares the columns for `-o table`. A row of the list is a shelf, so one column set serves both commands:

```go
printLine := func(w io.Writer, _ bookstorev1.ShelfServiceView, body []byte) error {
	_, err := fmt.Fprintf(w, "line: %s\n", body)
	return err
}

shelfColumns := []bookstorev1.ShelfServiceViewField{
	{Label: "ID", Path: "$.id"},
	{Label: "THEME", Path: "$.theme"},
}

shelfOpts := bookstorev1.ShelfServiceOptions{
	DefaultPrinter: "yaml",
	Printers: map[string]bookstorev1.ShelfServicePrinter{
		"line": printLine,
	},
	Views: map[string]bookstorev1.ShelfServiceView{
		"bookstore.v1.ShelfService.Create": {Fields: shelfColumns},
		"bookstore.v1.ShelfService.List": {Lists: []bookstorev1.ShelfServiceViewList{
			{Label: "shelves", Path: "$.shelves[*]", Fields: shelfColumns},
		}},
	},
}
```

The new format joins the help text and the completion list:

```console
$ bookstore shelf create --help
Global Flags:
  -o, --output string          Output format: json, jsonl, line, table, yaml.
                               Default: yaml.
```

Both take effect:

```console
$ bookstore shelf create --shelf.theme fiction
---
id: "1"
theme: fiction

$ bookstore shelf create --shelf.theme fiction -o line
line: {"id":"1","theme":"fiction"}
```

Options are for each service. `BookService` keeps the defaults here. [The gocobra target](gocobra.md) covers input formats, table columns and exit codes as separate tasks.

## 6. The complete main.go

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	bookstorev1 "example.com/guide/gen/bookstore/v1"
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

	printLine := func(w io.Writer, _ bookstorev1.ShelfServiceView, body []byte) error {
		_, err := fmt.Fprintf(w, "line: %s\n", body)
		return err
	}

	shelfColumns := []bookstorev1.ShelfServiceViewField{
		{Label: "ID", Path: "$.id"},
		{Label: "THEME", Path: "$.theme"},
	}

	shelfOpts := bookstorev1.ShelfServiceOptions{
		DefaultPrinter: "yaml",
		Printers: map[string]bookstorev1.ShelfServicePrinter{
			"line": printLine,
		},
		Views: map[string]bookstorev1.ShelfServiceView{
			"bookstore.v1.ShelfService.Create": {Fields: shelfColumns},
			"bookstore.v1.ShelfService.List": {Lists: []bookstorev1.ShelfServiceViewList{
				{Label: "shelves", Path: "$.shelves[*]", Fields: shelfColumns},
			}},
		},
	}

	root := &cobra.Command{Use: "bookstore", Short: "Bookstore API client"}
	root.AddCommand(
		bookstorev1.NewShelfServiceCommand(conn, shelfOpts),
		bookstorev1.NewBookServiceCommand(conn),
	)

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

`signal.NotifyContext` with `ExecuteContext` is what makes Ctrl-C cancel the call in flight. Without it the process dies. Exit code `130` never happens.

`fmt` and `io` are here only for the custom printer. Drop both imports if you do not add one.

## Next

- [The generated CLI](cli-usage.md). What you just shipped, from your users' side.
- [Field types](field-types.md). The text each proto type accepts.
- [Annotations](annotations.md). Every `cli.v1` option.
- [The gocobra target](gocobra.md). Caller tasks, one at a time.
- [Troubleshooting](troubleshooting.md). Warnings and errors, with the fix.
