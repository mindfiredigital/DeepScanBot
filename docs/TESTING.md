# Testing Guide

This document provides comprehensive testing procedures for the DeepScanBot npm distribution system.

## Prerequisites

Ensure you have the following installed:

```bash
# Check Go version (1.22.4+)
go version

# Check Node.js version (18+)
node --version

# Check npm version
npm --version

# Check GoReleaser (optional, for full testing)
goreleaser --version
```

## Test Suite

### 1. Go Code Tests

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -v -race ./...

# Run tests with coverage
go test -v -race -count=1 -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Expected**: All tests pass with no failures.

### 2. Code Quality Checks

```bash
# Check Go formatting
gofumpt -l .
# Expected: No output (all files formatted)

# Format code if needed
gofumpt -l -w .

# Organize imports
gci write --skip-vendor show-source=true standard show-source=true default show-source=true "prefix(github.com/mindfiredigital/DeepScanBot)" .

# Run linter
golangci-lint run ./...
# Expected: No linting errors

# Run go vet
go vet ./...
# Expected: No issues
```

### 3. GoReleaser Configuration

```bash
# Check GoReleaser configuration
goreleaser check
# Expected: "config is valid"

# Test snapshot build
goreleaser build --snapshot --clean
# Expected: Creates dist/ with 5 platform binaries
```

**Verify output**:
```bash
# List generated binaries
ls -lh dist/

# Expected structure:
# dist/
# ├── deepscanbot_linux_amd64_v1/
# │   └── deepscanbot
# ├── deepscanbot_linux_arm64_v8.0/
# │   └── deepscanbot
# ├── deepscanbot_darwin_amd64_v1/
# │   └── deepscanbot
# ├── deepscanbot_darwin_arm64_v8.0/
# │   └── deepscanbot
# ├── deepscanbot_windows_amd64_v1/
# │   └── deepscanbot.exe
# └── checksums.txt
```

### 4. Binary Verification

```bash
# Verify all platform binaries
node scripts/verify-binary.js
```

**Expected output**:
```
[verify-binary] DeepScanBot CLI Binary Verification
[verify-binary] ====================================
[verify-binary] Directory: /path/to/dist

[verify-binary] Available directories: deepscanbot_linux_amd64_v1, deepscanbot_linux_arm64_v8.0, ...

  ✓ linux/amd64 (X.X MB, ELF)
      Path: dist/deepscanbot_linux_amd64_v1/deepscanbot
  ✓ linux/arm64 (X.X MB, ELF)
  ✓ darwin/amd64 (X.X MB, Mach-O)
  ✓ darwin/arm64 (X.X MB, Mach-O)
  ✓ windows/amd64 (X.X MB, PE)

[verify-binary] Results: 5 passed, 0 failed
[verify-binary] All binaries verified successfully!
```

### 5. npm Package Preparation

```bash
# Copy binaries to npm package structure
bash scripts/copy-to-npm.sh
```

**Expected output**:
```
[copy-to-npm] Starting binary copy process...
[copy-to-npm] Source: /path/to/dist
[copy-to-npm] Target: /path/to/dist

[copy-to-npm] GoReleaser output directories:
[copy-to-npm]   Found directory: deepscanbot_linux_amd64_v1
[copy-to-npm]   ✓ Copied directory: deepscanbot_linux_amd64_v1
...

[copy-to-npm] Verifying platform directories...
[copy-to-npm]   ✓ linux_amd64 -> deepscanbot_linux_amd64_v1/deepscanbot (XXXX bytes)
...

[copy-to-npm] All expected platform directories present.
[copy-to-npm] Copy process complete. 5 directories processed.
```

```bash
# Generate checksums
bash scripts/generate-checksums.sh
```

**Expected output**:
```
[generate-checksums] Generating SHA256 checksums...
[generate-checksums] Directory: /path/to/dist

[generate-checksums]   ✓ deepscanbot_linux_amd64_v1/deepscanbot
[generate-checksums]   ✓ deepscanbot_linux_arm64_v8.0/deepscanbot
...

[generate-checksums] Generated checksums for 5 platform directories.
[generate-checksums] Output: /path/to/dist/checksums.txt
[generate-checksums] Done.
```

```bash
# Verify checksums file
cat dist/checksums.txt
```

**Expected format**:
```
<sha256_hash>  deepscanbot_linux_amd64_v1
<sha256_hash>  deepscanbot_linux_arm64_v8.0
...
```

```bash
# Synchronize version (if needed)
VERSION=1.0.0 node scripts/sync-version.js
```

