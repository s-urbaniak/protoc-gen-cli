# The generated CLI

This page describes the CLI that the `gocobra` target generates. Your users read `--help`. Read this to learn what they get.

A service becomes a command. An rpc becomes a subcommand. A request field becomes a flag.

```console
$ bookstore catalog create-shelf --shelf.theme fiction
```

## Build a request

A **source** supplies request bodies. Field flags set fields on every request the command sends.

| Flag | Takes | Repeats |
| --- | --- | --- |
| `-f`, `--filename` | A file of request bodies, or `-` for stdin | Yes |
| `-d`, `--data` | A request body, inline | Yes |
| Field flags | One field each | See [Field types](field-types.md) |

The sources apply in the order `-f`, then `-d`, then the flags. A source holds one body or many. A JSON file holds one object, a stream of objects, JSON Lines, or an array. A YAML file holds one document for each `---`.

An rpc that sends one request merges every body into it with `proto.Merge`, the flags last. The merge is field-wise, so a later source overwrites only the fields it sets:

```console
$ cat shelf.json
{"shelf":{"id":1,"theme":"fiction"}}

$ bookstore catalog create-shelf -f shelf.json --shelf.theme poetry --dry-run
{
  "shelf": {
    "id": "1",
    "theme": "poetry"
  }
}
```

The id survives. Only the theme changes.

