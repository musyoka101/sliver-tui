#!/bin/bash
# Build script for Sliver Graph Go version

set -e

echo "[*] Building sliver-graph..."
go build -o sliver-graph main.go

echo "[+] Build complete: sliver-graph"
echo ""
echo "Usage:"
echo "  ./sliver-graph       # Run the TUI"
echo "  r - Refresh"
echo "  q - Quit"
