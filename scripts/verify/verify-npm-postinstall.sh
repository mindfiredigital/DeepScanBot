#!/bin/bash
# scripts/verify/verify-npm-postinstall.sh
# Verifies npm postinstall script works correctly
set -euo pipefail

echo "=== NPM Postinstall Script Verification ==="

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "⚠ SKIPPED: Node.js not installed"
    exit 0
fi

NODE_VERSION=$(node --version)
echo "✓ Node.js is installed: ${NODE_VERSION}"

# Check Node.js version meets requirement (>=18.0.0)
NODE_MAJOR=$(node -e "console.log(process.versions.node.split('.')[0])")
if [ "$NODE_MAJOR" -lt 18 ]; then
    echo "❌ FAILED: Node.js version must be >= 18.0.0 (found: ${NODE_VERSION})"
    exit 1
fi
echo "✓ Node.js version meets requirement (>=18.0.0)"

# Check postinstall.js syntax
echo ""
echo "Checking postinstall.js syntax..."
if node --check postinstall.js; then
    echo "✓ postinstall.js syntax is valid"
else
    echo "❌ FAILED: postinstall.js has syntax errors"
    exit 1
fi

# Check bin directory exists or dist exists
echo ""
echo "Checking binary distribution..."
if [ -d "dist" ] && [ "$(ls -A dist 2>/dev/null)" ]; then
    echo "✓ dist/ directory exists with binaries"
elif [ -d "bin" ] && [ -f "bin/deepscanbot" ]; then
    echo "✓ bin/ directory exists with binary"
else
    echo "⚠ WARNING: No binary found in dist/ or bin/. This is expected before GoReleaser build."
fi

echo ""
echo "✓ NPM postinstall verification passed"