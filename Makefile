BINARY_MAIN    = qnaplcd
BINARY_TEST    = qnaplcd-test
DOCKER_IMAGE   = qnaplcd-builder

GOFLAGS        = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1
LDFLAGS        = -s -w
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: all build test vet fmt clean docker help

all: test build ## Run tests then build binaries (default)

build: $(BINARY_MAIN) $(BINARY_TEST) ## Build both binaries

$(BINARY_MAIN):
	$(GOFLAGS) go build -ldflags='$(LDFLAGS) -X main.version=$(VERSION)' -o $@ ./cmd/qnaplcd

$(BINARY_TEST):
	$(GOFLAGS) go build -ldflags='$(LDFLAGS)' -o $@ ./cmd/qnaplcd-test

test: ## Run all unit tests
	go test ./...

test-v: ## Run all unit tests (verbose)
	go test -v ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Run gofmt on all Go files
	gofmt -w .

docker: ## Build binaries via Docker
	./build.sh

clean: ## Remove build artifacts
	rm -f $(BINARY_MAIN) $(BINARY_TEST)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'
