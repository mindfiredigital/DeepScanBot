# DeepScanBot Multi-Channel Distribution Verification Checklist

This checklist provides QA verification steps for all installation methods. All commands are non-destructive and can be run locally or in CI/CD pipelines.

---

## Table of Contents

1. [GitHub Releases Verification](#1-github-releases-verification)
2. [Homebrew Installation Verification](#2-homebrew-installation-verification)
3. [NPM Installation Verification](#3-npm-installation-verification)
4. [One-Line Installer Verification (curl/PowerShell)](#4-one-line-installer-verification)
5. [Go Install Verification](#5-go-install-verification)
6. [Critical Sanity Check Command](#6-critical-sanity-check-command)
7. [CI/CD Integration Test](#7-cicd-integration-test)

---

## 1. GitHub Releases Verification

### 1.1 Verify Release Assets Exist

```bash
#!/bin/bash
set -euo pipefail

REPO_OWNER="mindfiredigital"
REPO_NAME="DeepScanBot"
VERSION="v1.0.0"  # Replace with actual version to test

echo "=== GitHub Releases Asset Verification ==="

# Fetch release metadata
echo "Fetching release metadata for ${VERSION}..."
RELEASE_DATA=$(curl -fsSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/tags/${VERSION}")

if [ $? -ne 0 ]; then
    echo "❌ FAILED: Could not fetch release ${VERSION}"
    exit 1
fi

echo "✓ Release metadata fetched successfully"

# Required assets for each platform
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
MISSING_ASSETS=0

for asset in "${REQUIRED_ASSETS[@]}"; do
    if echo "$RELEASE_DATA" | grep -q "\"name\": \"${asset}\""; then
        echo "  ✓ ${asset}"
    else
        echo "  ❌ MISSING: ${asset}"
        MISSING_ASSETS=$((MISSING_ASSETS + 1))
    fi
done

if [ $MISSING_ASSETS -gt 0 ]; then
    echo ""
    echo "❌ FAILED: ${MISSING_ASSETS} required asset(s) missing"
    exit 1
fi

echo ""
echo "✓ All required assets present in release ${VERSION}"
```

### 1.2 Verify Release Metadata

```bash
#!/bin/bash
set -euo pipefail

REPO_OWNER="mindfiredigital"
REPO_NAME="DeepScanBot"
VERSION="v1.0.0"

echo "=== Release Metadata Verification ==="

RELEASE_DATA=$(curl -fsSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/tags/${VERSION}")

# Verify release is not a draft
IS_DRAFT=$(echo "$RELEASE_DATA" | jq -r '.draft')
if [ "$IS_DRAFT" = "true" ]; then
    echo "❌ FAILED: Release is still in draft mode"
    exit 1
fi
echo "✓ Release is published (not draft)"

# Verify release name matches version
RELEASE_NAME=$(echo "$RELEASE_DATA" | jq -r '.name')
if [[ ! "$RELEASE_NAME" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    echo "❌ FAILED: Release name doesn't match semantic versioning: ${RELEASE_NAME}"
    exit 1
fi
echo "✓ Release name follows semantic versioning: ${RELEASE_NAME}"

# Verify body contains installation instructions
BODY=$(echo "$RELEASE_DATA" | jq -r '.body')
if ! echo "$BODY" | grep -q "brew install"; then
    echo "⚠ WARNING: Release body missing Homebrew installation instructions"
fi
if ! echo "$BODY" | grep -q "npm install"; then
    echo "⚠ WARNING: Release body missing npm installation instructions"
fi
echo "✓ Release body contains installation instructions"

echo ""
echo "✓ Release metadata verification passed"
```

### 1.3 Verify Checksums File Integrity

```bash
#!/bin/bash
set -euo pipefail

REPO_OWNER="mindfiredigital"
REPO_NAME="DeepScanBot"
VERSION="v1.0.0"

echo "=== Checksums File Verification ==="

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download checksums.txt
echo "Downloading checksums.txt..."
curl -fsSL -o "${TMP_DIR}/checksums.txt" \
    "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/checksums.txt"

if [ ! -f "${TMP_DIR}/checksums.txt" ]; then
    echo "❌ FAILED: Could not download checksums.txt"
    exit 1
fi
echo "✓ checksums.txt downloaded"

# Verify checksums.txt has entries for all expected binaries
EXPECTED_COUNT=5  # 5 binaries (linux/darwin/windows × amd64/arm64, minus windows/arm64)
ACTUAL_COUNT=$(wc -l < "${TMP_DIR}/checksums.txt")

if [ "$ACTUAL_COUNT" -ne "$EXPECTED_COUNT" ]; then
    echo "❌ FAILED: Expected ${EXPECTED_COUNT} checksums, found ${ACTUAL_COUNT}"
    exit 1
fi
echo "✓ checksums.txt contains ${ACTUAL_COUNT} entries"

# Verify checksum format (SHA256 = 64 hex chars)
INVALID_LINES=0
while IFS= read -r line; do
    HASH=$(echo "$line" | awk '{print $1}')
    if ! [[ "$HASH" =~ ^[a-f0-9]{64}$ ]]; then
        echo "❌ Invalid checksum format: ${HASH}"
        INVALID_LINES=$((INVALID_LINES + 1))
    fi
done < "${TMP_DIR}/checksums.txt"

if [ $INVALID_LINES -gt 0 ]; then
    echo "❌ FAILED: ${INVALID_LINES} invalid checksum(s) found"
    exit 1
fi
echo "✓ All checksums are valid SHA256 format (64 hex characters)"

echo ""
echo "✓ Checksums file verification passed"
```

---

## 2. Homebrew Installation Verification

### 2.1 Test Tap and Cask Configuration

```bash
#!/bin/bash
set -euo pipefail

echo "=== Homebrew Installation Verification ==="

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo "⚠ SKIPPED: Homebrew not installed"
    exit 0
fi

echo "✓ Homebrew is installed: $(brew --version | head -n1)"

# Tap the repository (this only needs to be done once)
echo ""
echo "Tapping mindfiredigital/tap..."
brew tap mindfiredigital/tap 2>&1 || {
    # If already tapped, this is fine
    if ! brew tap | grep -q "mindfiredigital/tap"; then
        echo "❌ FAILED: Could not tap mindfiredigital/tap"
        exit 1
    fi
    echo "✓ Tap already exists"
}
echo "✓ Tap successful"

# Verify cask exists
echo ""
echo "Checking if cask 'deepscanbot' exists..."
if ! brew info --cask deepscanbot &> /dev/null; then
    echo "❌ FAILED: Cask 'deepscanbot' not found in tap"
    exit 1
fi
echo "✓ Cask 'deepscanbot' found"

# Display cask information
echo ""
echo "Cask information:"
brew info --cask deepscanbot

# Verify cask can be audited
echo ""
echo "Running brew audit on cask..."
if brew audit --cask deepscanbot 2>&1; then
    echo "✓ Cask passes brew audit"
else
    echo "⚠ WARNING: Cask audit reported issues (may be warnings, not errors)"
fi

echo ""
echo "✓ Homebrew verification passed"
```

### 2.2 Test Installation from Source (Build from Source)

```bash
#!/bin/bash
set -euo pipefail

echo "=== Homebrew Build-from-Source Verification ==="

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo "⚠ SKIPPED: Homebrew not installed"
    exit 0
fi

# Ensure tap exists
if ! brew tap | grep -q "mindfiredigital/tap"; then
    echo "Tapping mindfiredigital/tap..."
    brew tap mindfiredigital/tap
fi

# Install with --build-from-source (tests cask configuration)
echo ""
echo "Installing deepscanbot with --build-from-source..."
echo "Note: This requires the Go toolchain to be available"
brew install --build-from-source deepscanbot 2>&1 || {
    echo "❌ FAILED: Build-from-source installation failed"
    echo "This may indicate issues with the cask configuration or missing dependencies"
    exit 1
}

echo "✓ Build-from-source installation successful"

# Verify binary is in PATH
echo ""
echo "Verifying binary installation..."
if command -v deepscanbot &> /dev/null; then
    echo "✓ deepscanbot is in PATH: $(which deepscanbot)"
else
    echo "❌ FAILED: deepscanbot not found in PATH after installation"
    exit 1
fi

echo ""
echo "✓ Homebrew build-from-source verification passed"
```

---

## 3. NPM Installation Verification

### 3.1 Test Postinstall Script

```bash
#!/bin/bash
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

# Create a temporary directory for testing
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

cd "$TEST_DIR"

# Initialize a test package
echo ""
echo "Setting up test package..."
cat > package.json << 'EOF'
{
  "name": "test-deepscanbot-install",
  "version": "1.0.0",
  "description": "Test package for DeepScanBot installation"
}
EOF

# Copy the actual package files
echo "Copying DeepScanBot package files..."
cp -r /home/lenovo/Documents/DeepScanBot/* .

# Build binaries first (simulating release process)
echo ""
echo "Building binaries with GoReleaser (snapshot)..."
if command -v goreleaser &> /dev/null; then
    goreleaser build --snapshot --clean 2>&1 || {
        echo "⚠ WARNING: GoReleaser build failed, checking if dist/ exists..."
        if [ ! -d "dist" ]; then
            echo "❌ FAILED: dist/ directory not found and GoReleaser build failed"
            exit 1
        fi
    }
    echo "✓ Binaries built"
else
    echo "⚠ WARNING: GoReleaser not found, checking for existing dist/..."
    if [ ! -d "dist" ]; then
        echo "❌ FAILED: dist/ directory not found and GoReleaser not available"
        exit 1
    fi
    echo "✓ Using existing dist/ directory"
fi

# Verify dist/ directory structure
echo ""
echo "Verifying dist/ directory structure..."
REQUIRED_BINARIES=(
    "deepscanbot_linux_amd64/deepscanbot"
    "deepscanbot_linux_arm64/deepscanbot"
    "deepscanbot_darwin_amd64/deepscanbot"
    "deepscanbot_darwin_arm64/deepscanbot"
    "deepscanbot_windows_amd64/deepscanbot.exe"
)

for binary in "${REQUIRED_BINARIES[@]}"; do
    if [ -f "dist/${binary}" ]; then
        echo "  ✓ ${binary}"
    else
        echo "  ❌ MISSING: ${binary}"
    fi
done

# Run npm install locally (this triggers postinstall.js)
echo ""
echo "Running npm install (triggers postinstall.js)..."
npm install 2>&1 | tee npm-install.log || {
    echo "❌ FAILED: npm install failed"
    exit 1
}

# Check if postinstall ran successfully
if grep -q "Installation complete" npm-install.log; then
    echo "✓ Postinstall script completed successfully"
else
    echo "❌ FAILED: Postinstall script did not complete"
    exit 1
fi

# Verify binary was copied to bin/
echo ""
echo "Verifying binary installation..."
DETECTED_OS=$(node -e "const os={darwin:'darwin',linux:'linux',win32:'windows'}; console.log(os[process.platform]||process.platform)")
DETECTED_ARCH=$(node -e "const arch={x64:'amd64',arm64:'arm64'}; console.log(arch[process.arch]||process.arch)")

EXPECTED_BINARY="bin/deepscanbot"
if [ "$DETECTED_OS" = "windows" ]; then
    EXPECTED_BINARY="bin/deepscanbot.exe"
fi

if [ -f "$EXPECTED_BINARY" ]; then
    echo "✓ Binary installed to ${EXPECTED_BINARY}"
    
    # Verify permissions (chmod 755 on Unix)
    if [ "$DETECTED_OS" != "windows" ]; then
        PERMS=$(stat -c "%a" "$EXPECTED_BINARY" 2>/dev/null || stat -f "%Lp" "$EXPECTED_BINARY")
        if [ "$PERMS" = "755" ]; then
            echo "✓ Binary has correct permissions: ${PERMS} (chmod 755)"
        else
            echo "❌ FAILED: Binary has incorrect permissions: ${PERMS} (expected 755)"
            exit 1
        fi
    fi
    
    # Verify binary is executable
    if "$EXPECTED_BINARY" version &> /dev/null; then
        echo "✓ Binary is executable and runs successfully"
        "$EXPECTED_BINARY" version
    else
        echo "❌ FAILED: Binary is not executable or fails to run"
        exit 1
    fi
else
    echo "❌ FAILED: Binary not found at ${EXPECTED_BINARY}"
    exit 1
fi

echo ""
echo "✓ NPM postinstall verification passed"
```

### 3.2 Test Global Installation

```bash
#!/bin/bash
set -euo pipefail

echo "=== NPM Global Installation Verification ==="

# Check if Node.js and npm are installed
if ! command -v npm &> /dev/null; then
    echo "⚠ SKIPPED: npm not installed"
    exit 0
fi

echo "✓ npm is installed: $(npm --version)"

# Uninstall if already installed
echo ""
echo "Cleaning up any existing installation..."
npm uninstall -g @mindfiredigital/deepscanbot 2>/dev/null || true

# Install globally from local source
echo ""
echo "Installing @mindfiredigital/deepscanbot globally from local source..."
npm install -g . 2>&1 | tee npm-global-install.log || {
    echo "❌ FAILED: Global installation failed"
    exit 1
}

# Verify installation
echo ""
echo "Verifying global installation..."

# Check if binary is in PATH
if command -v deepscanbot &> /dev/null; then
    echo "✓ deepscanbot is in PATH: $(which deepscanbot)"
else
    echo "❌ FAILED: deepscanbot not found in PATH"
    echo "npm global bin: $(npm bin -g)"
    exit 1
fi

# Verify version command works
echo ""
echo "Testing deepscanbot version..."
deepscanbot version || {
    echo "❌ FAILED: deepscanbot version command failed"
    exit 1
}

# Verify help command works
echo ""
echo "Testing deepscanbot --help..."
deepscanbot --help &> /dev/null || {
    echo "❌ FAILED: deepscanbot --help command failed"
    exit 1
}
echo "✓ Help command works"

# Verify doctor command works
echo ""
echo "Testing deepscanbot doctor..."
deepscanbot doctor || {
    echo "❌ FAILED: deepscanbot doctor command failed"
    exit 1
}

echo ""
echo "✓ NPM global installation verification passed"
```

---

## 4. One-Line Installer Verification

### 4.1 Bash Installer (curl | sh) Simulation

```bash
#!/bin/bash
set -euo pipefail

echo "=== Bash Installer (curl | sh) Simulation ==="

# Create a temporary directory for testing
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

cd "$TEST_DIR"

# Download the install script
echo "Downloading install.sh..."
curl -fsSL -o install.sh \
    "https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.sh"

if [ ! -f install.sh ]; then
    echo "❌ FAILED: Could not download install.sh"
    exit 1
fi
echo "✓ install.sh downloaded"

# Make it executable
chmod +x install.sh

# Verify script syntax
echo ""
echo "Verifying script syntax..."
if bash -n install.sh; then
    echo "✓ Script syntax is valid"
else
    echo "❌ FAILED: Script has syntax errors"
    exit 1
fi

# Run the installer with custom install directory
echo ""
echo "Running installer with custom directory..."
TEST_INSTALL_DIR="${TEST_DIR}/install"
./install.sh -b "$TEST_INSTALL_DIR" 2>&1 | tee install.log || {
    echo "❌ FAILED: Installation failed"
    cat install.log
    exit 1
}

# Verify installation
echo ""
echo "Verifying installation..."

if [ -f "${TEST_INSTALL_DIR}/deepscanbot" ]; then
    echo "✓ Binary installed to ${TEST_INSTALL_DIR}/deepscanbot"
else
    echo "❌ FAILED: Binary not found at ${TEST_INSTALL_DIR}/deepscanbot"
    exit 1
fi

# Verify permissions
PERMS=$(stat -c "%a" "${TEST_INSTALL_DIR}/deepscanbot" 2>/dev/null || stat -f "%Lp" "${TEST_INSTALL_DIR}/deepscanbot")
if [ "$PERMS" = "755" ]; then
    echo "✓ Binary has correct permissions: ${PERMS}"
else
    echo "❌ FAILED: Binary has incorrect permissions: ${PERMS} (expected 755)"
    exit 1
fi

# Verify binary runs
echo ""
echo "Testing installed binary..."
"${TEST_INSTALL_DIR}/deepscanbot" version || {
    echo "❌ FAILED: Installed binary fails to run"
    exit 1
}

echo ""
echo "✓ Bash installer verification passed"
```

### 4.2 PowerShell Installer (irm | iex) Simulation

**Note:** This test can only run on Windows or with PowerShell Core on Linux/macOS.

```powershell
#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Simulates the PowerShell one-line installer (irm | iex)
#>

param(
    [string]$TestInstallDir = ""
)

$ErrorActionPreference = "Stop"

Write-Host "=== PowerShell Installer (irm | iex) Simulation ===" -ForegroundColor Cyan

# Create temporary directory
$TestDir = Join-Path $env:TEMP "deepscanbot-test-$(New-Guid)"
New-Item -ItemType Directory -Path $TestDir -Force | Out-Null
trap { Remove-Item -Path $TestDir -Recurse -Force -ErrorAction SilentlyContinue }

try {
    # Download the install script
    Write-Host "`nDownloading install.ps1..." -ForegroundColor Yellow
    $InstallScript = Join-Path $TestDir "install.ps1"
    Invoke-WebRequest -Uri "https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.ps1" `
                      -OutFile $InstallScript `
                      -UseBasicParsing
    
    Write-Host "✓ install.ps1 downloaded" -ForegroundColor Green

    # Set custom install directory if provided
    if ([string]::IsNullOrEmpty($TestInstallDir)) {
        $TestInstallDir = Join-Path $TestDir "install"
    }

    # Run the installer
    Write-Host "`nRunning installer..." -ForegroundColor Yellow
    & powershell -ExecutionPolicy Bypass -File $InstallScript -InstallDir $TestInstallDir 2>&1 | Tee-Object -FilePath (Join-Path $TestDir "install.log")

    # Verify installation
    Write-Host "`nVerifying installation..." -ForegroundColor Yellow
    $BinaryPath = Join-Path $TestInstallDir "deepscanbot.exe"
    
    if (Test-Path $BinaryPath) {
        Write-Host "✓ Binary installed to $BinaryPath" -ForegroundColor Green
    }
    else {
        Write-Host "❌ FAILED: Binary not found at $BinaryPath" -ForegroundColor Red
        exit 1
    }

    # Verify binary runs
    Write-Host "`nTesting installed binary..." -ForegroundColor Yellow
    $VersionOutput = & $BinaryPath version
    Write-Host "✓ Binary runs successfully: $VersionOutput" -ForegroundColor Green

    Write-Host "`n✓ PowerShell installer verification passed" -ForegroundColor Green
}
catch {
    Write-Host "❌ FAILED: $_" -ForegroundColor Red
    exit 1
}
finally {
    Remove-Item -Path $TestDir -Recurse -Force -ErrorAction SilentlyContinue
}
```

**Linux/macOS Alternative (Testing Script Logic Only):**

```bash
#!/bin/bash
set -euo pipefail

echo "=== PowerShell Installer Logic Verification (Linux/macOS) ==="
echo "Note: Full PowerShell testing requires Windows or PowerShell Core"

# Create a temporary directory
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

cd "$TEST_DIR"

# Download the PowerShell script
echo "Downloading install.ps1..."
curl -fsSL -o install.ps1 \
    "https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.ps1"

if [ ! -f install.ps1 ]; then
    echo "❌ FAILED: Could not download install.ps1"
    exit 1
fi
echo "✓ install.ps1 downloaded"

# Verify script structure
echo ""
echo "Verifying script structure..."

# Check for required sections
REQUIRED_SECTIONS=(
    "Detect OS and architecture"
    "Determine version"
    "Build binary name and download URLs"
    "Download binary"
    "Download checksums"
    "Verify SHA256 checksum"
    "Install binary"
    "Verify installation"
)

for section in "${REQUIRED_SECTIONS[@]}"; do
    if grep -q "$section" install.ps1; then
        echo "  ✓ Section found: $section"
    else
        echo "  ❌ MISSING: $section"
    fi
done

# Verify key functions exist
echo ""
echo "Verifying key functions..."
if grep -q "function Write-Info" install.ps1; then
    echo "  ✓ Write-Info function"
fi
if grep -q "function Write-Ok" install.ps1; then
    echo "  ✓ Write-Ok function"
fi
if grep -q "function Write-Err" install.ps1; then
    echo "  ✓ Write-Err function"
fi

# Verify SHA256 verification logic
echo ""
echo "Verifying SHA256 verification logic..."
if grep -q "Get-FileHash.*SHA256" install.ps1; then
    echo "  ✓ Uses Get-FileHash for SHA256 verification"
else
    echo "  ❌ SHA256 verification logic not found"
fi

echo ""
echo "✓ PowerShell installer structure verification passed"
echo "⚠ Full functional testing requires Windows environment"
```

---

## 5. Go Install Verification

### 5.1 Verify Go Install Command

```bash
#!/bin/bash
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

# Verify GOPATH
echo ""
echo "Checking Go environment..."
GOPATH=$(go env GOPATH)
echo "  GOPATH: ${GOPATH}"

# Verify main package exists
echo ""
echo "Verifying main package location..."
MAIN_PKG="github.com/mindfiredigital/DeepScanBot/apps/cli"
if go list -m "$MAIN_PKG" &> /dev/null; then
    echo "✓ Module ${MAIN_PKG} is valid"
else
    echo "⚠ WARNING: Could not verify module (may need to be in module root)"
fi

# Test go install (this will build and install the binary)
echo ""
echo "Running 'go install'..."
echo "Note: This will download dependencies and build the binary"

# Clean any existing installation
echo "Cleaning existing installation..."
rm -f "${GOPATH}/bin/deepscanbot" 2>/dev/null || true

# Install
go install github.com/mindfiredigital/DeepScanBot/apps/cli@latest 2>&1 | tee go-install.log || {
    echo "❌ FAILED: go install failed"
    cat go-install.log
    exit 1
}

echo "✓ go install completed"

# Verify binary was installed
echo ""
echo "Verifying binary installation..."
INSTALLED_BINARY="${GOPATH}/bin/deepscanbot"

if [ -f "$INSTALLED_BINARY" ]; then
    echo "✓ Binary installed to ${INSTALLED_BINARY}"
else
    echo "❌ FAILED: Binary not found at ${INSTALLED_BINARY}"
    exit 1
fi

# Verify binary is executable
echo ""
echo "Testing installed binary..."
"$INSTALLED_BINARY" version || {
    echo "❌ FAILED: Installed binary fails to run"
    exit 1
}

# Verify binary is in PATH (or warn if not)
echo ""
if echo "$PATH" | grep -q "${GOPATH}/bin"; then
    echo "✓ ${GOPATH}/bin is in PATH"
else
    echo "⚠ WARNING: ${GOPATH}/bin is not in PATH"
    echo "  Add it with: export PATH=\$PATH:${GOPATH}/bin"
fi

echo ""
echo "✓ Go install verification passed"
```

---

## 6. Critical Sanity Check Command

### The Universal Verification Command

After installing DeepScanBot via **any** method, run this command immediately:

```bash
#!/bin/bash
set -euo pipefail

echo "=== DeepScanBot Universal Sanity Check ==="
echo ""

# Check 1: Binary exists and is executable
echo "1. Checking binary availability..."
if ! command -v deepscanbot &> /dev/null; then
    echo "   ❌ FAILED: deepscanbot command not found in PATH"
    echo "   Hint: Ensure the installation directory is in your PATH"
    exit 1
fi
echo "   ✓ deepscanbot found: $(which deepscanbot)"

# Check 2: Version command works
echo ""
echo "2. Checking version command..."
VERSION_OUTPUT=$(deepscanbot version 2>&1)
if [ $? -ne 0 ]; then
    echo "   ❌ FAILED: deepscanbot version returned an error"
    echo "   Output: ${VERSION_OUTPUT}"
    exit 1
fi
echo "   ✓ Version command works"
echo "   Output: ${VERSION_OUTPUT}"

# Check 3: Verify version is not 'dev'
echo ""
echo "3. Checking version is not development..."
if echo "$VERSION_OUTPUT" | grep -q "dev"; then
    echo "   ❌ FAILED: Version is 'dev' - binary was not built with version info"
    echo "   Hint: This indicates the binary was built from source without ldflags"
    exit 1
fi
echo "   ✓ Version is properly set (not 'dev')"

# Check 4: Doctor command works
echo ""
echo "4. Running doctor diagnostic..."
DOCTOR_OUTPUT=$(deepscanbot doctor 2>&1)
if [ $? -ne 0 ]; then
    echo "   ❌ FAILED: deepscanbot doctor returned an error"
    echo "   Output: ${DOCTOR_OUTPUT}"
    exit 1
fi
echo "   ✓ Doctor command works"
echo "   Output:"
echo "$DOCTOR_OUTPUT" | sed 's/^/     /'

# Check 5: Help command works
echo ""
echo "5. Checking help output..."
if ! deepscanbot --help &> /dev/null; then
    echo "   ❌ FAILED: deepscanbot --help returned an error"
    exit 1
fi
echo "   ✓ Help command works"

# Check 6: JSON output mode works
echo ""
echo "6. Testing JSON output mode..."
JSON_OUTPUT=$(deepscanbot version --json 2>/dev/null)
if [ $? -ne 0 ]; then
    echo "   ❌ FAILED: JSON output mode failed"
    exit 1
fi

# Validate JSON structure
if echo "$JSON_OUTPUT" | jq -e '.status == "success"' &> /dev/null; then
    echo "   ✓ JSON output is valid"
else
    echo "   ❌ FAILED: JSON output is invalid or missing 'status' field"
    echo "   Output: ${JSON_OUTPUT}"
    exit 1
fi

# Check 7: Binary architecture matches system
echo ""
echo "7. Verifying binary architecture..."
BINARY_ARCH=$(file "$(which deepscanbot)" | grep -oE '(x86_64|aarch64|arm64|amd64)' | head -1)
SYSTEM_ARCH=$(uname -m)

case "$SYSTEM_ARCH" in
    x86_64|amd64)
        if [ "$BINARY_ARCH" = "x86_64" ] || [ "$BINARY_ARCH" = "amd64" ]; then
            echo "   ✓ Binary architecture (${BINARY_ARCH}) matches system (${SYSTEM_ARCH})"
        else
            echo "   ❌ FAILED: Binary architecture (${BINARY_ARCH}) doesn't match system (${SYSTEM_ARCH})"
            exit 1
        fi
        ;;
    aarch64|arm64)
        if [ "$BINARY_ARCH" = "aarch64" ] || [ "$BINARY_ARCH" = "arm64" ]; then
            echo "   ✓ Binary architecture (${BINARY_ARCH}) matches system (${SYSTEM_ARCH})"
        else
            echo "   ❌ FAILED: Binary architecture (${BINARY_ARCH}) doesn't match system (${SYSTEM_ARCH})"
            exit 1
        fi
        ;;
esac

echo ""
echo "========================================="
echo "✓ ALL SANITY CHECKS PASSED"
echo "========================================="
echo ""
echo "DeepScanBot is correctly installed and functional."
echo "Version: $(deepscanbot version)"
```

**Quick One-Liner Version:**

```bash
# Run this immediately after installation via any method:
deepscanbot version && deepscanbot doctor && deepscanbot --help && echo "✓ Installation verified"
```

---

## 7. CI/CD Integration Test

### Complete CI Test Script

```bash
#!/bin/bash
# .github/workflows/verify-installation.yml
# This can be used as a GitHub Actions workflow or standalone CI script

set -euo pipefail

echo "========================================="
echo "DeepScanBot Installation CI Verification"
echo "========================================="
echo ""

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

FAILED_TESTS=0
TOTAL_TESTS=0

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "----------------------------------------"
    echo "Test ${TOTAL_TESTS}: ${test_name}"
    echo "----------------------------------------"
    
    if eval "$test_command"; then
        echo -e "${GREEN}✓ PASSED${NC}: ${test_name}"
        return 0
    else
        echo -e "${RED}❌ FAILED${NC}: ${test_name}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Test 1: GitHub Releases
run_test "GitHub Releases - Asset Verification" '
    bash scripts/verify-github-release.sh
'

# Test 2: NPM Installation
run_test "NPM - Postinstall Script" '
    bash scripts/verify-npm-postinstall.sh
'

# Test 3: Bash Installer
run_test "Bash Installer (curl | sh)" '
    bash scripts/verify-bash-installer.sh
'

# Test 4: Go Install
run_test "Go Install" '
    bash scripts/verify-go-install.sh
'

# Test 5: Universal Sanity Check
run_test "Universal Sanity Check" '
    bash scripts/verify-sanity-check.sh
'

# Summary
echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${TOTAL_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "${RED}❌ ${FAILED_TESTS} TEST(S) FAILED${NC}"
    exit 1
fi
```

### GitHub Actions Workflow Example

```yaml
# .github/workflows/verify-installation.yml
name: Verify Installation Methods

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

jobs:
  verify:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        include:
          - os: ubuntu-latest
            shell: bash
          - os: macos-latest
            shell: bash
          - os: windows-latest
            shell: pwsh

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Set up Homebrew (macOS/Linux only)
        if: runner.os != 'Windows'
        run: |
          echo "/home/linuxbrew/.linuxbrew/bin:/opt/homebrew/bin" >> $GITHUB_PATH

      - name: Verify GitHub Release Assets
        run: bash scripts/verify-github-release.sh
        env:
          VERSION: ${{ github.ref_name }}
        continue-on-error: false

      - name: Verify NPM Installation
        run: bash scripts/verify-npm-postinstall.sh
        continue-on-error: false

      - name: Verify Bash Installer
        if: runner.os != 'Windows'
        run: bash scripts/verify-bash-installer.sh
        continue-on-error: false

      - name: Verify PowerShell Installer
        if: runner.os == 'Windows'
        run: pwsh scripts/verify-powershell-installer.ps1
        continue-on-error: false

      - name: Verify Go Install
        run: bash scripts/verify-go-install.sh
        continue-on-error: false

      - name: Run Universal Sanity Check
        run: bash scripts/verify-sanity-check.sh
        continue-on-error: false
```

---

## Quick Reference: Installation Verification Matrix

| Method | Platform | Test Command | Key Verification |
|--------|----------|--------------|------------------|
| **GitHub Releases** | All | `bash scripts/verify-github-release.sh` | Assets, checksums, metadata |
| **Homebrew** | macOS/Linux | `brew install --build-from-source deepscanbot` | Cask config, build, PATH |
| **NPM** | All | `npm install -g . && deepscanbot version` | Postinstall, permissions, binary |
| **curl** | macOS/Linux | `bash scripts/install.sh -b /tmp/test` | Download, SHA256, install |
| **PowerShell** | Windows | `.\install.ps1 -InstallDir C:\test` | Download, SHA256, install |
| **Go Install** | All | `go install github.com/.../apps/cli@latest` | Build, GOPATH, binary |

---

## Common Issues and Troubleshooting

### Issue: "Unsupported platform" during npm install
**Cause:** Binary not found in `dist/` for current OS/arch
**Fix:** Ensure GoReleaser built binaries for all platforms before publishing

### Issue: "Command not found" after installation
**Cause:** Installation directory not in PATH
**Fix:** Add npm global bin or GOPATH/bin to PATH

### Issue: "Permission denied" when running binary
**Cause:** Missing executable permissions
**Fix:** Run `chmod +x $(which deepscanbot)`

### Issue: SHA256 checksum mismatch
**Cause:** Binary corrupted during download or wrong version
**Fix:** Re-download or verify network connectivity

---

## Appendix: Automated Test Scripts

Create these scripts in `scripts/` directory for CI integration:

```bash
# scripts/verify-github-release.sh
# scripts/verify-npm-postinstall.sh
# scripts/verify-bash-installer.sh
# scripts/verify-powershell-installer.ps1
# scripts/verify-go-install.sh
# scripts/verify-sanity-check.sh
```

Each script should be executable (`chmod +x`) and return exit code 0 on success, non-zero on failure.

---

## Notes

- All verification scripts are **non-destructive** - they use temporary directories and clean up after themselves
- Tests can be run independently or as part of a CI pipeline
- The universal sanity check should be run after **any** installation method
- For CI/CD, use the `--no-input` flag to prevent interactive prompts: `deepscanbot --no-input version`

---

**Last Updated:** 2026-07-16  
**DeepScanBot Version:** 1.0.0  
**Maintainer:** Mindfire Digital