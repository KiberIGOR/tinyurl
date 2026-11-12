.PHONY: build shortener reset test

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS = -X main.buildVersion=$(VERSION) \
	-X main.buildDate=$(DATE) \
	-X main.buildCommit=$(COMMIT)

shortener:
	go build -buildvcs=false -ldflags "$(LDFLAGS)" -o shortener ./cmd/shortener

build: shortener

reset:
	go run ./cmd/reset

test:
	go test ./...
