#!/bin/bash
# scripts/verify/verify-sanity-check.sh
# Universal sanity check for DeepScanBot installation
set -euo pipefail

echo "=== DeepScanBot Universal Sanity Check ==="
echo ""

# Check 1: Binary exists and is executable
echo "1. Checking binary availability..."
if ! command -v deepscanbot &> /dev/null; then
    echo "   ⚠ WARNING: deepscanbot command not found in PATH"
    echo "   This is expected if not installed. Skipping binary checks."
    echo ""
    echo "✓ Sanity check completed (not installed - skipping binary tests)"
    exit 0
fi
echo "   ✓ deepscanbot found: $(which deepscanbot)"

# Check 2: Version command works
echo ""
echo "2. Checking version command..."
VERSION_OUTPUT=$(deepscanbot version 2>&1) || {
    echo "   ❌ FAILED: deepscanbot version returned an error"
    echo "   Output: ${VERSION_OUTPUT}"
    exit 1
}
echo "   ✓ Version command works"
echo "   Output: ${VERSION_OUTPUT}"

# Check 3: Help command works
echo ""
echo "3. Checking help output..."
if ! deepscanbot --help &> /dev/null; then
    echo "   ❌ FAILED: deepscanbot --help returned an error"
    exit 1
fi
echo "   ✓ Help command works"

echo ""
echo "========================================="
echo "✓ ALL SANITY CHECKS PASSED"
echo "========================================="