#!/usr/bin/env bash

# ML Notes Installation Script
# This script downloads and installs ml-notes to your system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="streed/ml-notes"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="ml-notes"

# Functions
print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${BLUE}→ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Detect OS and Architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux)
            PLATFORM="linux"
            ;;
        darwin)
            PLATFORM="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64)
            ARCHITECTURE="amd64"
            ;;
        aarch64|arm64)
            ARCHITECTURE="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    echo "${PLATFORM}-${ARCHITECTURE}"
}

# Check for required tools
check_requirements() {
    local missing_tools=()
    
    for tool in curl tar; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_info "Please install the missing tools and try again."
        exit 1
    fi
}

# Get latest release version
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
install_ml_notes() {
    local version="$1"
    local platform="$2"
    
    if [ -z "$version" ]; then
        version=$(get_latest_version)
        if [ -z "$version" ]; then
            print_warning "Could not determine latest version, using 'latest'"
            version="latest"
        fi
    fi
    
    print_info "Installing ML Notes ${version} for ${platform}..."
    
    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Download URL
    if [ "$version" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}-${platform}.tar.gz"
    else
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${version}-${platform}.tar.gz"
    fi
    
    print_info "Downloading from: $DOWNLOAD_URL"
    
    # Download the binary
    if ! curl -L -o "$TEMP_DIR/${BINARY_NAME}.tar.gz" "$DOWNLOAD_URL"; then
        print_error "Failed to download ML Notes"
        print_info "You can try building from source instead:"
        print_info "  git clone https://github.com/${REPO}.git"
        print_info "  cd ml-notes && make install"
        exit 1
    fi
    
    # Extract the binary
    print_info "Extracting archive..."
    tar -xzf "$TEMP_DIR/${BINARY_NAME}.tar.gz" -C "$TEMP_DIR"
    
    # Find the binary (it might be named differently)
    BINARY_PATH=$(find "$TEMP_DIR" -name "${BINARY_NAME}*" -type f | head -n 1)
    
    if [ -z "$BINARY_PATH" ]; then
        print_error "Binary not found in archive"
        exit 1
    fi
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        SUDO=""
    else
        SUDO="sudo"
        print_info "Root access required to install to $INSTALL_DIR"
    fi
    
    # Install the binary
    print_info "Installing to $INSTALL_DIR..."
    $SUDO install -m 755 "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    
    print_success "ML Notes installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        VERSION=$("$BINARY_NAME" version 2>/dev/null || echo "unknown")
        print_success "ML Notes is installed and accessible"
        print_info "Version: $VERSION"
        return 0
    else
        print_warning "ML Notes was installed but is not in your PATH"
        print_info "Add $INSTALL_DIR to your PATH or run: $INSTALL_DIR/$BINARY_NAME"
        return 1
    fi
}

# Post-installation setup
post_install() {
    echo ""
    print_info "Next steps:"
    echo "  1. Initialize configuration: ml-notes init"
    echo "  2. Add your first note: ml-notes add -t \"My Note\" -c \"Content\""
    echo "  3. List notes: ml-notes list"
    echo "  4. Search notes: ml-notes search --vector \"query\""
    echo ""
    print_info "For more information, run: ml-notes --help"
}

# Main installation flow
main() {
    echo "ML Notes Installation Script"
    echo "============================"
    echo ""
    
    # Parse arguments
    VERSION=""
    PLATFORM=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --platform)
                PLATFORM="$2"
                shift 2
                ;;
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --version VERSION     Install specific version (default: latest)"
                echo "  --platform PLATFORM   Override platform detection"
                echo "  --install-dir DIR     Installation directory (default: /usr/local/bin)"
                echo "  --help               Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Check requirements
    check_requirements
    
    # Detect platform if not specified
    if [ -z "$PLATFORM" ]; then
        PLATFORM=$(detect_platform)
        print_info "Detected platform: $PLATFORM"
    fi
    
    # Install
    install_ml_notes "$VERSION" "$PLATFORM"
    
    # Verify
    if verify_installation; then
        post_install
    fi
}

# Run main function
main "$@"