An rpc that sends many requests sends one for each body instead. See [Streaming](#streaming).

Both proto field names and JSON names work.

### Which format a source is

Whichever one reads it. There is no format flag. The first decoder to read a body takes the source. The extension and the name do not matter. One command can mix formats:

```console
$ bookstore catalog create-shelf -f shelf.json -f overrides.yaml -d '{"shelf":{"theme":"cli"}}'
```

The built-ins are JSON, then YAML. A file no decoder reads names what was tried:

```console
$ bookstore catalog create-shelf -f notes.md
Error: cannot read notes.md as any of: json, yaml
```

The caller can add formats and set their order. See [The gocobra target](gocobra.md).

Every flag but the request fields and `--help` sits under `Global Flags:`.

## `--example`

Flags reach the top of a request. A field deeper than `request-expand-depth` has no flag of its own. Its parent takes a JSON document instead. `--example` shows what that document holds.

It prints the whole request, filled in. It sends nothing:

```console
$ bookstore catalog create-shelf --example
{
  "shelf": {
    "id": "998",
    "theme": "Aut et repudiandae."
  }
}
```

Every string is lorem ipsum. Edit them.

The example is a request the server accepts. It round-trips:

```console
$ bookstore catalog create-shelf --example | bookstore catalog create-shelf -f -
```

`-o` selects the format, the same as for a response. See [Output](#output).

The example leaves out a `FieldMask` and an `Any`. Only the server knows those values. A request whose fields are all server-defined prints `{}`. The plugin warns about it during generation.


## Output

`-o`/`--output` selects the format.

| Format | Result |
| --- | --- |
| `json` | Indented JSON |
| `jsonl` | One compact line for each response |
| `yaml` | A YAML document for each response, opened by `---` |
| `table` | A table for each response |

With no `-o`, output is `json` on a terminal and `jsonl` in a pipe. The caller can add formats, remove them, and set a different default. See [The gocobra target](gocobra.md).

`-o` selects the format of a request body too, from `--example` or `--dry-run`. A request body has no derived columns, so `table` prints nothing for it.

A table fits the width of the terminal. Each repeated message field becomes its own titled sub-table, one row for each element. A streaming command prints one table for each response, as it arrives:

```console
$ bookstore catalog get-book --shelf 1 --book 2 -o table
+----+------------------+--------------+
| ID | AUTHOR           | TITLE        |
+----+------------------+--------------+
| 2  | Vladimir Nabokov | Buddenbrooks |
+----+------------------+--------------+
```

The columns are the response's fields. A field of a message field gets its own column too, such as `lot.id`, down to `response-expand-depth`. Past that budget the message field gets one column, and the cell holds the whole message. `--columns` names the paths for one run. It renders a table without `-o table`:

```console
$ bookstore catalog get-book --shelf 1 --book 2 --columns 'ID:$.id,TITLE:$.title'
+----+--------------+
| ID | TITLE        |
+----+--------------+
| 2  | Buddenbrooks |
+----+--------------+
```

One entry is a label, a colon, and an [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535) JSONPath. A comma separates the entries. Read the paths off `-o json`. The field names there are the path steps, under `$`. A path addresses the whole response. Reach into a list with `[*]`, as in `--columns 'IDS:$.auctions[*].id'`, which joins every match into one cell.

A proto owner can declare the columns in the proto. A caller can set them for one command. See [Annotations](annotations.md) and [The gocobra target](gocobra.md).

## Shell completion

Cobra gives your root command `completion bash`, `zsh`, `fish`, and `powershell`. The generated commands add what they know:

- An enum flag completes its value names.
- `-o` completes the formats that service really has, including the ones the caller added.
- Setting one member of a oneof hides the others.

No flag suggests a filename. A flag that takes free text completes nothing. It never offers the working directory.

## Streaming

Every command takes the same input. The rpc decides what the CLI does with it.

| Shape | Requests | Output |
| --- | --- | --- |
| Unary | Every body merges into one | One response |
| Server streaming | Every body merges into one | Each response prints as it arrives |
| Client streaming | One for each body | One response |
| Bidirectional | One for each body | Each response prints as it arrives |

A streaming command sends each body as it reads it. A large file needs no extra memory:

```console
$ bookstore inventory import-books -f catalogue.jsonl
$ producer | bookstore inventory import-books -f -
```

Field flags describe one body, so they send exactly one request:

```console
$ bookstore inventory import-books --book.title "Moby-Dick"
```

Nothing reads stdin unless `-f -` names it, for every shape. A command given neither a source nor a field flag sends one empty request, and a streaming command sends none.

A source that is a terminal writes a note to stderr. It reads bodies until Ctrl-D.

A client-streaming command answers once, at the end, so it shows nothing while it sends. On a terminal it writes a note to stderr saying so. In a pipe it writes nothing.

## Timeouts and cancellation

`--timeout` sets a deadline for each call. `30s` and `2m` both work. `0` means no deadline.

Ctrl-C cancels the call in flight. This works only when the caller passes a signal-aware context to `ExecuteContext`.

## `--dry-run`

`--dry-run` assembles the requests. It prints them. It sends nothing:

```console
$ bookstore catalog create-shelf --shelf.theme fiction --dry-run
{
  "shelf": {
    "theme": "fiction"
  }
}
```

A streaming command prints one body for each request. It reads a whole catalogue without touching the server.

`-o` selects the format, the same as for a response. See [Output](#output).

## Exit codes

The codes are terminal concepts. They are never gRPC status codes.

| Code | Meaning | Examples |
| --- | --- | --- |
| `0` | Success | |
| `2` | The invocation was wrong | An unknown flag, two flags of one oneof, an `-o` value with no printer |
| `124` | The call hit its deadline | `--timeout 2s` on a slower call |
| `130` | An interrupt | Ctrl-C |
| `1` | Everything else | The server returned an error, the connection failed, a flag value was not valid JSON |

A wrong invocation prints the usage. A failure after the call starts does not.

An unknown subcommand under a service prints the service help and exits `0`.

Error text has no gRPC framing. The CLI removes the `rpc error: code = ... desc = ...` wrapper. It prints the message the server sent:

```console
$ bookstore catalog get-book --shelf 1 --book 999
Error: book 999 not found on shelf 1
```

## Reserved flag names

The generated commands claim `--filename`, `--data`, `--help`, `--output`, `--columns`, `--example`, `--timeout`, and `--dry-run`. They also claim the shorthands `-f`, `-d`, `-h`, and `-o`.

A request field whose derived name collides gets an `arg-` prefix instead. A field named `timeout` produces `--arg-timeout`. The plugin warns during generation.
