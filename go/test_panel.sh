#!/bin/bash

# Simple test to check if tactical panel renders

echo "Testing tactical panel rendering..."
echo "===================================="
echo ""

cd "$(dirname "$0")"

# Run for 3 seconds and capture output
timeout 3s ./sliver-graph 2>&1 | head -20

echo ""
echo "If you see the tactical panel on the right with stats, it's working!"
echo "If not, the issue might be terminal size detection."
echo ""
echo "Try running: ./sliver-graph directly and resize your terminal"
