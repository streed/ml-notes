# Makefile for ML Notes

# Variables
BINARY_NAME := ml-notes
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build variables
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
CGO_ENABLED := 1
GOFLAGS := -v

# Directories
PREFIX := /usr/local
BINDIR := $(PREFIX)/bin
INSTALL := install
INSTALL_PROGRAM := $(INSTALL) -m 755

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
	PLATFORM := linux
endif
ifeq ($(UNAME_S),Darwin)
	PLATFORM := darwin
endif

ifeq ($(UNAME_M),x86_64)
	ARCH := amd64
endif
ifeq ($(UNAME_M),aarch64)
	ARCH := arm64
endif
ifeq ($(UNAME_M),arm64)
	ARCH := arm64
endif

# Default target
.PHONY: all
all: build install

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) $(VERSION) for $(PLATFORM)/$(ARCH)..."
	@echo "Go version: $(GO_VERSION)"
	@echo "Git commit: $(GIT_COMMIT)"
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

# Development build with race detector
.PHONY: dev
dev:
	@echo "Building development version with race detector..."
	CGO_ENABLED=1 go build -race $(LDFLAGS) -o $(BINARY_NAME)-dev .
	@echo "Development build complete: ./$(BINARY_NAME)-dev"

# Install the binary to system PATH
.PHONY: install
install: $(BINARY_NAME)
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@$(INSTALL_PROGRAM) $(BINARY_NAME) $(BINDIR)/
	@echo "Installation complete!"
	@echo "Run 'ml-notes init' to set up your configuration."

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Removing $(BINARY_NAME) from $(BINDIR)..."
	@rm -f $(BINDIR)/$(BINARY_NAME)
	@echo "Uninstall complete."

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		go vet ./...; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted."

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-dev
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "Clean complete."

# Update dependencies
.PHONY: deps
deps:
	@echo "Updating dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated."

# Build for all platforms (may require cross-compilation tools)
.PHONY: build-all
build-all: build-linux build-darwin-safe build-windows-safe

# Safe cross-platform builds (skip if tools missing)
.PHONY: build-darwin-safe
build-darwin-safe:
	@echo "Attempting macOS builds..."
	@$(MAKE) build-darwin || echo "⚠️  macOS cross-compilation failed (missing tools). Build on macOS for best results."

.PHONY: build-windows-safe  
build-windows-safe:
	@echo "Attempting Windows builds..."
	@$(MAKE) build-windows || echo "⚠️  Windows cross-compilation failed (missing tools). Build on Windows for best results."

# Native builds (when building on target platform)
.PHONY: build-native
build-native:
	@echo "Building for native platform: $(PLATFORM)/$(ARCH)..."
	@mkdir -p dist
	CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-$(PLATFORM)-$(ARCH)$(if $(filter windows,$(PLATFORM)),.exe) .
	@echo "Native build complete: dist/$(BINARY_NAME)-$(PLATFORM)-$(ARCH)$(if $(filter windows,$(PLATFORM)),.exe)"

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p dist
	@echo "  Building Linux AMD64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	@echo "Linux AMD64 build complete."

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p dist
	@echo "  Building macOS AMD64 (Intel)..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	@echo "  Building macOS ARM64 (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	@echo "macOS builds complete."

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p dist
	@echo "  Building Windows AMD64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Windows build complete."

# Create release packages
.PHONY: release
release: clean build-all
	@echo "Creating release packages..."
	@mkdir -p dist/release
	
	# Package only successfully built binaries
	@if [ -f "dist/$(BINARY_NAME)-linux-amd64" ]; then \
		echo "  📦 Packaging Linux AMD64..."; \
		tar -czf dist/release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C dist $(BINARY_NAME)-linux-amd64; \
	fi
	
	@if [ -f "dist/$(BINARY_NAME)-darwin-amd64" ]; then \
		echo "  📦 Packaging macOS AMD64 (Intel)..."; \
		tar -czf dist/release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C dist $(BINARY_NAME)-darwin-amd64; \
	fi
	
	@if [ -f "dist/$(BINARY_NAME)-darwin-arm64" ]; then \
		echo "  📦 Packaging macOS ARM64 (Apple Silicon)..."; \
		tar -czf dist/release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C dist $(BINARY_NAME)-darwin-arm64; \
	fi
	
	@if [ -f "dist/$(BINARY_NAME)-windows-amd64.exe" ]; then \
		echo "  📦 Packaging Windows AMD64..."; \
		cd dist && zip release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe; \
	fi
	
	@echo ""
	@echo "✅ Release packages created in dist/release/"
	@echo "📦 Available packages:"
	@ls -la dist/release/ 2>/dev/null | grep -E '\.(tar\.gz|zip)$$' || echo "   No packages created (check build output above)"

# Install development tools
.PHONY: tools
tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed."

# Show help
.PHONY: help
help:
	@echo "ML Notes - Makefile targets:"
	@echo ""
	@echo "🏗️  Build targets:"
	@echo "  make build          - Build the binary for current platform"
	@echo "  make build-native   - Build for native platform (auto-detect)"
	@echo "  make build-linux    - Build for Linux AMD64"
	@echo "  make build-darwin   - Build for macOS (Intel & Apple Silicon)"
	@echo "  make build-windows  - Build for Windows AMD64"
	@echo "  make build-all      - Build for all platforms (with fallback)"
	@echo ""
	@echo "📦 Package targets:"
	@echo "  make release        - Create release packages for all platforms"
	@echo ""
	@echo "🛠️  Development targets:"
	@echo "  make install        - Build and install to $(BINDIR)"
	@echo "  make uninstall      - Remove from $(BINDIR)"
	@echo "  make dev            - Build with race detector"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make deps           - Update dependencies"
	@echo "  make tools          - Install development tools"
	@echo ""
	@echo "ℹ️  Information:"
	@echo "  VERSION=$(VERSION)"
	@echo "  PLATFORM=$(PLATFORM)/$(ARCH)"
	@echo "  PREFIX=$(PREFIX)"
	@echo ""
	@echo "📝 Notes:"
	@echo "  - Cross-compilation for macOS/Windows requires appropriate toolchains"
	@echo "  - For best results, build natively on target platforms"
	@echo "  - CGO is required for sqlite-vec support"

# Ensure binary exists for install target
$(BINARY_NAME):
	@$(MAKE) build

.DEFAULT_GOAL := help