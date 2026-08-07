# Annotations

The `cli.v1` options give a proto's owner control of the generated CLI. Without them, everything derives from the proto.

Add the schema to your module:

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
deps:
  - buf.build/braveokafor/protoc-gen-cli
```

Run `buf dep update` to resolve it. Without that step, generation fails on the unresolved import.

Then import the schema in each proto that uses an option:

```protobuf
import "cli/v1/cli.proto";
```

[`examples/bookstore-annotated`](../examples/bookstore-annotated) configures a whole CLI this way. The caller passes no options.

An annotation costs one Go dependency. `protoc-gen-go` writes `_ "github.com/braveokafor/protoc-gen-cli/proto/cli/v1"` into its own output to register the option types. `go mod tidy` then adds `require github.com/braveokafor/protoc-gen-cli` to your `go.mod`.

A proto with no annotations puts nothing from this repository in your build.

## What they do

This section shows one request, before and after the annotations. The proto gains `hoist` on the message field, `skip` on the server-assigned id, two shorthands, and a `help` override:

```protobuf
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

The flags change from this:

```console
      --book string          The book to add.
      --book.author string   The author of the book.
      --book.id int          A unique book id.
      --book.title string    The book title.
  -h, --help                 help for add
      --shelf int            The shelf that receives the book.
```

to this:

```console
  -a, --author string   The author of the book.
      --book string     The book to add.
  -h, --help            help for add
      --shelf int       The shelf that receives the book.
  -t, --title string    Title to print on the spine.
```

`hoist` dropped the `--book.` prefix and kept `--book` itself. `skip` removed the id. The shorthands came from the field options. `help` replaced the title's comment.

## Command names

`cli.v1.service` controls the command that a service generates. `cli.v1.command` controls the subcommand that an rpc generates. Both have the same three fields.

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `name` | `string` | derived | Replaces the derived name. |
| `aliases` | `repeated string` | none | Other names for the command. |
| `help` | `string` | the comment | Replaces the command's derived help text. |

Names derive from the proto:

- A command name is the kebab-case service name, minus a `-service` suffix. `BookstoreService` derives `bookstore`.
- A subcommand name is the kebab-case rpc name. `CreateBook` derives `create-book`.

Rename a service when its derived name repeats the binary name. A root command named `bookstore` and a service named `BookstoreService` produce `bookstore bookstore get-book`.

```protobuf
service BookstoreService {
  option (cli.v1.service).name = "catalog";
  option (cli.v1.service).aliases = "books";

  rpc CreateBook(CreateBookRequest) returns (Book) {
    option (cli.v1.command).name = "add";
  }
}
```

## Params

`cli.v1.param` controls the param that a request field generates.

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `name` | `string` | derived | Replaces the field's derived segment in param names. |
| `shorthand` | `string` | none | The one-letter form of the param. |
| `skip` | `bool` | `false` | Derives no params from the field or its subtree. |
| `expand_depth` | `optional int32` | the plugin option | The sub-param budget below the field. |
| `hoist` | `bool` | `false` | Moves the field's sub-params into the parent namespace. |
| `help` | `string` | the comment | Replaces the field's derived help text. |

```protobuf
message Book {
  int64 id = 1 [(cli.v1.param).skip = true];
  string author = 2 [(cli.v1.param).shorthand = "a"];
  string title = 3 [(cli.v1.param).shorthand = "t", (cli.v1.param).help = "Title to print on the spine."];
}
```

### help

The plugin reads a proto comment as the help text. `help` replaces it. The request message's comment also joins the command's long help.

A comment is API reference prose. It is often wrong for a one-line flag. Use `help` when the comment is too long for the slot. Also use `help` when the comment contains detail an end user does not need:

```protobuf
// The title as it appears on the copyright page, transliterated to Latin
// script where the original is not, per the cataloguing rules in RFC 9999.
string title = 3 [(cli.v1.param).help = "Title to print on the spine."];
```

```console
  -t, --title string   Title to print on the spine.
```

`cli.v1.service` and `cli.v1.command` have the same field. On a command it replaces the rpc's comment. The request message's comment and the streaming notes still append.

### skip

`skip` removes the field and its whole subtree from the flags. Whole-request input continues to set the field. `-f` and `-d` still reach it. Use it for a server-assigned field such as an id.

### shorthand

`shorthand` is one ASCII letter. Do not use `f`, `d`, `h`, or `o`. The generated commands claim those. Two params of one command cannot share a letter.

Each of these is a generation error, not a warning.

### expand_depth

A message field always gets its own flag, which takes a JSON document. It also gets a dotted flag for each of its sub-fields, down to a budget.

`request-expand-depth` sets that budget for the whole run. `expand_depth` overrides it below one field.

`expand_depth = 0` keeps only the field's own flag. The field then takes a JSON document and nothing else:

```protobuf
message CreateShelfRequest {
  // The shelf to create.
  Shelf shelf = 1 [(cli.v1.param).expand_depth = 0];
}
```

