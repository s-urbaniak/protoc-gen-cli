# Troubleshooting

Each message below is what the plugin or the generated CLI prints, and what to change. The CLI messages come from the `gocobra` target.

## The generated file does not compile

```
undefined: NewBookstoreServiceClient
undefined: CreateShelfRequest
```

`<file>_cli.pb.go` holds only the command tree. It names the message types and the gRPC client stubs directly. It needs the output of `protoc-gen-go` and `protoc-gen-go-grpc` in the same package.

Run all three plugins into one directory:

```yaml
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
```

## A command name repeats

A service command takes the kebab-case service name, minus a `-service` suffix. `BookstoreService` gives `bookstore`. A root of the same name doubles it.

Rename the command in the proto, which fixes it for every caller:

```protobuf
service BookstoreService {
  option (cli.v1.service).name = "catalog";
}
```

The caller can also set `cmd.Use` after construction. See [Annotations](annotations.md) and [The gocobra target](gocobra.md).

## Plugin options

These stop the run.

| Message | Fix |
| --- | --- |
| `opt=target=<name> is required (available: …)` | Add `target=gocobra` to `opt:`. |
| `unknown target "x" (available: …)` | The `target` value names no target. Use one from the list. |
| `opt=request-expand-depth=N is negative; use 0 or more (0 expands no message fields)` | Use `0` or more. `0` gives no dotted flags. |
| `opt=response-expand-depth=N is negative; use 0 or more (0 expands no message fields)` | Use `0` or more. `0` gives a message field one column. |
| ``gocobra: template: pattern matches no files: `tmpl/*.tmpl` `` | The `templates` directory holds no `*.tmpl` files. Point it at the right directory. |
| `gocobra: template: …` | A `*.tmpl` file in the `templates` directory does not parse. Fix the `{{define}}` block it names. |

## Generation errors

These stop the run.

| Message | Fix |
| --- | --- |
| `service X has no rpcs to generate commands from; add an rpc or exclude the file from generation` | An annotated file holds an empty service. Add an rpc, or exclude the file. |
| `service X: command name "n" must be nonempty, without spaces or a leading '-'` | A `(cli.v1.*)` name or alias is empty or malformed. The same check names `alias`, `subcommand name` and `flag name`. Give it a plain name. |
| `rpc X.Y: flag --x is reserved by a built-in flag; pick another (cli.v1.param).name` | You named a param after a built-in. Pick another `name`. The reserved set is below. |
| `rpc X.Y: flag --x: shorthand -f is reserved by a built-in flag; pick another letter` | Built-in flags use `f`, `d`, `h` and `o`. Pick another letter. |
| `rpc X.Y: flag --x: shorthand "ab" must be one ASCII letter` | A shorthand is one letter. |
| `rpc X.Y: flags --x and --y both claim the shorthand -a; change one` | Two params of one command cannot share a letter. |
| `rpc X.Y: fields "a" and "b" both derive the flag --x; rename one of the fields` | Two fields derive one flag name. Rename a field, or set `(cli.v1.param).name` on one. |
| `rpc X.Y: fields "a.b_c" and "a_b.c" collapse to the same generated identifier; rename one of the fields` | The plugin drops `.` and `_` when it derives an identifier. These two then meet. Rename a field. |
| `rpc X.Y declares the alias "z" twice; drop one` | Remove the repeat. |
| `rpc X.Y's alias "z" duplicates its own subcommand name; drop the alias` | The alias adds nothing. Remove it. |
| `rpc X.Y's alias "z" collides with the subcommand of X.B; rename or realias one` | Two subcommands of one command claim the name. Change one. |
| `rpcs X.A and X.B both derive the subcommand "c"; rename one of the rpcs` | Two rpc names give one subcommand. Rename an rpc, or set `(cli.v1.command).name`. |
| `rpcs X.A and X.B both declare the alias "z"; realias one of the rpcs` | Two subcommands claim one alias. Change one. |
| `services X and Y both derive the command "c"; rename one of the services` | Two service names give one command. Rename a service, or set `(cli.v1.service).name`. |
| `services X and Y both declare the alias "z"; realias one of the services` | Two services claim one alias. Change one. |
| `service X declares the alias "z" twice; drop one` | Remove the repeat. |
| `service X's alias "z" duplicates its own command name; drop the alias` | The alias adds nothing. Remove it. |
| `service X's alias "z" collides with the command of service Y; rename or realias one` | An alias claims another service's command name. Change one. |
| `service X's alias "z" collides with an alias of rpc X.A; realias one` | A service alias and a subcommand alias share the name. Change one. |
| `service X's alias "z" collides with the subcommand of rpc X.A; typing "z" would run the command, not the subcommand; realias the service or rename the rpc` | A service alias hides an rpc's subcommand. Realias the service, or rename the rpc. |
| `message M: the view declares the label "l" twice; relabel one` | Two `(cli.v1.view)` fields share a label. Relabel one. |
| `a.proto and b.proto both generate v1/api_cli.pb.go; rename one of the files or give them distinct go_package paths` | Two proto filenames map to one output file. Rename one, or split their `go_package`. |