**Expected output**:
```
[sync-version] Version updated: 0.1.0 -> 1.0.0
```

```bash
# Run pre-publish check
node scripts/prepublish-check.js
```

**Expected output**:
```
[prepublish] Checking required files...
  ✓ package.json
  ✓ postinstall.js
  ✓ README.md

[prepublish] Checking package.json...
  ✓ Version: 1.0.0
  ✓ Bin entries: deepscanbot

[prepublish] Checking dist/ directory...
  ✓ dist/ contains 5 subdirectories, 1 files
  ✓ linux/amd64 (X.X KB)
  ✓ linux/arm64 (X.X KB)
  ✓ darwin/amd64 (X.X KB)
  ✓ darwin/arm64 (X.X KB)
  ✓ windows/amd64 (X.X KB)
  ✓ checksums.txt

[prepublish] Checking postinstall.js...
  ✓ postinstall.js syntax OK

[prepublish] =================================
[prepublish] ✅ All checks passed. Ready to publish.
```

### 6. npm Package Creation

```bash
# Create npm package
npm pack
```

**Expected output**:
```
npm notice
npm notice 📦  @mindfiredigital/deep-scan-bot@1.0.0
npm notice === Tarball Contents ===
npm notice 1.7kB  package.json
npm notice 1.2kB  postinstall.js
npm notice 15.3MB dist/deepscanbot_linux_amd64_v1/deepscanbot
npm notice 14.8MB dist/deepscanbot_linux_arm64_v8.0/deepscanbot
...
npm notice Tarball Details
npm notice name:          @mindfiredigital/deep-scan-bot
npm notice version:       1.0.0
npm notice filename:      mindfiredigital-deep-scan-bot-1.0.0.tgz
npm notice package size:  15.2 MB
npm notice unpacked size: 16.1 MB
npm notice total files:   12
```

**Verify package contents**:
```bash
# List package contents
tar -tzf mindfiredigital-deep-scan-bot-1.0.0.tgz
```

**Expected**: Should include `dist/`, `postinstall.js`, `package.json`, `README.md`, `LICENSE.md`

### 7. Local Installation Test

```bash
# Install package locally
npm install -g ./mindfiredigital-deep-scan-bot-1.0.0.tgz
```

**Expected output**:
```
npm notice 
npm notice 📦  @mindfiredigital/deep-scan-bot@1.0.0
npm notice === Tarball Contents ===
...
npm notice === Tarball Details ===
...
npm notice 
npm notice New major version of package available!
npm notice To update: npm install -g @mindfiredigital/deep-scan-bot@latest
npm notice 
npm WARN deprecated @mindfiredigital/deep-scan-bot@1.0.0: This is a test install
+ @mindfiredigital/deep-scan-bot@1.0.0
added 1 package in 2s

[deepscanbot] Detected platform: linux/amd64
[deepscanbot] dist/ subdirectories: deepscanbot_linux_amd64_v1, ...
[deepscanbot] Found binary directory: dist/deepscanbot_linux_amd64_v1
[deepscanbot] Installing binary: dist/deepscanbot_linux_amd64_v1/deepscanbot -> bin/deepscanbot
[deepscanbot] Applied executable permissions to bin/deepscanbot
[deepscanbot] Binary verification: OK (version: 1.0.0)
[deepscanbot] Installation complete!
[deepscanbot] Run "deepscanbot -h" to get started.
```

### 8. Binary Functionality Test

```bash
# Test version command
deepscanbot --version
```

**Expected output**:
```
1.0.0
```

```bash
# Test help command
deepscanbot -h
```

**Expected output**:
```
Usage of deepscanbot:
  concurrency=int
        Maximum concurrent requests (0 = CPU count) (default 0)
  -content-types string
        MIME types to download (quoted, space/comma separated) (default "text/html")
  ...
```

```bash
# Test actual crawling
deepscanbot scan https://example.com depth=1 -json
```

**Expected**: JSON output with crawl results

```bash
# Verify binary location
which deepscanbot
```

**Expected**: Path to global npm bin directory

```bash
# Verify binary permissions (Unix/Linux/macOS)
ls -l $(which deepscanbot)
```

**Expected**: `-rwxr-xr-x` (executable permissions)

### 9. Uninstall Test

```bash
# Uninstall package
npm uninstall -g @mindfiredigital/deep-scan-bot
```

**Expected output**:
```
removed 1 package in 0.5s
```

```bash
# Verify binary is removed
which deepscanbot
```

**Expected**: "deepscanbot not found" or similar

