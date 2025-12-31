#!/bin/bash

# Sliver Graphs Release Script
# This script helps create a new GitHub release

set -e

echo "=========================================="
echo "   Sliver Graphs Release Helper"
echo "=========================================="
echo ""

# Check if we're on the right branch
CURRENT_BRANCH=$(git branch --show-current)
echo "[*] Current branch: $CURRENT_BRANCH"

# Get version number
echo ""
read -p "Enter version number (e.g., 1.0.0): " VERSION
if [ -z "$VERSION" ]; then
    echo "[!] Version number required"
    exit 1
fi

TAG="v$VERSION"
echo "[*] Creating release: $TAG"

# Pre-release checklist
echo ""
echo "=========================================="
echo "   Pre-Release Checklist"
echo "=========================================="
echo ""
echo "✓ Code is tested and working"
echo "✓ README.md is up to date"
echo "✓ All changes are committed"
echo "✓ Branch is up to date with remote"
echo ""
read -p "Continue with release? (y/N): " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "[!] Release cancelled"
    exit 0
fi

# Build the binary
echo ""
echo "[*] Building binary..."
bash build.sh

if [ ! -f "sliver-graph" ]; then
    echo "[!] Build failed - binary not found"
    exit 1
fi

echo "[+] Binary built successfully"

# Create git tag
echo ""
echo "[*] Creating git tag: $TAG"
git tag -a "$TAG" -m "Release $TAG"

echo "[+] Tag created"

# Push to remote
echo ""
read -p "Push tag to GitHub? (y/N): " PUSH_CONFIRM
if [[ "$PUSH_CONFIRM" =~ ^[Yy]$ ]]; then
    echo "[*] Pushing tag to origin..."
    git push origin "$TAG"
    echo "[+] Tag pushed"
fi

# Create GitHub release
echo ""
read -p "Create GitHub release? (requires gh cli) (y/N): " RELEASE_CONFIRM
if [[ "$RELEASE_CONFIRM" =~ ^[Yy]$ ]]; then
    if ! command -v gh &> /dev/null; then
        echo "[!] GitHub CLI (gh) not found. Install from: https://cli.github.com"
        echo "[*] You can create the release manually at:"
        echo "    https://github.com/musyoka101/sliver-graphs/releases/new"
        exit 1
    fi
    
    echo "[*] Creating GitHub release..."
    
    if [ -f "RELEASE_NOTES.md" ]; then
        gh release create "$TAG" \
            --title "Sliver C2 Network Topology Visualizer $TAG" \
            --notes-file RELEASE_NOTES.md \
            sliver-graph
    else
        gh release create "$TAG" \
            --title "Sliver C2 Network Topology Visualizer $TAG" \
            --generate-notes \
            sliver-graph
    fi
    
    echo "[+] Release created successfully!"
    echo ""
    echo "View release at: https://github.com/musyoka101/sliver-graphs/releases/tag/$TAG"
else
    echo ""
    echo "[*] To create release manually:"
    echo "    1. Go to: https://github.com/musyoka101/sliver-graphs/releases/new"
    echo "    2. Select tag: $TAG"
    echo "    3. Add release notes from RELEASE_NOTES.md"
    echo "    4. Upload binary: sliver-graph"
    echo "    5. Publish release"
fi

echo ""
echo "=========================================="
echo "   Release Complete!"
echo "=========================================="
