#!/bin/bash
# Sliver TUI Installer
# Builds and installs the Sliver C2 Network Topology Visualizer

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="sliver-tui"
INSTALL_DIR="$HOME/.local/bin"
REPO_URL="https://github.com/musyoka101/sliver-graphs.git"

# Print colored message
print_msg() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print section header
print_header() {
    echo ""
    print_msg "$BLUE" "============================================"
    print_msg "$BLUE" "$1"
    print_msg "$BLUE" "============================================"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main installation
main() {
    print_header "Sliver TUI Installer"
    
    # Check for Go installation
    print_msg "$YELLOW" "Checking prerequisites..."
    if ! command_exists go; then
        print_msg "$RED" "❌ Go is not installed!"
        print_msg "$YELLOW" "Please install Go 1.21+ from: https://go.dev/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_msg "$GREEN" "✓ Go found: $GO_VERSION"
    
    # Check for git
    if ! command_exists git; then
        print_msg "$RED" "❌ Git is not installed!"
        exit 1
    fi
    print_msg "$GREEN" "✓ Git found"
    
    # Determine installation method
    if [ -f "main.go" ] && [ -f "go.mod" ]; then
        # Already in the repo directory
        print_msg "$GREEN" "✓ Already in sliver-graphs/go directory"
        BUILD_DIR="."
    else
        # Need to clone the repo
        print_header "Cloning Repository"
        TEMP_DIR=$(mktemp -d)
        print_msg "$YELLOW" "Cloning from: $REPO_URL"
        git clone "$REPO_URL" "$TEMP_DIR" 2>&1 | grep -v "^remote:" || true
        BUILD_DIR="$TEMP_DIR/go"
        print_msg "$GREEN" "✓ Repository cloned"
    fi
    
    # Build the binary
    print_header "Building Sliver TUI"
    print_msg "$YELLOW" "Downloading dependencies..."
    cd "$BUILD_DIR"
    go mod download
    
    print_msg "$YELLOW" "Compiling binary..."
    go build -o "$BINARY_NAME" -ldflags="-s -w" .
    
    if [ ! -f "$BINARY_NAME" ]; then
        print_msg "$RED" "❌ Build failed!"
        exit 1
    fi
    
    BINARY_SIZE=$(du -h "$BINARY_NAME" | cut -f1)
    print_msg "$GREEN" "✓ Build successful! ($BINARY_SIZE)"
    
    # Install the binary
    print_header "Installing Binary"
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        print_msg "$YELLOW" "Creating install directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi
    
    # Copy binary to install directory
    print_msg "$YELLOW" "Installing to: $INSTALL_DIR/$BINARY_NAME"
    cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        print_msg "$YELLOW" "⚠️  $INSTALL_DIR is not in your PATH"
        print_msg "$YELLOW" "Add this line to your ~/.bashrc or ~/.zshrc:"
        print_msg "$BLUE" "    export PATH=\"\$HOME/.local/bin:\$PATH\""
        print_msg "$YELLOW" "Then run: source ~/.bashrc (or ~/.zshrc)"
    fi
    
    print_msg "$GREEN" "✓ Installation complete!"
    
    # Check for Sliver config
    print_header "Configuration Check"
    if [ -d "$HOME/.sliver-client/configs" ]; then
        CONFIG_COUNT=$(find "$HOME/.sliver-client/configs" -name "*.cfg" 2>/dev/null | wc -l)
        if [ "$CONFIG_COUNT" -gt 0 ]; then
            print_msg "$GREEN" "✓ Found $CONFIG_COUNT Sliver config(s)"
        else
            print_msg "$YELLOW" "⚠️  No Sliver configs found in ~/.sliver-client/configs/"
            print_msg "$YELLOW" "Make sure you have a Sliver client config before running"
        fi
    else
        print_msg "$YELLOW" "⚠️  ~/.sliver-client/configs/ directory not found"
        print_msg "$YELLOW" "Make sure you have Sliver client installed and configured"
    fi
    
    # Cleanup if we cloned
    if [ "$BUILD_DIR" != "." ]; then
        print_msg "$YELLOW" "Cleaning up temporary files..."
        rm -rf "$TEMP_DIR"
    fi
    
    # Installation summary
    print_header "Installation Complete!"
    print_msg "$GREEN" "Binary installed to: $INSTALL_DIR/$BINARY_NAME"
    print_msg "$BLUE" ""
    print_msg "$BLUE" "To run Sliver TUI:"
    print_msg "$GREEN" "    $BINARY_NAME"
    print_msg "$BLUE" ""
    print_msg "$BLUE" "Keyboard shortcuts:"
    print_msg "$YELLOW" "    d - Dashboard view"
    print_msg "$YELLOW" "    v - Change view (Tree/List/Dashboard)"
    print_msg "$YELLOW" "    t - Change theme"
    print_msg "$YELLOW" "    r - Manual refresh"
    print_msg "$YELLOW" "    q - Quit"
    print_msg "$BLUE" ""
    print_msg "$GREEN" "Enjoy! 🎉"
    echo ""
}

# Run main function
main
