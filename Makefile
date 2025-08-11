SHELL := /bin/bash

GO ?= go
PKGS := ./...

.PHONY: all tidy deps build test lint release

all: tidy build test

 tidy:
	$(GO) mod tidy

 deps:
	$(GO) mod download

 build:
	$(GO) build $(PKGS)

 test:
	$(GO) test -race -count=1 $(PKGS)

 lint:
	golangci-lint run

 release:
	goreleaser release --clean
