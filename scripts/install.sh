#!/usr/bin/env bash
#
# DeepScanBot CLI - Install Script
# ==================================
# Detects OS/Arch, fetches the latest binary from GitHub Releases,
# verifies SHA256 checksum, and installs to a standard system path.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.sh | bash -s -- -b /usr/local/bin
#
# Options:
#   -b <path>   Set the installation directory (default: /usr/local/bin)
#   -v <version> Install a specific version (default: latest)
#

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
REPO_OWNER="mindfiredigital"
REPO_NAME="DeepScanBot"
PROJECT_NAME="deepscanbot"
DEFAULT_INSTALL_DIR="/usr/local/bin"
GITHUB_API="https://api.github.com"
GITHUB_DL="https://github.com"

# ---------------------------------------------------------------------------
# Colors for output
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info()  { printf "${BLUE}%s${NC}\n" "[INFO] $*"; }
ok()    { printf "${GREEN}%s${NC}\n" "[OK]   $*"; }
warn()  { printf "${YELLOW}%s${NC}\n" "[WARN] $*"; }
err()   { printf "${RED}%s${NC}\n" "[ERR]  $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
VERSION=""

while getopts "b:v:h" opt; do
  case "$opt" in
    b) INSTALL_DIR="$OPTARG" ;;
    v) VERSION="$OPTARG" ;;
    h)
      echo "Usage: $0 [-b install_dir] [-v version]"
      echo "  -b <path>     Installation directory (default: $DEFAULT_INSTALL_DIR)"
      echo "  -v <version>  Version to install (default: latest)"
      exit 0
      ;;
    *) exit 1 ;;
  esac
done

# ---------------------------------------------------------------------------
# Detect OS and architecture
# ---------------------------------------------------------------------------
detect_os() {
  local os
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    linux*)  echo "linux" ;;
    darwin*) echo "darwin" ;;
    *)       err "Unsupported OS: $os" ;;
  esac
}

detect_arch() {
  local arch
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)            err "Unsupported architecture: $arch" ;;
  esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

info "Detected OS: ${OS}, Architecture: ${ARCH}"

# ---------------------------------------------------------------------------
# Determine version
# ---------------------------------------------------------------------------
if [ -z "$VERSION" ]; then
  info "Fetching latest release version..."
  VERSION=$(curl -fsSL "${GITHUB_API}/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    err "Failed to fetch latest version. Check network or rate limits."
  fi
  info "Latest release: ${VERSION}"
else
  # Ensure version has 'v' prefix for tag matching
  case "$VERSION" in
    v*) ;;
    *) VERSION="v${VERSION}" ;;
  esac
  info "Installing version: ${VERSION}"
fi

# ---------------------------------------------------------------------------
# Build binary name and download URLs
# ---------------------------------------------------------------------------
BINARY_NAME="${PROJECT_NAME}_${OS}_${ARCH}"
CHECKSUM_FILE="checksums.txt"
BINARY_URL="${GITHUB_DL}/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${BINARY_NAME}"
CHECKSUM_URL="${GITHUB_DL}/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${CHECKSUM_FILE}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

cd "$TMP_DIR"

# ---------------------------------------------------------------------------
# Download binary
# ---------------------------------------------------------------------------
info "Downloading ${BINARY_NAME}..."
if ! curl -fsSL -o "${BINARY_NAME}" "${BINARY_URL}"; then
  err "Failed to download binary from ${BINARY_URL}"
fi
ok "Binary downloaded successfully"

# ---------------------------------------------------------------------------
# Download checksums
# ---------------------------------------------------------------------------
info "Downloading checksums..."
if ! curl -fsSL -o "${CHECKSUM_FILE}" "${CHECKSUM_URL}"; then
  err "Failed to download checksums from ${CHECKSUM_URL}"
fi
ok "Checksums downloaded successfully"

# ---------------------------------------------------------------------------
# Verify SHA256 checksum
# ---------------------------------------------------------------------------
info "Verifying SHA256 checksum..."
if command -v sha256sum &>/dev/null; then
  EXPECTED_HASH=$(grep "${BINARY_NAME}" "${CHECKSUM_FILE}" | awk '{print $1}')
  COMPUTED_HASH=$(sha256sum "${BINARY_NAME}" | awk '{print $1}')
elif command -v shasum &>/dev/null; then
  EXPECTED_HASH=$(grep "${BINARY_NAME}" "${CHECKSUM_FILE}" | awk '{print $1}')
  COMPUTED_HASH=$(shasum -a 256 "${BINARY_NAME}" | awk '{print $1}')
else
  err "No SHA-256 utility found. Install coreutils or shasum."
fi

if [ -z "$EXPECTED_HASH" ]; then
  err "Binary name '${BINARY_NAME}' not found in checksums.txt. Ensure version matches."
fi

if [ "$EXPECTED_HASH" != "$COMPUTED_HASH" ]; then
  err "Checksum mismatch! Expected: ${EXPECTED_HASH}, Computed: ${COMPUTED_HASH}"
fi
ok "SHA256 checksum verified successfully"

# ---------------------------------------------------------------------------
# Install binary
# ---------------------------------------------------------------------------
chmod +x "${BINARY_NAME}"

if [ ! -d "$INSTALL_DIR" ]; then
  info "Creating installation directory: ${INSTALL_DIR}"
  mkdir -p "$INSTALL_DIR"
fi

if [ -f "${INSTALL_DIR}/${PROJECT_NAME}" ]; then
  warn "Overwriting existing binary at ${INSTALL_DIR}/${PROJECT_NAME}"
fi

if ! mv "${BINARY_NAME}" "${INSTALL_DIR}/${PROJECT_NAME}"; then
  err "Failed to install binary to ${INSTALL_DIR}/${PROJECT_NAME}. Try with sudo or use -b with a writable directory."
fi

ok "DeepScanBot ${VERSION} installed successfully to ${INSTALL_DIR}/${PROJECT_NAME}"

# ---------------------------------------------------------------------------
# Verify installation
# ---------------------------------------------------------------------------
if "${INSTALL_DIR}/${PROJECT_NAME}" version &>/dev/null; then
  ok "Installation verified: $("${INSTALL_DIR}/${PROJECT_NAME}" version)"
else
  warn "Binary installed but verification failed. Check PATH and permissions."
fi

# ---------------------------------------------------------------------------
# PATH reminder
# ---------------------------------------------------------------------------
case ":${PATH}:" in
  *:"${INSTALL_DIR}":*) ;;
  *)
    warn "${INSTALL_DIR} is not in your PATH."
    warn "Add it by running: export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac

info "Installation complete! Run 'deepscanbot --help' to get started."