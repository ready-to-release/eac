#!/bin/bash
# EAC CLI Installer for Linux and macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash
#    or: ./install.sh [--system] [--version VERSION]

set -euo pipefail

# Configuration
REPO="ready-to-release/eac"
BINARY_NAME="eac"
USER_INSTALL_DIR="$HOME/.local/bin"
SYSTEM_INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
INSTALL_DIR=""
SYSTEM_INSTALL=false
USE_UPX=false
VERSION="latest"

while [[ $# -gt 0 ]]; do
    case $1 in
        --system)
            SYSTEM_INSTALL=true
            shift
            ;;
        --upx)
            USE_UPX=true
            shift
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help|-h)
            echo "EAC CLI Installer"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --system          Install to $SYSTEM_INSTALL_DIR (requires sudo)"
            echo "  --upx             Install UPX-compressed binary (smaller, slightly slower startup)"
            echo "  --version VER     Install specific version (default: latest)"
            echo "  --install-dir DIR Custom installation directory"
            echo "  --help, -h        Show this help message"
            echo ""
            echo "Default install location: $USER_INSTALL_DIR"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Determine install directory (custom > system > user)
if [[ -z "$INSTALL_DIR" ]]; then
    if $SYSTEM_INSTALL; then
        INSTALL_DIR="$SYSTEM_INSTALL_DIR"
    else
        INSTALL_DIR="$USER_INSTALL_DIR"
    fi
fi

# Detect OS and architecture
detect_platform() {
    local os arch

    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $os${NC}"
            exit 1
            ;;
    esac

    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $arch${NC}"
            exit 1
            ;;
    esac

    echo -e "${BLUE}Detected platform: ${OS}/${ARCH}${NC}"
}

# Get the latest release version from GitHub
get_latest_version() {
    # Skip API call in test mode
    if [[ "${__EAC_TEST_MOCK:-}" == "1" ]]; then
        echo -e "${BLUE}Test mode: Using mock version eac-cli/v0.0.0-test${NC}" >&2
        echo "eac-cli/v0.0.0-test"
        return
    fi

    local latest
    # Fetch releases and find the latest eac-cli/* release (monorepo has multiple release tags)
    latest=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases?per_page=20" | grep '"tag_name":' | grep 'eac-cli/' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ -z "$latest" ]]; then
        echo -e "${RED}No eac-cli release found${NC}"
        exit 1
    fi
    echo "$latest"
}

# Download and install the binary
install_binary() {
    local version="$1"
    local download_url
    local tmp_dir

    # Construct binary name based on platform
    # Format: eac-{os}-{arch} for all platforms
    # Add -upx suffix if UPX version requested (only available for linux-amd64)
    local binary_filename
    if $USE_UPX && [[ "$OS" == "linux" ]] && [[ "$ARCH" == "amd64" ]]; then
        binary_filename="eac-${OS}-${ARCH}-upx"
    else
        binary_filename="eac-${OS}-${ARCH}"
        if $USE_UPX && [[ "$OS" != "linux" || "$ARCH" != "amd64" ]]; then
            echo -e "${YELLOW}Note: UPX binaries only available for linux-amd64, using standard binary${NC}"
        fi
    fi

    if $USE_UPX; then
        echo -e "${BLUE}Binary type: UPX-compressed (smaller, slightly slower startup)${NC}"
    else
        echo -e "${BLUE}Binary type: Standard${NC}"
    fi

    download_url="https://github.com/${REPO}/releases/download/${version}/${binary_filename}"

    # Create temporary directory
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Test mode: Use pre-built binary from out/build instead of downloading
    if [[ "${__EAC_TEST_MOCK:-}" == "1" ]]; then
        # In test mode, still validate that version looks realistic
        # Reject obviously invalid versions like v999.999.999
        if [[ "$version" =~ v999\. ]]; then
            echo -e "${BLUE}Test mode: Simulating download failure for invalid version${NC}"
            echo -e "${RED}Failed to download binary${NC}"
            echo -e "${YELLOW}Please check if the release exists: https://github.com/${REPO}/releases/tag/${version}${NC}"
            exit 1
        fi

        echo -e "${BLUE}Test mode: Using pre-built binary from out/build (skipping download)${NC}"

        # Use the actual built eac-cli binary from the build output
        # This is available because eac-installer depends on eac-cli module
        # When running from build output: out/build/eac-installer/bash-scripts/
        # Go up 2 levels to out/build, then access eac-cli/go-go/
        script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        built_binary="${script_dir}/../../eac-cli/go-go/eac-${OS}-${ARCH}"

        if [[ ! -f "$built_binary" ]]; then
            echo -e "${RED}Test mode: Pre-built binary not found at ${built_binary}${NC}"
            echo -e "${YELLOW}Ensure eac-cli module is built before running installer tests${NC}"
            exit 1
        fi

        # Copy the actual binary
        cp "$built_binary" "$tmp_dir/$BINARY_NAME"
    else
        # Real download in production mode
        echo -e "${BLUE}Downloading EAC CLI ${version}...${NC}"
        echo -e "${BLUE}URL: ${download_url}${NC}"

        # Download binary (show progress bar)
        if ! curl -fL --progress-bar "$download_url" -o "$tmp_dir/$BINARY_NAME"; then
            echo -e "${RED}Failed to download binary${NC}"
            echo -e "${YELLOW}Please check if the release exists: https://github.com/${REPO}/releases/tag/${version}${NC}"
            exit 1
        fi
    fi

    # Make executable
    chmod +x "$tmp_dir/$BINARY_NAME"

    # Create install directory if needed
    if [[ ! -d "$INSTALL_DIR" ]]; then
        if $SYSTEM_INSTALL; then
            sudo mkdir -p "$INSTALL_DIR"
        else
            mkdir -p "$INSTALL_DIR"
        fi
    fi

    # Install binary
    echo -e "${BLUE}Installing to ${INSTALL_DIR}/${BINARY_NAME}...${NC}"
    if $SYSTEM_INSTALL; then
        sudo mv "$tmp_dir/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    else
        mv "$tmp_dir/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    fi

    echo -e "${GREEN}Successfully installed EAC CLI ${version}${NC}"
}

# Check if install directory is in PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
        echo ""
        echo "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo -e "  ${GREEN}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
        echo ""
        echo "Then restart your shell or run: source ~/.bashrc"
    fi
}

# Main installation flow
main() {
    echo -e "${GREEN}EAC CLI Installer${NC}"
    echo ""

    detect_platform

    # Resolve version
    if [[ "$VERSION" == "latest" ]]; then
        echo -e "${BLUE}Fetching latest version...${NC}"
        VERSION=$(get_latest_version)
    fi
    echo -e "${BLUE}Version: ${VERSION}${NC}"

    # Install
    install_binary "$VERSION"

    # Verify installation
    if "$INSTALL_DIR/$BINARY_NAME" --version &>/dev/null; then
        echo ""
        echo -e "${GREEN}Installation verified!${NC}"
        "$INSTALL_DIR/$BINARY_NAME" --version
    fi

    # Check PATH
    if ! $SYSTEM_INSTALL; then
        check_path
    fi

    echo ""
    echo -e "${GREEN}Done! Run 'eac --help' to get started.${NC}"
}

main
