#!/usr/bin/env bash
#
# scripts/copy-to-npm.sh
#
# Copies pre-built Go binaries from the GoReleaser dist/ directory
# into the npm package directory structure.
#
# GoReleaser output layout:
#   dist/deepscanbot_linux_amd64_v1/deepscanbot
#   dist/deepscanbot_linux_arm64_v8.0/deepscanbot
#   dist/deepscanbot_darwin_amd64_v1/deepscanbot
#   dist/deepscanbot_darwin_arm64_v8.0/deepscanbot
#   dist/deepscanbot_windows_amd64_v1/deepscanbot.exe
#   dist/checksums.txt
#
# The npm package keeps the same subdirectory structure so postinstall.js
# can dynamically find the correct binary for the user's platform.
#
# Usage:
#   bash scripts/copy-to-npm.sh                    # copies all binaries from dist/
#   VERSION=1.0.0 bash scripts/copy-to-npm.sh      # with version

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
GORELEASER_DIST="${PROJECT_ROOT}/dist"
NPM_DIST="${PROJECT_ROOT}/dist"
BINARY_NAME="deepscanbot"
PROJECT_NAME="deepscanbot"

echo "[copy-to-npm] Starting binary copy process..."
echo "[copy-to-npm] Source: ${GORELEASER_DIST}"
echo "[copy-to-npm] Target: ${NPM_DIST}"

# Ensure npm dist directory exists

mkdir -p "${NPM_DIST}"

# List all GoReleaser output directories
echo "[copy-to-npm] GoReleaser output directories:"
BINARY_COUNT=0
for dir in "${GORELEASER_DIST}/${PROJECT_NAME}"_*/; do
    if [ -d "$dir" ]; then
        dirname=$(basename "$dir")
        echo "[copy-to-npm]   Found directory: ${dirname}"

        # Find the binary inside (my-cli or my-cli.exe)
        binary_path=""
        if [ -f "${dir}${BINARY_NAME}" ]; then
            binary_path="${dir}${BINARY_NAME}"
        elif [ -f "${dir}${BINARY_NAME}.exe" ]; then
            binary_path="${dir}${BINARY_NAME}.exe"
        fi

        if [ -n "$binary_path" ]; then
            # Copy the entire directory to npm dist/
            cp -r "$dir" "${NPM_DIST}/"
            chmod -R 755 "${NPM_DIST}/${dirname}"
            echo "[copy-to-npm]   ✓ Copied directory: ${dirname}"
            BINARY_COUNT=$((BINARY_COUNT + 1))
        else
            echo "[copy-to-npm]   ⚠ No binary found in ${dirname}, skipping"
        fi
    fi
done

# Also copy checksums if they exist
if [ -f "${GORELEASER_DIST}/checksums.txt" ]; then
    cp "${GORELEASER_DIST}/checksums.txt" "${NPM_DIST}/checksums.txt"
    echo "[copy-to-npm] ✓ Copied: checksums.txt"
fi

# Verify we got all expected platform directories
declare -A PLATFORM_MAP
PLATFORM_MAP["linux_amd64"]=""
PLATFORM_MAP["linux_arm64"]=""
PLATFORM_MAP["darwin_amd64"]=""
PLATFORM_MAP["darwin_arm64"]=""
PLATFORM_MAP["windows_amd64"]=""

echo "[copy-to-npm] Verifying platform directories..."
MISSING=0
for platform in "${!PLATFORM_MAP[@]}"; do
    found=false
    for dir in "${NPM_DIST}/${PROJECT_NAME}"_*/; do
        if [ -d "$dir" ]; then
            dirname=$(basename "$dir")
            # Check if directory name contains this platform's os and arch
            os="${platform%_*}"
            arch="${platform#*_}"
            if echo "$dirname" | grep -qi "${os}" && echo "$dirname" | grep -qi "${arch}"; then
                found=true
                binary_file=""
                if [ -f "${dir}${BINARY_NAME}" ]; then
                    binary_file="${dir}${BINARY_NAME}"
                elif [ -f "${dir}${BINARY_NAME}.exe" ]; then
                    binary_file="${dir}${BINARY_NAME}.exe"
                fi
                if [ -n "$binary_file" ]; then
                    filesize=$(stat -c%s "$binary_file" 2>/dev/null || stat -f%z "$binary_file" 2>/dev/null)
                    echo "[copy-to-npm]   ✓ ${platform} -> ${dirname}/${BINARY_NAME} (${filesize} bytes)"
                else
                    echo "[copy-to-npm]   ✗ ${platform} -> ${dirname} (missing binary)"
                    MISSING=$((MISSING + 1))
                fi
                break
            fi
        fi
    done
    if [ "$found" = false ]; then
        echo "[copy-to-npm]   ✗ ${platform} MISSING"
        MISSING=$((MISSING + 1))
    fi
done

if [ $MISSING -gt 0 ]; then
    echo "[copy-to-npm] WARNING: ${MISSING} expected platform(s) are missing!"
    echo "[copy-to-npm] NPM publish will still proceed, but some platforms may not work."
else
    echo "[copy-to-npm] All expected platform directories present."
fi

echo "[copy-to-npm] Copy process complete. ${BINARY_COUNT} directories processed."