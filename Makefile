VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
BINARY := quill

.PHONY: build test install clean lint

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/quill

test:
	go test ./... -v

install:
	go install $(LDFLAGS) ./cmd/quill

clean:
	rm -f $(BINARY)
	rm -rf dist/

lint:
	go vet ./...