```console
      --shelf string   The shelf to create.
```

The plugin drops `expand_depth` with a warning on a field that does not expand. A scalar, a repeated field, and a map field never expand.

### hoist

`hoist` moves a field's sub-flags up into the parent namespace. The field's own flag remains.

Without it, a request that wraps its payload in one field produces a prefix on every flag:

```
--lot  --lot.book  --lot.book.author  --lot.book.title  --lot.condition
```

With `(cli.v1.param).hoist = true` on `lot`:

```
--lot  --book  --book.author  --book.title  --condition
```

Use it when the wrapper field has no meaning for the user.

`hoist` needs the field to have sub-flags. The plugin drops it with a warning otherwise. That covers a scalar, a repeated field, a map field, and a field held at `expand_depth = 0`.

The warning goes to stderr and does not stop generation. Read the output of `buf generate` after you add the option.

## Deprecation

There is no `cli.v1` option for this. The plugin reads protobuf's own `deprecated`, which every service, rpc and field has:

```protobuf
service InventoryService {
  option deprecated = true;
}

message WatchAuctionRequest {
  int32 limit = 3 [deprecated = true];
}
```

Help hides a deprecated command or flag. It still works. It says so when you use it:

```console
$ bookstore catalog delete-shelf --shelf 1
Command "delete-shelf" is deprecated, the API can remove it at any time.

$ bookstore auctions watch --auction 1 --limit 1
Flag --limit has been deprecated, the API can remove it at any time.
```

Use it for something that old scripts still call. A new user does not need to learn it. Use `skip` when no one ever needs to set the field.

`deprecated` also reaches `protoc-gen-go`, which marks the generated Go types deprecated.

## Views

`cli.v1.view` declares the table columns of a message. It applies wherever the message displays, both as a whole response and as a row of a sub-table.

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `fields` | `repeated ViewField` | derived | Replace the derived view fields in all views of the message. |
| `expand_depth` | `optional int32` | the plugin option | The sub-field budget of the message's derived view. |

`ViewField` holds two fields.

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `label` | `string` | the path | The display text for the column. |
| `path` | `string` | none | Proto field names joined with dots, from the viewed message. |

```protobuf
message Auction {
  option (cli.v1.view) = {
    fields: [
      {label: "ID", path: "id"},
      {label: "TITLE", path: "lot.book.title"},
      {label: "ENDS", path: "ends_at"}
    ]
  };
}
```

Each segment but the last must be a singular message field that expands. The last segment is the value the column shows. The plugin drops a path that does not resolve. It writes a warning.

### expand_depth

A message field gets a column for each of its own fields, such as `lot.book.title`, down to a budget.

`response-expand-depth` sets that budget for the whole run. `expand_depth` overrides it for one message.

`expand_depth = 0` gives the message field one column instead. The cell then holds the whole message:

```protobuf
message FlatRequest {
  option (cli.v1.view).expand_depth = 0;

  Widget widget = 1;
  string note = 2;
}
```

```console
$ kitchen-sink ann flat --widget.leaf top --widget.gadget.leaf mid --note hello -o table
+--------------------------+-------+
| WIDGET                   | NOTE  |
+--------------------------+-------+
| gadget=leaf=mid,leaf=top | hello |
+--------------------------+-------+
```

`fields` wins over `expand_depth`. A message that declares both gets its declared columns, and the budget does nothing.

The caller can override a view with no change to the proto. See [The gocobra target](gocobra.md).

The two forms spell paths differently. An annotation path uses proto field names. A field named `ends_at` is `ends_at`. A caller path is an RFC 9535 JSONPath over protojson output. The same field is `$.endsAt`.

## Warnings

The plugin writes these to stderr during generation. Each one names the proto element and the fix. None of them stops the run.

| Warning | Cause |
| --- | --- |
| `--<name> is reserved by a built-in flag; its flag is --arg-<name>` | A derived flag name collides with a built-in. |
| `hoist is dropped; the field does not expand into sub-flags` | `hoist` on a scalar, a repeated or map field, or a field held at `expand_depth = 0`. |
| `expand_depth is dropped; the field does not expand into sub-flags` | `expand_depth` on the same kinds of field. |
| `view path "<path>" does not resolve; the declared field is dropped` | A `ViewField.path` that names no field. |
| `group-encoded fields get no flag; set it with -f or -d, or make the field a message` | A proto2 group. Set it with `-f` or `-d`. |
| `no example request; --example prints {}, so build it with -f or -d` | Every field of the request is server-defined (for example, `Any`). |

## Reserved names

The generated commands claim these long names: `filename`, `data`, `help`, `output`, `columns`, `example`, `timeout`, and `dry-run`. They claim the shorthands `f`, `d`, `h`, and `o`.

A derived name that collides gains an `arg-` prefix. The plugin writes a warning. A `name` annotation that collides is a generation error, because you chose it.
