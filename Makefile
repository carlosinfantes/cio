# CIO - Chief Intelligence Officer - Build Makefile

BINARY_NAME=cio
VERSION?=1.0.0
BUILD_DIR=dist
GO=go

# Build flags for smaller binaries
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all build clean test test-coverage test-verbose install cross-compile

all: build

# Build for current platform
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/cio

# Build with debug info
build-debug:
	$(GO) build -o $(BINARY_NAME) ./cmd/cio

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GO) test ./...

# Run tests with verbose output
test-verbose:
	$(GO) test -v ./...

# Run tests with coverage report
test-coverage:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install to GOPATH/bin
install:
	$(GO) install $(LDFLAGS) ./cmd/cio

# Cross-compile for all platforms
cross-compile: clean
	mkdir -p $(BUILD_DIR)

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/cio

	# macOS AMD64 (Intel)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/cio

	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/cio

	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/cio

	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/cio

	# List built binaries
	@echo "\nBuilt binaries:"
	@ls -lh $(BUILD_DIR)/

# Build for current platform with size optimization
build-release:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/cio
	@echo "Binary size:"
	@ls -lh $(BINARY_NAME)

# Format code
fmt:
	$(GO) fmt ./...

# Lint code
lint:
	golangci-lint run

# Update dependencies
deps:
	$(GO) mod tidy
	$(GO) mod download

# Show help
help:
	@echo "CIO - Chief Intelligence Officer - Build Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build for current platform"
	@echo "  make build-release  Build optimized release binary"
	@echo "  make cross-compile  Build for all platforms"
	@echo "  make install        Install to GOPATH/bin"
	@echo "  make test           Run tests"
	@echo "  make test-verbose   Run tests with verbose output"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make clean          Remove build artifacts"
	@echo "  make fmt            Format code"
	@echo "  make lint           Run linter"
	@echo "  make deps           Update dependencies"
