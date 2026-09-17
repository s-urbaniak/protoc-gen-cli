# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory
BIN := .tmp/bin
VERSION ?= dev

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: all
all: build lint test ## Build, lint and test

.PHONY: build
build: ## Build the plugin
	CGO_ENABLED=0 go build -trimpath -ldflags '-X main.version=$(VERSION)' -o $(BIN)/protoc-gen-cli ./cmd/protoc-gen-cli

.PHONY: test
test: ## Run tests
	go test -vet=off -race -cover ./...

.PHONY: lint
lint: lint-go lint-proto ## Lint Go and proto

.PHONY: lint-go
lint-go: ## Lint Go files
	go vet ./...
	golangci-lint run --modules-download-mode=readonly --timeout=3m0s

.PHONY: lint-proto
lint-proto: ## Lint proto files
	buf lint

.PHONY: lintfix
lintfix: ## Fix lint errors
	golangci-lint run --fix --modules-download-mode=readonly --timeout=3m0s

.PHONY: gen
gen: clean ## Regenerate proto code
	$(MAKE) gen-proto
	$(MAKE) gen-examples

.PHONY: gen-proto
gen-proto: fmt-proto ## Regenerate cli.v1 schema
	buf generate

.PHONY: gen-examples
gen-examples: fmt-proto ## Regenerate examples
	cd examples/bookstore && buf generate
	cd examples/bookstore-annotated && buf generate
	cd examples/kitchen-sink && buf generate
	cd examples/connect && buf generate

.PHONY: verify-examples-regen
verify-examples-regen: gen ## Fail if regenerating examples changes anything
	@status=$$(git status --porcelain -- examples/ proto/); \
	if [ -n "$$status" ]; then echo "$$status"; exit 1; fi

.PHONY: fmt
fmt: fmt-go fmt-proto ## Format code

.PHONY: fmt-proto
fmt-proto: ## Format proto files
	buf format -w

.PHONY: fmt-go
fmt-go: ## Format Go files
	golangci-lint fmt

.PHONY: install
install: ## Install protoc-gen-cli
	CGO_ENABLED=0 go install -trimpath -ldflags '-X main.version=$(VERSION)' ./cmd/protoc-gen-cli

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	go get -u -t ./... && go mod tidy -v

.PHONY: release-check
release-check: ## Validate the GoReleaser config
	go run github.com/goreleaser/goreleaser/v2@latest check

.PHONY: release-snapshot
release-snapshot: ## Build a local release snapshot without publishing
	go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean

.PHONY: clean
clean: ## Delete build artifacts
	rm -rf .tmp dist
	rm -rf examples/bookstore/go-cobra/gen
	rm -rf examples/bookstore-annotated/go-cobra/gen
	rm -rf examples/kitchen-sink/go-cobra/gen
	rm -rf examples/connect/go-cobra/gen
