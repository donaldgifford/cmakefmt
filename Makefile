BINARY := cmakefmt
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: all build test lint clean install

all: build

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/cmakefmt

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

install:
	go install $(LDFLAGS) ./cmd/cmakefmt

# Format all CMake files in the project (dogfooding).
fmt:
	bin/$(BINARY) -i testdata/

# Check formatting (for CI).
check:
	bin/$(BINARY) -check testdata/