### 10. Cross-Platform Testing

Test the installation process on different platforms:

#### macOS Intel
```bash
# Should install deepscanbot (not .exe)
file $(which deepscanbot)
# Expected: Mach-O 64-bit executable x86_64
```

#### macOS Apple Silicon
```bash
# Should install deepscanbot (not .exe)
file $(which deepscanbot)
# Expected: Mach-O 64-bit executable arm64
```

#### Linux AMD64
```bash
# Should install deepscanbot
file $(which deepscanbot)
# Expected: ELF 64-bit LSB executable, x86-64
```

#### Linux ARM64
```bash
# Should install deepscanbot
file $(which deepscanbot)
# Expected: ELF 64-bit LSB executable, ARM aarch64
```

#### Windows AMD64
```bash
# Should install deepscanbot.exe
where deepscanbot
# Expected: Path to deepscanbot.exe
```

## Automated Test Script

Create a comprehensive test script:

```bash
#!/usr/bin/env bash
# test-install.sh

set -euo pipefail

echo "=== DeepScanBot npm Distribution Test Suite ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

# Test function
test_step() {
  local name="$1"
  local cmd="$2"

  echo -n "Testing: $name... "

  if eval "$cmd" > /dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    ((PASS++))
  else
    echo -e "${RED}FAIL${NC}"
    ((FAIL++))
  fi
}

# Go tests
echo "--- Go Tests ---"
test_step "Go tests" "go test ./..."
test_step "Go vet" "go vet ./..."
test_step "golangci-lint" "golangci-lint run ./..."

# Build tests
echo ""
echo "--- Build Tests ---"
test_step "GoReleaser check" "goreleaser check"
test_step "GoReleaser snapshot build" "goreleaser build --snapshot --clean"

# Binary verification
echo ""
echo "--- Binary Verification ---"
test_step "Binary verification" "node scripts/verify-binary.js"

# Package preparation
echo ""
echo "--- Package Preparation ---"
test_step "Copy to npm" "bash scripts/copy-to-npm.sh"
test_step "Generate checksums" "bash scripts/generate-checksums.sh"
test_step "Pre-publish check" "node scripts/prepublish-check.js"

# npm packaging
echo ""
echo "--- npm Packaging ---"
test_step "npm pack" "npm pack"

# Installation
echo ""
echo "--- Installation ---"
PACKAGE=$(ls mindfiredigital-deep-scan-bot-*.tgz 2>/dev/null | head -1)
test_step "npm install" "npm install -g ./$PACKAGE"

# Functionality
echo ""
echo "--- Functionality ---"
test_step "Binary exists" "command -v deepscanbot"
test_step "Version command" "deepscanbot --version"
test_step "Help command" "deepscanbot -h"

# Cleanup
echo ""
echo "--- Cleanup ---"
npm uninstall -g @mindfiredigital/deep-scan-bot > /dev/null 2>&1 || true
echo "Uninstalled test package"

# Results
echo ""
echo "=== Test Results ==="
echo -e "Passed: ${GREEN}$PASS${NC}"
echo -e "Failed: ${RED}$FAIL${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
  echo -e "${GREEN}✅ All tests passed!${NC}"
  exit 0
else
  echo -e "${RED}❌ Some tests failed${NC}"
  exit 1
fi
```

**Usage**:
```bash
chmod +x test-install.sh
./test-install.sh
```

## CI/CD Testing

### GitHub Actions CI

The CI workflow (`.github/workflows/ci.yml`) automatically runs:

1. **Build and Test Job**:
   - Go tests with race detection
   - Code coverage
   - Go formatting check
   - golangci-lint
   - go vet

2. **GoReleaser Check Job**:
   - Configuration validation
   - Snapshot build
   - Output verification

3. **Cross-Platform Build Job**:
   - Matrix build for all platforms
   - Verifies cross-compilation

4. **npm Package Validation Job**:
   - package.json validation
   - postinstall.js syntax check
   - Scripts syntax check

### GitHub Actions Release

The release workflow (`.github/workflows/release.yml`) runs on tag push:

1. Builds all platform binaries
2. Copies to npm package
3. Updates version
4. Generates checksums
5. Verifies binaries
6. Publishes to npm
7. Creates GitHub Release

## Performance Testing

### Package Size

```bash
# Check package size
ls -lh mindfiredigital-deep-scan-bot-*.tgz

# Check individual binary sizes
ls -lh dist/*/deepscanbot*
```

**Expected**:
- Total package: ~15-20 MB
- Per binary: ~5-7 MB

### Installation Time

