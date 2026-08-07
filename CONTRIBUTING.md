# Contributing

Thanks for your help with `protoc-gen-cli`.

## Development

`make help` lists every target. The common ones:

| Command | What it does |
| --- | --- |
| `make all` | Build, lint, and test. This is the gate every change must pass. |
| `make gen` | Regenerate the `cli.v1` schema (and fixtures). |
| `make verify-examples-regen` | Fail when regeneration changes a committed fixture. |
| `make lint` | `go vet`, golangci-lint, and `buf lint`. |
| `make release-snapshot` | Build a local release. It publishes nothing. |
| `make release-check` | Validate the GoReleaser config. |

Run `make all` before you open a pull request. When you change code generation, run `make gen` and commit the regenerated output. CI runs `make verify-examples-regen`, which fails on drift.

## Pre-commit hooks

The repository includes a `.pre-commit-config.yaml`. The hooks:

- fix trailing whitespace and file endings,
- make sure YAML is valid,
- block large files,
- run golangci-lint.

Install them once:

```sh
pip install pre-commit   # or: brew install pre-commit
pre-commit install
```

The hooks then run on each commit. Run them across every file with `pre-commit run --all-files`.

## Architecture

The plugin reads each proto file into an intermediate representation. Then a target renders that IR into source:

- `internal/ir` defines the IR.
- `internal/irbuild` builds it.
- `internal/target` defines a target; `gocobra` is the Go and Cobra one.

`internal/ir` and `internal/irbuild` know no target. The IR holds what applies to every target, whatever it generates. A `Param` is a request field a user supplies, not a flag. A target other than a CLI renders it another way. A name only one target wants, such as a Go identifier, belongs to that target. `cmd/protoc-gen-cli` is the one place that names the targets, in its `targets` map.

A target emits one self-contained file for each service file. `<file>_cli.pb.go` imports nothing from this repository. This is a maintenance decision rather than a feature. There is no runtime package. No release must stay compatible with a CLI an older plugin generated.

A target's templates live in its `templates/` directory, one file for each concept. `file`, `service`, and `command` mirror the IR. `options` is what the caller configures. `input`, `output`, `call`, and `exit` are the generated CLI's run lifecycle. A template writes an identifier's prefix with `{{ prefix }}` for the unexported names and `{{ export }}` for the exported ones. Every file then has its own.

A Go comment in a template renders into every generated file. Write one only for the CLI developer who reads that file. A `{{/* */}}` comment stays in the template. Machinery a generated file needs goes into these templates once, under a file-prefixed name. A new language target copies this shape and registers itself in `cmd/protoc-gen-cli/main.go`.

`proto/cli/v1` is public Go API, and the one package here that is. `protoc-gen-go` writes a blank import of it into its own output for any annotated proto. A consumer who annotates then has this module in their build graph. A breaking change to that package breaks them.

## Signatures

Parameters follow Go's own conventions rather than a local invention, so a Go developer
infers them without reading this file:

| rule | where it comes from |
| --- | --- |
| `ctx context.Context` first, and never inside an options struct | [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#contexts) |
| an options struct is the last argument | [Google Go Style Guide](https://google.github.io/styleguide/go/best-practices#option-structure) |
| destination before source | `io.Copy(dst, src)`, `proto.Merge(dst, src)`, `fmt.Fprintf(w, …)` |
| the callback a function invokes goes last | `sort.Slice(x, less)`, `filepath.Walk(root, fn)` |
| variadic last | the language |

Two more, for names. Never shadow a predeclared identifier: a printer is `printer`, not
`print`. Use one word for one concept across the whole repository. The proto option
`expand_depth`, the plugin option `RequestExpandDepth` and the local `depth` are one
idea, so they share a word.

A generated file follows the generator's conventions instead, because it is not
hand-written Go. Its identifiers use a `cli_` prefix built from the proto path, `Cli_` when
exported, matching the `file_..._proto_` names `protoc-gen-go` writes into the
same package.

## Commits and pull requests

A squash merge uses the pull request title as the commit message. Write the title as a Conventional Commit: `feat:`, `fix:`, `docs:`, `chore:`, and so on. GoReleaser groups the release notes by that prefix.

## Releasing

A `vX.Y.Z` tag triggers the release. It calls the CI workflow first. It then publishes:

- the binaries, with the Homebrew cask in `braveokafor/homebrew-tap`,
- the annotation module `buf.build/braveokafor/protoc-gen-cli`.

The workflow reads the `HOMEBREW_TAP_TOKEN` and `BUF_TOKEN` Actions secrets.
