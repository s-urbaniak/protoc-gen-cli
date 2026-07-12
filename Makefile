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
	go build -ldflags '-X main.version=$(VERSION)' -o $(BIN)/protoc-gen-cli ./cmd/protoc-gen-cli

.PHONY: test
test: build ## Run tests
	go test -vet=off -race -cover ./...

.PHONY: lint
lint: ## Lint
	go vet ./...
	golangci-lint run --modules-download-mode=readonly --timeout=3m0s

.PHONY: lintfix
lintfix: ## Fix lint errors
	golangci-lint run --fix --modules-download-mode=readonly --timeout=3m0s

.PHONY: install
install: ## Install protoc-gen-cli
	go install -ldflags '-X main.version=$(VERSION)' ./cmd/protoc-gen-cli

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	go get -u -t ./... && go mod tidy -v

.PHONY: clean
clean: ## Delete build artifacts
	rm -rf .tmp dist