```bash
# Time the installation
time npm install -g ./mindfiredigital-deep-scan-bot-*.tgz
```

**Expected**: 2-5 seconds

### Binary Startup Time

```bash
# Time binary startup
time deepscanbot --version
```

**Expected**: <100ms

## Security Testing

### Checksum Verification

```bash
# Generate checksums
bash scripts/generate-checksums.sh

# Verify a binary manually
sha256sum dist/deepscanbot_linux_amd64_v1/deepscanbot
cat dist/checksums.txt | grep deepscanbot_linux_amd64_v1

# They should match
```

### Binary Format Verification

```bash
# Check binary format (Linux)
file dist/deepscanbot_linux_amd64_v1/deepscanbot
# Expected: ELF 64-bit LSB executable

# Check binary format (macOS)
file dist/deepscanbot_darwin_amd64_v1/deepscanbot
# Expected: Mach-O 64-bit executable

# Check binary format (Windows)
file dist/deepscanbot_windows_amd64_v1/deepscanbot.exe
# Expected: PE32+ executable
```

## Troubleshooting Tests

### Test Unsupported Platform

```bash
# Simulate unsupported platform (if possible)
# postinstall.js should fail with clear error message
```

### Test Corrupted Binary

```bash
# Corrupt a binary
echo "corrupted" > dist/deepscanbot_linux_amd64_v1/deepscanbot

# Try to install
npm install -g ./package.tgz
# Expected: Warning about verification failure
```

### Test Missing Binary

```bash
# Remove a binary
rm dist/deepscanbot_linux_amd64_v1/deepscanbot

# Try to install on that platform
npm install -g ./package.tgz
# Expected: Error about binary not found
```

## Regression Testing

After any changes:

1. Run full test suite: `./test-install.sh`
2. Test on all supported platforms
3. Verify npm package can be installed
4. Verify binary works correctly
5. Check documentation is updated

## Continuous Testing

### Pre-Commit Hooks

If using lefthook or similar:

```bash
# .lefthook.yml
pre-commit:
  commands:
    go-test:
      glob: "*.go"
      run: go test ./...
    go-lint:
      glob: "*.go"
      run: golangci-lint run ./...
    npm-check:
      glob: "*.{js,json}"
      run: node scripts/prepublish-check.js
```

### CI Pipeline

Ensure all CI checks pass:
- [ ] Go tests
- [ ] Linting
- [ ] GoReleaser build
- [ ] npm package validation
- [ ] postinstall.js syntax

## Test Coverage

### Required Coverage

- **Go code**: >80% coverage
- **postinstall.js**: All platform detection paths
- **Scripts**: All success and error paths

### Manual Testing Checklist

- [ ] Install on macOS Intel
- [ ] Install on macOS Apple Silicon
- [ ] Install on Linux AMD64
- [ ] Install on Linux ARM64
- [ ] Install on Windows AMD64
- [ ] Verify binary runs on each platform
- [ ] Verify --version works
- [ ] Verify --help works
- [ ] Verify crawling works
- [ ] Verify uninstall works
- [ ] Verify reinstall works

## Load Testing

### Concurrent Installs

```bash
# Test multiple concurrent installations
for i in {1..10}; do
  npm install -g ./package.tgz &
done
wait
```

### Large Crawl Test

```bash
# Test with larger website
deepscanbot scan https://example.com depth=3 concurrency=10 json=true output=test_results
```

## Documentation Testing

- [ ] README.md installation instructions work
- [ ] All code examples are valid
- [ ] All commands execute successfully
- [ ] Links are valid
- [ ] Documentation builds (if using Docusaurus)

## Final Validation Checklist

Before releasing:

- [ ] All Go tests pass
- [ ] All linting passes
- [ ] GoReleaser build succeeds
- [ ] All 5 platform binaries generated
- [ ] Binary verification passes
- [ ] npm pack succeeds
- [ ] Local installation works
- [ ] Binary executes correctly
- [ ] --version works
- [ ] --help works
- [ ] Actual crawl works
- [ ] Uninstall works
- [ ] Documentation updated
- [ ] No hardcoded version suffixes
- [ ] Package name is correct
- [ ] Version is synchronized
- [ ] Checksums are generated
- [ ] All scripts have no syntax errors

## References

- [npm Testing Guide](https://docs.npmjs.com/cli/v10/using-npm/scripts#test)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [GoReleaser Testing](https://goreleaser.com/guides/testing/)
- [GitHub Actions Testing](https://docs.github.com/en/actions/automating-builds-and-tests)