VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
BINARY := quill

.PHONY: build test coverage install clean lint

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/quill

test:
	go test ./... -v

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

install:
	go install $(LDFLAGS) ./cmd/quill

clean:
	rm -f $(BINARY)
	rm -rf dist/
	rm -f coverage.out coverage.html

lint:
	go vet ./...
