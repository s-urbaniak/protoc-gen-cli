# The gocobra target

`gocobra` generates Go with [Cobra](https://github.com/spf13/cobra). A proto file with services gives one `<file>_cli.pb.go`.

Each service exports one constructor:

```go
func NewShelfServiceCommand(
	conn grpc.ClientConnInterface,
	opts ...ShelfServiceOptions,
) *cobra.Command
```

The name is `New` plus the service's Go name plus `Command`. The options type is the same Go name plus `Options`. A service named `ItemService` gives `NewItemServiceCommand` and `ItemServiceOptions`.

Options are optional. Later ones overwrite earlier ones for each entry.

Each task below is complete enough to paste. For the whole picture, start with [the guide](guide.md).

## Set the default output format

With no `-o`, output is `json` on a terminal and `jsonl` in a pipe. `--columns` selects `table` on its own. `DefaultPrinter` replaces the default:

```go
root.AddCommand(bookstorev1.NewShelfServiceCommand(conn, bookstorev1.ShelfServiceOptions{
	DefaultPrinter: "table",
}))
```

The name must be one of the service's printers. An unknown name panics when the caller constructs the command.

Printers and decoders are separate. `table` is a printer, and never a decoder. `yaml` is both, under one name.

## Add an output format

`Printers` adds a format to `-o`. A print function renders **one** JSON body. `view` gives it the response's columns.

This `csv` format follows the same columns as `-o table`:

```go
import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/theory/jsonpath"
)

printCSV := func(w io.Writer, v bookstorev1.ShelfServiceView, body []byte) error {
	// --dry-run and --example print a request body, which has no view.
	if len(v.Fields) == 0 {
		return nil
	}
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	labels := make([]string, len(v.Fields))
	cells := make([]string, len(v.Fields))
	for i, f := range v.Fields {
		labels[i] = f.Label
		path, err := jsonpath.Parse(f.Path)
		if err != nil {
			return err
		}
		if found := path.Select(doc); len(found) > 0 {
			cells[i] = fmt.Sprint(found[0])
		}
	}
	if err := cw.Write(labels); err != nil {
		return err
	}
	if err := cw.Write(cells); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

root.AddCommand(bookstorev1.NewShelfServiceCommand(conn, bookstorev1.ShelfServiceOptions{
	Printers: map[string]bookstorev1.ShelfServicePrinter{
		"csv": printCSV,
	},
}))
```

`-o csv` now works. `csv` joins the completion list.

Every service exports its own spelling of these types. The services of one proto file share one underlying type, so one printer serves them all. Types differ across proto files: a caller that mounts services from several protos writes one printer for each file.

`v.Fields` holds the response's own columns. `v.Lists` holds one entry for each repeated message field. `-o table` uses it to render a list response as a sub-table. A printer that ignores both takes `_` for that parameter.

[`examples/bookstore`](../examples/bookstore/go-cobra/cli/main.go) has a `csv` printer that reads both. It prints rows for a single response and for a list.

### What a printer receives

A body is one **whole message**, as protojson. protojson decides the spelling. It uses JSON names. `ends_at` arrives as `endsAt`. A 64-bit integer arrives as a quoted string. An enum arrives as its value name.

protojson varies its spacing between runs. A printer that reads values, through a JSONPath or a parser, never notices. Normalize the bytes yourself if you print them verbatim. The built-in `json` format indents, and `jsonl` compacts.

How many times your printer runs follows the rpc:

| Call | Printer calls |
| --- | --- |
| Unary | One, with the whole response message |
| Server streaming or bidirectional | One for each response, as it arrives |

`--example` and `--dry-run` print request bodies rather than responses. Your printer gets each one with an empty view. `--dry-run` on a streaming command calls it once for each request. That is the guard at the top of `printCSV` above.

A unary rpc that returns a list gives you **one** body. That body holds the whole wrapper, such as `{"items":[…]}`. You do not get one call for each element. Walk the array yourself if you want a line for each element. The built-in `table` format splits it into a sub-table with the view's lists.

The built-in `yaml` format starts every document with `---`, so streamed bodies form one valid YAML stream. A printer that needs the same property does the same.

## Remove a built-in output format

A nil printer removes a format:

```go
Printers: map[string]bookstorev1.ShelfServicePrinter{
	"json": nil,
},
```

`-o json` then fails. The name disappears from the completion list and from the list of valid names in the error message.

## Add an input format

Input runs in stages, and a decoder owns exactly one of them:

```
source        -f PATH  (- is stdin) | -d TEXT     bytes
decoder       Decode(r)                           request bodies
flags         field flags                         one body's worth of fields
composition   what the rpc accepts                the request, or the requests
```

**A decoder answers one question. How many request bodies are in this source, and what is each one?** It never learns which rpc it feeds, whether the bodies merge or stream, or which flags the command has. Its two fields are in the [options reference](#options-reference).

A body is a JSON document, the same bytes a `Printer` receives. Yield one for each body you find, in order. A format with no separator yields once.

There is no format flag. The first decoder to read a body takes the source. One command can mix `-f shelf.json -f shelf.yaml -d '{…}'`.

A TOML decoder has this form:

```go
import (
	"encoding/json"
	"io"
	"iter"

	"github.com/BurntSushi/toml"
)

tomlDecoder := bookstorev1.ShelfServiceDecoder{
	Name: "toml",
	Decode: func(r io.Reader) iter.Seq2[[]byte, error] {
		return func(yield func([]byte, error) bool) {
			raw, err := io.ReadAll(r)
			if err != nil {
				yield(nil, err)
				return
			}
			var m map[string]any
			if err := toml.Unmarshal(raw, &m); err != nil {
				yield(nil, err)
				return
			}
			body, err := json.Marshal(m)
			if err != nil {
				yield(nil, err)
				return
			}
			yield(body, nil)
		}
	},
}

root.AddCommand(bookstorev1.NewShelfServiceCommand(conn, bookstorev1.ShelfServiceOptions{
	Decoders: append(bookstorev1.ShelfServiceDecoders, tomlDecoder),
}))
```

A TOML source now works, with no flag and no extension: `json` and `yaml` fail on it and TOML reads it.

### Decoder order

`Decoders` is a list, not a map, and a non-empty one **replaces** the built-ins. The order is the order the CLI tries them.

Each built-in has its own name, so your list holds them wherever you want:

```go
// json, then toml, then yaml
Decoders: []bookstorev1.ShelfServiceDecoder{
	bookstorev1.ShelfServiceDecoderJSON,
	tomlDecoder,
	bookstorev1.ShelfServiceDecoderYAML,
},

// json alone, which drops yaml
Decoders: []bookstorev1.ShelfServiceDecoder{bookstorev1.ShelfServiceDecoderJSON},
```

`ShelfServiceDecoders` is those two in their default order, for the common case of adding one at the end.

Order matters when two decoders can read the same bytes. `yaml` reads almost any text, so put a decoder for a plain-text format **after** the built-ins. A body has to be a JSON object. A decoder whose first body is a string or a number did not read the source. The next decoder gets it.

A decoder that fails hands the source to the next one. You never rewind it. A decoder that streams keeps streaming.

### One file, many bodies

Nothing in the pipeline assumes a source holds one body. A decoder is also how a caller teaches the CLI a bulk format its API needs. This one reads a catalogue, one title a line, and it streams:

```go
Decode: func(r io.Reader) iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		lines := bufio.NewScanner(r)
		for lines.Scan() {
			title := strings.TrimSpace(lines.Text())
			if title == "" {
				continue
			}
			body, err := json.Marshal(map[string]any{
				"book": map[string]any{"title": title},
			})
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(body, nil) {
				return
			}
		}
		if err := lines.Err(); err != nil {
			yield(nil, err)
		}
	}
}
```

It goes after the built-ins: `json` fails on a bare title. `yaml` reads one as a string rather than an object. Neither takes the source.

[`examples/bookstore`](../examples/bookstore/go-cobra/cli/main.go) registers exactly this one.

**The CLI sends the bodies a decoder yields.** Your API's own framing, such as a chunked upload, lives in the decoder.

`Decoders` is on every service. Every rpc shape reads input the same way.

## Choose table columns

A derived table shows the response's fields. A field of a message field gets a column too, down to `response-expand-depth`. A deep response gives a wide table, and `Views` replaces it with the columns you choose.

`Views` maps a method's full proto name to its view. That name is the proto package, a dot, the service name, a dot, and the method name. The entry replaces the derived view whole.

A view names its columns as `ViewField` values, a `Label` and a `Path`. The `Path` is an [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535) JSONPath into one row. A `Fields` view reads the response itself. A `Lists` view makes the elements of a repeated field the rows:

```go
shelfColumns := []bookstorev1.ShelfServiceViewField{
	{Label: "ID", Path: "$.id"},
	{Label: "THEME", Path: "$.theme"},
}

root.AddCommand(bookstorev1.NewShelfServiceCommand(conn, bookstorev1.ShelfServiceOptions{
	Views: map[string]bookstorev1.ShelfServiceView{
		"bookstore.v1.ShelfService.Create": {Fields: shelfColumns},
		"bookstore.v1.ShelfService.List": {Lists: []bookstorev1.ShelfServiceViewList{
			{Label: "shelves", Path: "$.shelves[*]", Fields: shelfColumns},
		}},
	},
}))
```

`Create` returns one shelf, so its columns read the response. `List` wraps the shelves in a repeated field, so its view lists that field and each shelf becomes a row. A list path selects the elements, which is why it ends in `[*]`.

Paths use JSON names. A field named `ends_at` is `endsAt`. A path that selects several nodes gives a comma-joined cell, so `$.shelves[*].theme` lists every theme in one column.

A method key sets the view of one command. Two commands that return one message can still differ. Many methods of a real API return `google.longrunning.Operation`. Each one needs its own columns.

The view type belongs to one proto file. A caller that mounts services from several protos needs one `Views` map for each proto file. It cannot share one map across files.

Two mistakes here are quiet. Look at them first when a table looks wrong. Neither one warns:

- A key that names no method has no effect at all. A typo is one such key. The table keeps its derived view.
- A path that resolves to nothing gives an empty column under its label.

An end user sets the columns of one run with `--columns 'LABEL:$.path,LABEL:$.path'`. A proto owner declares them for a message with `(cli.v1.view)`. See [Annotations](annotations.md).

## Mount several services

Each service is one `*cobra.Command`. A root holds as many as you need:

```go
root := &cobra.Command{Use: "bookstore"}
root.AddCommand(
	bookstorev1.NewShelfServiceCommand(conn, shelfOpts),
	bookstorev1.NewBookServiceCommand(conn),
)
```

Options are for each service. One service can take a custom format while another keeps the defaults.

## Rename a command

A service command takes the kebab-case service name, minus a `-service` suffix. A root named after the same thing gives a doubled name, such as `shelf shelf create`.

The command is an ordinary Cobra command. The caller can rename it:

```go
shelves := bookstorev1.NewShelfServiceCommand(conn)
shelves.Use = "shelves"
root.AddCommand(shelves)
```

A proto owner can do it once for every caller with `(cli.v1.service).name`. See [Annotations](annotations.md).

## Translate errors to exit codes

Errors have their own exit code. Read it with `errors.As` on an anonymous interface, the same way you read `os/exec.ExitError`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

if err := root.ExecuteContext(ctx); err != nil {
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) {
		os.Exit(coded.ExitCode())
	}
	os.Exit(1)
}
```

Pass a signal-aware context to `ExecuteContext`. Without it, Ctrl-C kills the process. It never cancels the call. Exit code `130` then never happens.

Cobra reports an unknown command, so that error has no code. The fallback decides it.

The codes are in [The generated CLI](cli-usage.md).

## Replace the templates

The generated file comes from a set of template fragments. Each fragment is a
`{{define}}` block named after the file it fills. A dotted name is a part of the
fragment before the dot: it renders inside that parent's body, and replacing the
parent replaces where its parts get called.

```
file      service    command                  options
                     command.input            input
                     command.flag             output
                     command.flag.value       output.viewFields
                     command.flag.json        call
                     command.flag.jsonArray   exit
                     command.flag.mapString
                     command.flag.mapJSON
```

Point the plugin at a directory to replace any of them:

```yaml
  - local: protoc-gen-cli
    out: gen
    opt: [paths=source_relative, target=gocobra, templates=tmpl]
```

Every `*.tmpl` file in the directory parses after the built-ins. Each
`{{define}}` in them replaces the fragment of that name. The file names mean
nothing. One file can hold one fragment or several.

Each fragment renders with the data its parent passes:

| Fragment | `.` holds |
| --- | --- |
| `file` | The whole model |
| `service` | One service |
| `command`, `command.input` | One command |
| `command.flag` | One param |
| `command.flag.value`, `.json`, `.jsonArray`, `.mapString`, `.mapJSON` | One param — the arm its `goBinding` overlay names |
| `output.viewFields` | One view's field list |
| `options`, `input`, `output`, `call`, `exit` | Nothing |

Generate with `dump-ir=true` and read `<file>.cli.ir.json` for the model the
templates render, field for field.

Start from a built-in: copy it from
[`internal/target/gocobra/templates/`](../internal/target/gocobra/templates/)
and edit it. An identifier's prefix comes from `{{ prefix }}` for the
unexported names and `{{ export }}` for the exported ones, so every generated
file has its own.

The other template functions, all usable from an override:

| Function | Gives |
| --- | --- |
| `goName` | The Go name of a service, method, or request message, from its full proto name |
| `goRequestType` | A command's request type, package-qualified when it is foreign |
| `goBinding` | How a param crosses pflag: the method stem, its zero default, and the overlay arm |
| `goDoc` | Text as a `//` comment block |
| `quoteJoin` | Names as a quoted, comma-separated list |
| `flagUsage` | A param's usage string, with its enum values |
| `oneofGroups` | A command's oneof groups, each a list of flag names |
| `imports` | The generated file's import map, alias to path |
| `messageViews` | The derived view of each response message |
| `fileHasShape` | Whether any command in the file has one of the named shapes |

A define whose name matches no fragment replaces nothing. Text outside
`{{define}}` blocks renders nothing. A directory with no `*.tmpl` files fails
the generate. Replacing `file` replaces the whole generated file.

## Options reference

```go
type ShelfServiceOptions struct {
	Printers       map[string]ShelfServicePrinter
	DefaultPrinter string
	Decoders       []ShelfServiceDecoder
	Views          map[string]ShelfServiceView
}

type ShelfServiceViewField struct {
	Label string
	Path  string // RFC 9535 JSONPath into the response.
}

type ShelfServiceViewList struct {
	Label  string
	Path   string
	Fields []ShelfServiceViewField
}

type ShelfServiceView struct {
	Lists  []ShelfServiceViewList
	Fields []ShelfServiceViewField
}

type ShelfServicePrinter func(w io.Writer, view ShelfServiceView, body []byte) error

type ShelfServiceDecoder struct {
	Name   string
	Decode func(r io.Reader) iter.Seq2[[]byte, error]
}

var (
	ShelfServiceDecoderJSON ShelfServiceDecoder
	ShelfServiceDecoderYAML ShelfServiceDecoder

	ShelfServiceDecoders []ShelfServiceDecoder
)
```

Each name is an alias of a type the proto file owns, exported under a `Cli_`
prefix built from the proto path. Services of one file alias the same underlying
types. Two files never share them.

| Field | Zero value | What a value does |
| --- | --- | --- |
| `Printers` | The built-in printers | Adds one under the name `-o` takes; a nil printer removes one |
| `DefaultPrinter` | `table` with `--columns`; else `json` on a terminal, `jsonl` in a pipe | Names the printer when `-o` is absent |
| `Decoders` | `json` then `yaml` | Replaces the list, in the order the CLI tries them |
| `Views` | The view derived from the response message | Replaces one method's view |

A method's full proto name keys `Views`, not a message name.

An unknown `DefaultPrinter` panics when the caller constructs the command.

## Generated dependencies

`<file>_cli.pb.go` imports `cobra`, `go-pretty/v6`, `theory/jsonpath`, `sjson`, `x/term`, `grpc`, `protobuf`, `sigs.k8s.io/yaml`, and `go.yaml.in/yaml/v3`. It imports nothing from this repository. The plugin version and your generated CLI stay independent.

An annotation changes that, through a different file. `protoc-gen-go` writes a blank import of the schema package into its own output. An annotated proto then puts `github.com/braveokafor/protoc-gen-cli/proto/cli/v1` in your build graph.

A proto with no annotations puts nothing from this repository there. See [Annotations](annotations.md).
