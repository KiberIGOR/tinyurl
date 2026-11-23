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

proto:
	protoc -I $$HOME/local/protoc-36.2/include -I api/proto \
		--go_out=. --go_opt=module=github.com/KiberIGOR/tinyurl \
		--go-grpc_out=. --go-grpc_opt=module=github.com/KiberIGOR/tinyurl \
		api/proto/shortener/v1/shortener.proto

test:
	go test ./...
