#!/bin/bash
# scripts/verify/verify-github-release.sh
# Verifies GitHub Release assets exist and are correct
set -euo pipefail

REPO_OWNER="mindfiredigital"
REPO_NAME="DeepScanBot"
VERSION="${VERSION:-v1.0.0}"

echo "=== GitHub Releases Asset Verification ==="

# Fetch release metadata
echo "Fetching release metadata for ${VERSION}..."
RELEASE_DATA=$(curl -fsSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/tags/${VERSION}") || {
    echo "❌ FAILED: Could not fetch release ${VERSION}"
    exit 1
}
echo "✓ Release metadata fetched successfully"

# Required assets
REQUIRED_ASSETS=(
    "deepscanbot_linux_amd64"
    "deepscanbot_linux_arm64"
    "deepscanbot_darwin_amd64"
    "deepscanbot_darwin_arm64"
    "deepscanbot_windows_amd64.exe"
    "checksums.txt"
)

echo ""
echo "Checking for required assets..."
MISSING=0
for asset in "${REQUIRED_ASSETS[@]}"; do
    if echo "$RELEASE_DATA" | grep -q "\"name\": \"${asset}\""; then
        echo "  ✓ ${asset}"
    else
        echo "  ❌ MISSING: ${asset}"
        MISSING=$((MISSING + 1))
    fi
done

if [ $MISSING -gt 0 ]; then
    echo ""
    echo "❌ FAILED: ${MISSING} required asset(s) missing"
    exit 1
fi

echo ""
echo "✓ All required assets present in release ${VERSION}"