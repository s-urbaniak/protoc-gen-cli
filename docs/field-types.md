# Field types

Every flag value ends up in a protojson document. This page gives the text that each proto type accepts. The flag syntax comes from the `gocobra` target.

Run `<command> --example` to see a filled request for any command. That is the fastest answer to "what shape does this take".

## Scalars

| Proto type | Flag takes | Example |
| --- | --- | --- |
| `string` | Free text | `--title "Moby-Dick"` |
| `bool` | A bool flag. Use `=` to pass false. | `--enabled` or `--enabled=false` |
| `int32` `int64` `sint*` `fixed*` `sfixed*` | A whole number | `--shelf 1` |
| `uint32` `uint64` | A whole number, zero or above | `--count 7` |
| `double` `float` | A number, with or without a fraction | `--weight 1.5` |
| `bytes` | Base64 | `--bytes-field aGk=` |
| `enum` | A value name | `--choice CHOICE_FIRST` |

An enum flag lists its values in the help text and completes them on TAB. The plugin does not validate the name itself. An unknown name fails when protojson decodes the request.

## Well-known types

| Proto type | Flag takes | Example |
| --- | --- | --- |
| `google.protobuf.Timestamp` | RFC 3339 | `--timestamp 2026-01-01T00:00:00Z` |
| `google.protobuf.Duration` | Seconds with an `s` | `--duration 90s` |
| `google.protobuf.FieldMask` | Field paths joined with commas | `--field-mask a,b` |
| `google.protobuf.Any` | A JSON document with its own `@type` | `--any '{"@type":"type.googleapis.com/pkg.Msg","f":1}'` |
| `google.protobuf.Struct` `Value` `ListValue` | A JSON document | `--struct '{"a":1}'` |
| `google.protobuf.Empty` | A JSON object | `--empty '{}'` |
| `BoolValue` `StringValue` `Int64Value` and the other wrappers | The value the wrapper holds | `--bool-value=false` |

A wrapper tracks presence. A set wrapper survives as its zero value:

```console
$ kitchen-sink fields wrappers --bool-value=false --string-value "" --dry-run -o jsonl
{"boolValue":false,"stringValue":""}
```

## Messages

A message field always gets its own flag, which takes a JSON document:

```console
$ kitchen-sink fields messages --outer '{"stringLeaf":"fiction"}' --dry-run -o jsonl
{"outer":{"stringLeaf":"fiction"}}
```

It also gets a dotted flag for each sub-field, down to `request-expand-depth`:

```console
$ kitchen-sink fields messages --outer.string-leaf fiction --dry-run -o jsonl
{"outer":{"stringLeaf":"fiction"}}
```

Both reach the same field. Two flags that set one leaf are an error.

Past the depth limit, only the JSON document flag remains. A recursive message stops at the cycle, whatever the depth limit. Use the JSON flag or `-f` to reach anything deeper.

## Repeated fields

Repeat the flag. Each use adds one element.

```console
$ kitchen-sink fields repeated --strings a --strings b --dry-run -o jsonl
{"strings":["a","b"]}
```

A repeated message takes one JSON document for each element:

```console
$ kitchen-sink fields repeated --outers '{"stringLeaf":"a"}' --outers '{"stringLeaf":"b"}' --dry-run -o jsonl
{"outers":[{"stringLeaf":"a"},{"stringLeaf":"b"}]}
```

**A repeated string keeps commas. Repeated numbers and bools split on them.**

```console
$ kitchen-sink fields repeated --strings "a,b" --dry-run -o jsonl
{"strings":["a,b"]}

$ kitchen-sink fields repeated --ints 1,2 --ints 3 --dry-run -o jsonl
{"ints":["1","2","3"]}
```

A repeated field never expands into dotted flags.

## Maps

Each use adds one entry as `key=value`. The split is on the first `=`.

```console
$ kitchen-sink fields maps --string-values k=v --string-values 'k2=a,b' --dry-run -o jsonl
{"stringValues":{"k":"v","k2":"a,b"}}
```

A map value never splits on commas. A repeat of one key takes the last value.

A key is always text, even when the proto key type is an integer or a bool. protojson writes every map key as a JSON string:

```console
$ kitchen-sink fields maps --int64-keys 1=a --dry-run -o jsonl
{"int64Keys":{"1":"a"}}
```

A map whose values are messages, numbers, or bools takes JSON for the value. A map field never expands into dotted flags.

## Oneofs

Each member of a oneof gets its own flag. The CLI rejects more than one:

```console
$ kitchen-sink fields oneofs --text hi --count 3
Error: if any flags in the group [text count pick] are set none of the others can be; [count text] were all set
```

A proto3 `optional` field is not a oneof for this purpose. It becomes an ordinary flag. Its zero value survives because the field tracks presence.

## Fields with no flag

Two kinds of field never get a flag. Set them with `-f` or `-d`.

- A proto2 group. The plugin writes a warning during generation.
- A field below `skip`. See [Annotations](annotations.md).

`--example` leaves out a `FieldMask` and an `Any`. Only the server knows those values.
