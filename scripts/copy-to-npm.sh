#!/bin/bash
# copy-to-npm.sh
# Copies GoReleaser-built binaries from dist/ to bin/ for npm packaging
# Optimized to only copy the binary for the current platform to reduce package size

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_ROOT/dist"
BIN_DIR="$PROJECT_ROOT/bin"

echo "=== Copying binaries to npm package ==="

# Ensure bin directory exists
mkdir -p "$BIN_DIR"

# Check if dist/ directory exists
if [ ! -d "$DIST_DIR" ]; then
  echo "Error: dist/ directory not found. Run GoReleaser build first."
  exit 1
fi

# Detect current platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map OS names
case "$OS" in
  linux*)   OS="linux" ;;
  darwin*)  OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)       
    echo "Error: Unsupported operating system: $OS"
    exit 1
    ;;
esac

# Map architecture names
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)       
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "Detected platform: $OS/$ARCH"

# Determine binary name
BINARY_NAME="deepscanbot"
if [ "$OS" = "windows" ]; then
  BINARY_NAME="deepscanbot.exe"
fi

# Find the matching binary in dist/
# GoReleaser creates directories like: deepscanbot_<version>_<os>_<arch>
TARGET_BINARY=""
for dir in "$DIST_DIR"/*; do
  if [ -d "$dir" ]; then
    dirname="$(basename "$dir")"
    # Check if directory name contains OS and ARCH
    if [[ "$dirname" == *"$OS"* ]] && [[ "$dirname" == *"$ARCH"* ]]; then
      TARGET_BINARY="$dir/$BINARY_NAME"
      if [ -f "$TARGET_BINARY" ]; then
        break
      fi
    fi
  fi
done

if [ -z "$TARGET_BINARY" ] || [ ! -f "$TARGET_BINARY" ]; then
  echo "Error: Binary for $OS/$ARCH not found in dist/"
  echo "Available directories:"
  ls -1 "$DIST_DIR"
  exit 1
fi

echo "Found binary: $TARGET_BINARY"

# Copy to bin directory
cp "$TARGET_BINARY" "$BIN_DIR/$BINARY_NAME"

# Make executable on Unix systems
if [[ "$BINARY_NAME" != *.exe ]]; then
  chmod +x "$BIN_DIR/$BINARY_NAME"
  echo "  → Copied to: $BIN_DIR/$BINARY_NAME (made executable)"
else
  echo "  → Copied to: $BIN_DIR/$BINARY_NAME"
fi

# Verify the copy
echo ""
echo "=== Binaries in bin/ directory ==="
ls -lah "$BIN_DIR"

echo ""
echo "=== Copy complete ==="
echo "Note: Only the binary for the current platform ($OS/$ARCH) was copied to minimize package size."
