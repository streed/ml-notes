# Makefile for ML Notes

# Variables
CLI_BINARY_NAME := ml-notes-cli
GUI_BINARY_NAME := ml-notes
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build variables
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
CGO_ENABLED := 1
GOFLAGS := -v

# Go binary paths
GOPATH := $(shell go env GOPATH)
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
	GOBIN := $(GOPATH)/bin
endif

# Add Go bin to PATH for this Makefile
export PATH := $(PATH):$(GOBIN)

# Check if Wails is available (after adding GOBIN to PATH)
WAILS_AVAILABLE := $(shell command -v wails 2> /dev/null)

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
all: build-cli build-gui install

# Build both CLI and GUI binaries
.PHONY: build
build: build-cli build-gui

# Build the CLI binary
.PHONY: build-cli
build-cli:
	@echo "Building $(CLI_BINARY_NAME) $(VERSION) for $(PLATFORM)/$(ARCH)..."
	@echo "Go version: $(GO_VERSION)"
	@echo "Git commit: $(GIT_COMMIT)"
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(CLI_BINARY_NAME) ./app/cli
	@echo "CLI build complete: ./$(CLI_BINARY_NAME)"

# Build the GUI binary using Wails
.PHONY: build-gui
build-gui:
ifdef WAILS_AVAILABLE
	@echo "Building $(GUI_BINARY_NAME) $(VERSION) using Wails..."
	@echo "Go version: $(GO_VERSION)"
	@echo "Git commit: $(GIT_COMMIT)"
	wails build -clean -o $(GUI_BINARY_NAME)
	@echo "GUI build complete: ./build/bin/$(GUI_BINARY_NAME)"
else
	@echo "⚠️  Wails not found. Skipping GUI build."
	@echo "   Install Wails with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
endif

# Development build with race detector for CLI
.PHONY: dev-cli
dev-cli:
	@echo "Building CLI development version with race detector..."
	CGO_ENABLED=1 go build -race $(LDFLAGS) -o $(CLI_BINARY_NAME)-dev ./app/cli
	@echo "CLI development build complete: ./$(CLI_BINARY_NAME)-dev"

# Development build for GUI using Wails
.PHONY: dev-gui
dev-gui:
ifdef WAILS_AVAILABLE
	@echo "Starting Wails development server..."
	wails dev
else
	@echo "⚠️  Wails not found. Cannot start development server."
	@echo "   Install Wails with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
endif

# Development build for both
.PHONY: dev
dev: dev-cli dev-gui

# Install binaries to system PATH
.PHONY: install
install: $(CLI_BINARY_NAME) $(GUI_BINARY_NAME)
	@echo "Installing $(CLI_BINARY_NAME) to $(BINDIR)..."
	@$(INSTALL_PROGRAM) $(CLI_BINARY_NAME) $(BINDIR)/
ifdef WAILS_AVAILABLE
	@if [ -f "./build/bin/$(GUI_BINARY_NAME)" ]; then \
		echo "Installing $(GUI_BINARY_NAME) to $(BINDIR)..."; \
		$(INSTALL_PROGRAM) ./build/bin/$(GUI_BINARY_NAME) $(BINDIR)/; \
	fi
endif
	@echo "Installation complete!"
	@echo "Run '$(CLI_BINARY_NAME) init' to set up your configuration."
	@echo "Run '$(GUI_BINARY_NAME)' to start the desktop application."

# Uninstall the binaries
.PHONY: uninstall
uninstall:
	@echo "Removing binaries from $(BINDIR)..."
	@rm -f $(BINDIR)/$(CLI_BINARY_NAME)
	@rm -f $(BINDIR)/$(GUI_BINARY_NAME)
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
	@rm -f $(CLI_BINARY_NAME) $(CLI_BINARY_NAME)-dev
	@rm -f $(GUI_BINARY_NAME) $(GUI_BINARY_NAME)-dev
	@rm -rf build/
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
	@echo "  make build          - Build both CLI and GUI binaries"
	@echo "  make build-cli      - Build the CLI binary only"
	@echo "  make build-gui      - Build the GUI binary using Wails"
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
	@echo "  make install        - Build and install both binaries to $(BINDIR)"
	@echo "  make uninstall      - Remove both binaries from $(BINDIR)"
	@echo "  make dev            - Build CLI with race detector"
	@echo "  make dev-cli        - Build CLI with race detector"
	@echo "  make dev-gui        - Start Wails development server"
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
	@echo "  CLI_BINARY_NAME=$(CLI_BINARY_NAME)"
	@echo "  GUI_BINARY_NAME=$(GUI_BINARY_NAME)"
	@echo "  PLATFORM=$(PLATFORM)/$(ARCH)"
	@echo "  PREFIX=$(PREFIX)"
ifdef WAILS_AVAILABLE
	@echo "  WAILS=available"
else
	@echo "  WAILS=not available (GUI builds disabled)"
endif
	@echo ""
	@echo "📝 Notes:"
	@echo "  - The CLI binary provides all command-line functionality"
	@echo "  - The GUI binary is a desktop app built with Wails"
	@echo "  - Wails is required for GUI builds: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
	@echo "  - Cross-compilation for macOS/Windows requires appropriate toolchains"
	@echo "  - For best results, build natively on target platforms"
	@echo "  - CGO is required for sqlite-vec support"

# Ensure binaries exist for install target
$(CLI_BINARY_NAME):
	@$(MAKE) build-cli

$(GUI_BINARY_NAME):
	@$(MAKE) build-gui

.DEFAULT_GOAL := help