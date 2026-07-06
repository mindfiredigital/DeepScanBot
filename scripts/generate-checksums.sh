#!/usr/bin/env bash
#
# scripts/generate-checksums.sh
#
# Generates SHA256 checksums for all binaries in the npm dist/ directory.
#
# GoReleaser output layout:
#   dist/deepscanbot_linux_amd64_v1/my-cli
#   dist/deepscanbot_linux_arm64_v8.0/my-cli
#   dist/deepscanbot_darwin_amd64_v1/my-cli
#   dist/deepscanbot_darwin_arm64_v8.0/my-cli
#   dist/deepscanbot_windows_amd64_v1/my-cli.exe
#
# This script generates checksums for the binary files inside each
# platform subdirectory, using the directory name as the checksum key
# (matching GoReleaser's convention).
#
# Usage:
#   bash scripts/generate-checksums.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DIST_DIR="${PROJECT_ROOT}/dist"
CHECKSUMS_FILE="${DIST_DIR}/checksums.txt"
BINARY_NAME="my-cli"
PROJECT_NAME="deepscanbot"

echo "[generate-checksums] Generating SHA256 checksums..."
echo "[generate-checksums] Directory: ${DIST_DIR}"

if [ ! -d "${DIST_DIR}" ]; then
    echo "[generate-checksums] ERROR: dist/ directory not found!"
    echo "[generate-checksums] Run GoReleaser first to generate binaries."
    exit 1
fi

# Check if sha256sum or shasum is available
if command -v sha256sum &>/dev/null; then
    SHASUM_CMD="sha256sum"
elif command -v shasum &>/dev/null; then
    SHASUM_CMD="shasum -a 256"
else
    echo "[generate-checksums] ERROR: Neither sha256sum nor shasum found."
    echo "[generate-checksums] Install coreutils (Linux) or ensure shasum is available (macOS)."
    exit 1
fi

# Remove existing checksums file if it exists (we regenerate it)
rm -f "${CHECKSUMS_FILE}"

# Generate checksums for binaries inside each platform subdirectory
COUNT=0
for dir in "${DIST_DIR}/${PROJECT_NAME}"_*/; do
    if [ -d "$dir" ]; then
        dirname=$(basename "$dir")

        # Find the binary inside
        binary_path=""
        if [ -f "${dir}${BINARY_NAME}" ]; then
            binary_path="${dir}${BINARY_NAME}"
        elif [ -f "${dir}${BINARY_NAME}.exe" ]; then
            binary_path="${dir}${BINARY_NAME}.exe"
        fi

        if [ -n "$binary_path" ]; then
            # Generate checksum using the directory name as the key
            # (matching GoReleaser's convention: <hash>  <dirname>)
            ${SHASUM_CMD} "$binary_path" | awk -v dirname="$dirname" '{print $1, "  ", dirname}' >> "${CHECKSUMS_FILE}"
            echo "[generate-checksums]   ✓ ${dirname}/${BINARY_NAME}"
            COUNT=$((COUNT + 1))
        else
            echo "[generate-checksums]   ⚠ No binary found in ${dirname}, skipping"
        fi
    fi
done

if [ ${COUNT} -eq 0 ]; then
    echo "[generate-checksums] WARNING: No platform directories found matching ${PROJECT_NAME}_*"
    echo "[generate-checksums] Creating empty checksums file."
    touch "${CHECKSUMS_FILE}"
else
    echo "[generate-checksums] Generated checksums for ${COUNT} platform directories."
    echo "[generate-checksums] Output: ${CHECKSUMS_FILE}"
fi

echo "[generate-checksums] Done."