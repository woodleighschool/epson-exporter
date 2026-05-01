BINARY ?= epson_exporter

.PHONY: all build test fmt vet tidy check clean

all: check build

build:
	go build -o bin/$(BINARY) ./cmd/epson_exporter

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt tidy test vet

clean:
	rm -rf bin dist coverage.out
