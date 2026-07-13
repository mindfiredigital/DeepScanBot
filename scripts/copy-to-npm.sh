#!/bin/bash
# copy-to-npm.sh
# Copies GoReleaser-built binaries from dist/ to bin/ for npm packaging

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_ROOT/dist"
BIN_DIR="$PROJECT_ROOT/bin"

echo "=== Copying binaries to npm package ==="

# Ensure bin directory exists
mkdir -p "$BIN_DIR"

# Find all platform-specific directories in dist/
# GoReleaser creates directories like: deepscanbot_<version>_<os>_<arch>
if [ ! -d "$DIST_DIR" ]; then
  echo "Error: dist/ directory not found. Run GoReleaser build first."
  exit 1
fi

# Copy all binaries from dist/ to bin/
# GoReleaser creates archives, but for npm we need the raw binaries
# The binaries are typically in archives, so we need to extract them
# or they might already be extracted depending on the GoReleaser config

# Find all executable files in dist/ and copy them to bin/
find "$DIST_DIR" -type f \( -name "deepscanbot" -o -name "deepscanbot.exe" \) | while read -r binary; do
  echo "Found binary: $binary"
  
  # Determine the target filename
  filename=$(basename "$binary")
  
  # Copy to bin directory
  cp "$binary" "$BIN_DIR/$filename"
  
  # Make executable on Unix systems
  if [[ "$filename" != *.exe ]]; then
    chmod +x "$BIN_DIR/$filename"
    echo "  → Copied to: $BIN_DIR/$filename (made executable)"
  else
    echo "  → Copied to: $BIN_DIR/$filename"
  fi
done

# Verify the copy
echo ""
echo "=== Binaries in bin/ directory ==="
ls -lah "$BIN_DIR" 2>/dev/null || echo "No binaries found in bin/"

echo ""
echo "=== Copy complete ==="