### Reserved names

The generated commands claim the long names `filename`, `data`, `help`, `output`, `columns`, `example`, `timeout` and `dry-run`. They also claim the shorthands `f`, `d`, `h` and `o`.

The plugin renames a **derived** name that collides. It adds an `arg-` prefix and prints a warning. A name you **chose** with `(cli.v1.param).name` is an error.

## Warnings

These do not stop the run. Each one names the proto element.

| Message | Cause and fix |
| --- | --- |
| `--timeout is reserved by a built-in flag; its flag is --arg-timeout` | A field derives a built-in name. Accept `--arg-timeout`, or set `(cli.v1.param).name`. |
| `hoist is dropped; the field does not expand into sub-flags` | `hoist` only moves sub-flags. A scalar, a repeated or map field, and a field held at `expand_depth = 0` have none. Remove it. |
| `expand_depth is dropped; the field does not expand into sub-flags` | Same cause. `expand_depth` only sets a budget for sub-flags. Remove it. |
| `view path "x" does not resolve; the declared field is dropped` | A `ViewField.path` names no field. Each segment is a proto field name. Each one but the last must be a singular message field that expands. |
| `group-encoded fields get no flag; set it with -f or -d, or make the field a message` | A proto2 group. Use `-f` or `-d`, or change the field to a message. |
| `no example request; --example prints {}, so build it with -f or -d` | Every field of the request is server-defined, such as a `FieldMask` or an `Any`. The plugin cannot invent those values. |

## The CLI panics at start

This is a mistake in the caller. It fails before a user runs anything.

```
panic: Options.DefaultPrinter: unknown printer "tabel"; use one of: json, jsonl, table, yaml
```

`DefaultPrinter` must name a printer the service has. Make sure the spelling is right. Remember that a nil `Printers` entry removes one.

A `Views` entry cannot panic. A path that names nothing gives an empty column instead.

## A table column is empty

A path that resolves to nothing gives an empty column under its label. The paths are [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535) JSONPaths into protojson output, so they start at `$`. Read the shape from `-o json` first.

Three cases surprise people:

- A `google.protobuf.Any` holds its type under `@type`, which is not a bare name step. Write it as `$["@type"]`.
- A wrapper holds no `value` key. A `google.protobuf.StringValue display_name` reads as `$.displayName`, not `$.displayName.value`.
- A `google.protobuf.FieldMask` reads as one string of comma-joined paths.

A `Views` key that names no method also gives no warning. Make sure the key matches the method's full proto name.

## `--columns` is rejected

```console
$ bookstore catalog get-book --shelf 1 --book 2 --columns 'ID'
Error: --columns entry "ID" has no colon; write LABEL:path

$ bookstore catalog get-book --shelf 1 --book 2 --columns 'ID:not a path'
Error: --columns entry "ID:not a path": jsonpath: unexpected identifier at position 1
```

An entry is a label, a colon, and a JSONPath. Both cases exit `2`.

## A table column holds a whole message

The field sits below `response-expand-depth`, or below a `(cli.v1.view).expand_depth` on its message. The cell then reads `key=value,key=value`. Raise the budget, or name the path you want with `--columns`.

## The CLI exits 1 where you expected 2

Exit `2` covers a wrong invocation. That means an unknown flag, two flags of one oneof, or an unknown `-o` value.

A flag value that is not valid JSON exits `1`. The CLI catches it after the command starts, not while it parses flags.

The full table is in [The generated CLI](cli-usage.md).

## A flag is missing

Work through these in order.

- The field has `(cli.v1.param).skip`. It derives no flag. Use `-f` or `-d`.
- The field is deeper than `request-expand-depth`. Its parent still takes a JSON document.
- The field is inside a repeated or map field. Those never expand into dotted flags.
- The message is recursive. Expansion stopped at the cycle.
- The field has no flag because its type has none. A proto2 group is the only such type.

## Two flags set one field

```console
$ kitchen-sink fields messages --outer '{"stringLeaf":"a"}' --outer.string-leaf b --dry-run
Error: flags: proto: (line 1:28): duplicate field "string_leaf"
```

The JSON document and the dotted flag both write the field. Write the document with proto field names, or drop one of the two flags.

## Annotations do not apply

Make sure the proto imports the schema:

```protobuf
import "cli/v1/cli.proto";
```

and that the module depends on it:

```yaml
# buf.yaml
deps:
  - buf.build/braveokafor/protoc-gen-cli
```

Run `buf dep update` after you add the dependency.
