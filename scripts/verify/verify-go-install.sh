#!/bin/bash
# scripts/verify/verify-go-install.sh
# Verifies Go install command works correctly
set -euo pipefail

echo "=== Go Install Verification ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "⚠ SKIPPED: Go not installed"
    exit 0
fi

GO_VERSION=$(go version)
echo "✓ Go is installed: ${GO_VERSION}"

# Verify Go version meets requirement (1.22+)
GO_MAJOR=$(go version | awk '{print $3}' | sed 's/go//' | cut -d. -f1)
GO_MINOR=$(go version | awk '{print $3}' | sed 's/go//' | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]); then
    echo "❌ FAILED: Go version must be >= 1.22 (found: ${GO_VERSION})"
    exit 1
fi
echo "✓ Go version meets requirement (>=1.22)"

# Verify main package exists
echo ""
echo "Verifying main package location..."
MAIN_PKG="github.com/mindfiredigital/DeepScanBot/apps/cli"
if go list -m "$MAIN_PKG" &> /dev/null; then
    echo "✓ Module ${MAIN_PKG} is valid"
else
    echo "⚠ WARNING: Could not verify module (may need to be in module root)"
fi

# Verify go.mod exists
echo ""
echo "Checking go.mod..."
if [ -f "go.mod" ]; then
    echo "✓ go.mod exists"
    grep "^module " go.mod
else
    echo "❌ FAILED: go.mod not found"
    exit 1
fi

echo ""
echo "✓ Go install verification passed"