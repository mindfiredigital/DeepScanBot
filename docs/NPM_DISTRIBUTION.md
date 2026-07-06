# npm Distribution System

This document explains how DeepScanBot is distributed via npm and how the packaging system works.

## Overview

DeepScanBot is distributed as an npm package (`@mindfiredigital/deep-scan-bot`) containing pre-built Go binaries for all supported platforms. This approach provides the best of both worlds:

- **Performance**: Native Go binary (no runtime required)
- **Convenience**: Standard npm package management
- **Cross-platform**: Automatic platform detection and binary selection

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      npm Install Flow                        │
└─────────────────────────────────────────────────────────────┘

User runs: npm install -g @mindfiredigital/deep-scan-bot
                            │
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  npm downloads the package (includes all platform binaries) │
└─────────────────────────────────────────────────────────────┘
                            │
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  postinstall.js executes                                    │
│  1. Detects OS (darwin/linux/windows)                       │
│  2. Detects architecture (amd64/arm64)                      │
│  3. Scans dist/ for matching binary directory               │
│  4. Copies binary to bin/                                   │
│  5. Sets executable permissions                             │
│  6. Verifies binary works                                   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Binary installed to npm global bin directory               │
│  User can run: deepscanbot --version                        │
└─────────────────────────────────────────────────────────────┘
```

## Package Structure

The npm package contains:

```
@mindfiredigital/deep-scan-bot/
├── dist/                              # Pre-built binaries
│   ├── deepscanbot_linux_amd64_v1/
│   │   └── deepscanbot               # Linux AMD64 binary
│   ├── deepscanbot_linux_arm64_v8.0/
│   │   └── deepscanbot               # Linux ARM64 binary
│   ├── deepscanbot_darwin_amd64_v1/
│   │   └── deepscanbot               # macOS Intel binary
│   ├── deepscanbot_darwin_arm64_v8.0/
│   │   └── deepscanbot               # macOS Apple Silicon binary
│   ├── deepscanbot_windows_amd64_v1/
│   │   └── deepscanbot.exe           # Windows AMD64 binary
│   └── checksums.txt                  # SHA256 checksums
├── bin/
│   └── deepscanbot                    # Installed binary (created by postinstall)
├── postinstall.js                     # Platform detection & binary installation
├── package.json                       # npm package metadata
└── README.md                          # Documentation
```

## Build Process

### 1. GoReleaser Build

GoReleaser builds binaries for all supported platforms:

```bash
goreleaser release --snapshot --clean
```

Output structure in `dist/`:

```
dist/
├── deepscanbot_linux_amd64_v1/
│   └── deepscanbot
├── deepscanbot_linux_arm64_v8.0/
│   └── deepscanbot
├── deepscanbot_darwin_amd64_v1/
│   └── deepscanbot
├── deepscanbot_darwin_arm64_v8.0/
│   └── deepscanbot
├── deepscanbot_windows_amd64_v1/
│   └── deepscanbot.exe
└── checksums.txt
```

**Key points:**
- Binary name: `deepscanbot` (or `deepscanbot.exe` on Windows)
- Directory naming: `{project}_{os}_{arch}_{goarm}` (e.g., `deepscanbot_linux_amd64_v1`)
- Version suffix varies by Go version (_v1, _v8.0, etc.) - **must not be hardcoded**

### 2. Copy to npm Package

The `scripts/copy-to-npm.sh` script copies binaries to the npm package structure:

```bash
bash scripts/copy-to-npm.sh
```

This script:
- Scans `dist/` for platform directories
- Verifies each platform has a binary
- Copies directories to `dist/` (npm package structure)
- Sets executable permissions (755)
- Generates verification report

### 3. Version Synchronization

The `scripts/sync-version.js` script synchronizes versions:

```bash
VERSION=1.0.0 node scripts/sync-version.js
```

This updates `package.json` version to match the Git tag.

### 4. Generate Checksums

The `scripts/generate-checksums.sh` script generates SHA256 checksums:

```bash
bash scripts/generate-checksums.sh
```

Creates `dist/checksums.txt` with checksums for all binaries.

## Post-Installation Process

The `postinstall.js` script runs automatically after `npm install`:

### Platform Detection

```javascript
// Maps Node.js platform to Go OS naming
function detectOS() {
  const mapping = {
    darwin: "darwin",    // macOS
    linux: "linux",      // Linux
    win32: "windows"     // Windows
  };
  return mapping[process.platform];
}

// Maps Node.js arch to Go arch naming
function detectArch() {
  const mapping = {
    x64: "amd64",        // 64-bit x86
    arm64: "arm64"       // 64-bit ARM
  };
  return mapping[process.arch];
}
```

### Binary Discovery

Scans `dist/` for the matching platform directory:

```javascript
// Finds directory like: deepscanbot_linux_amd64_v1
function findBinaryDirectory(os, arch) {
  const entries = fs.readdirSync(DIST_DIR, { withFileTypes: true });
  
  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    
    const dirName = entry.name.toLowerCase();
    if (dirName.includes("deepscanbot") &&
        dirName.includes(os) &&
        dirName.includes(arch)) {
      return path.join(DIST_DIR, entry.name);
    }
  }
  return null;
}
```

**Key feature**: The version suffix (_v1, _v8.0) is **not hardcoded**. The script scans for directories containing the project name, OS, and architecture.

### Binary Installation

```javascript
// 1. Copy binary to bin/
fs.copyFileSync(srcPath, targetPath);

// 2. Set executable permissions (Unix only)
if (os !== "windows") {
  fs.chmodSync(targetPath, 0o755);
}

// 3. Verify binary works
verifyBinary(targetPath, os);
```

### Binary Verification

The script verifies the binary by running:

```javascript
// Try --version first
"${binaryPath}" --version

// Fallback to --help
"${binaryPath}" --help
```

If verification fails, a warning is displayed but installation continues.

## Supported Platforms

| OS       | Architecture | GoReleaser Directory              | Binary Name      |
|----------|-------------|-----------------------------------|------------------|
| Linux    | amd64       | deepscanbot_linux_amd64_v1        | deepscanbot      |
| Linux    | arm64       | deepscanbot_linux_arm64_v8.0      | deepscanbot      |
| macOS    | amd64       | deepscanbot_darwin_amd64_v1       | deepscanbot      |
| macOS    | arm64       | deepscanbot_darwin_arm64_v8.0     | deepscanbot      |
| Windows  | amd64       | deepscanbot_windows_amd64_v1      | deepscanbot.exe  |

## Local Development Workflow

### Complete Local Testing

```bash
# 1. Build binaries for all platforms
goreleaser build --snapshot --clean

# 2. Verify binaries
node scripts/verify-binary.js

# 3. Create npm package
npm pack

# 4. Install locally
npm install -g ./deep-scan-bot-*.tgz

# 5. Test installation
deepscanbot --version
deepscanbot -h

# 6. Test crawling
deepscanbot scan https://example.com depth=1

# 7. Uninstall after testing
npm uninstall -g @mindfiredigital/deep-scan-bot
```

### Quick Local Build

```bash
# Build just for current platform
go build -o deepscanbot ./apps/cli

# Test directly
./deepscanbot --version
```

## Publishing Workflow

### Automated Release (GitHub Actions)

The release workflow is triggered by pushing a Git tag:

```bash
# 1. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 2. GitHub Actions automatically:
#    - Builds binaries via GoReleaser
#    - Generates checksums
#    - Copies binaries to npm package
#    - Updates package.json version
#    - Publishes to npm registry
#    - Creates GitHub Release
```

### Manual Publishing

```bash
# 1. Build release binaries
goreleaser release --clean

# 2. Copy to npm package
bash scripts/copy-to-npm.sh

# 3. Update version
VERSION=1.0.0 node scripts/sync-version.js

# 4. Generate checksums
bash scripts/generate-checksums.sh

# 5. Verify package
node scripts/prepublish-check.js

# 6. Publish to npm
npm publish
```

## Version Synchronization

All components use the same version from the Git tag:

```
Git Tag: v1.0.0
    ↓
GoReleaser: v1.0.0
    ↓
package.json: 1.0.0
    ↓
GitHub Release: v1.0.0
    ↓
npm package: 1.0.0
```

The `scripts/sync-version.js` script ensures `package.json` version matches the Git tag.

## Binary Verification

### During Build (verify-binary.js)

Verifies all platform binaries:

```bash
node scripts/verify-binary.js
```

Checks:
- Directory exists for each platform
- Binary file exists and is not empty
- File has correct permissions
- Binary format is valid (ELF/Mach-O/PE)
- SHA256 checksum matches (if checksums.txt exists)

### During Installation (postinstall.js)

Verifies the installed binary:

```bash
# Tries --version
deepscanbot --version

# Falls back to --help
deepscanbot --help
```

## Troubleshooting

### "Unsupported platform" Error

**Cause**: Binary not available for your OS/architecture combination.

**Solution**: Check supported platforms in the package. Currently supported:
- macOS (Intel x64, Apple Silicon arm64)
- Linux (amd64, arm64)
- Windows (amd64)

### Binary Verification Fails

**Cause**: Binary may be corrupted or incompatible.

**Solution**:
```bash
# Reinstall the package
npm uninstall -g @mindfiredigital/deep-scan-bot
npm install -g @mindfiredigital/deep-scan-bot

# Check binary integrity
node scripts/verify-binary.js
```

### "Command not found" After Installation

**Cause**: npm global bin directory not in PATH.

**Solution**:
```bash
# Find npm global bin
npm bin -g

# Add to PATH (bash/zsh)
export PATH=$(npm bin -g):$PATH

# Add to PATH (fish)
set -U fish_user_paths (npm bin -g) $fish_user_paths
```

### Permission Denied

**Cause**: Binary doesn't have executable permissions.

**Solution**:
```bash
# Find binary location
which deepscanbot

# Set executable permissions
chmod +x $(which deepscanbot)
```

## Publishing to Different Registries

### npmjs (Default)

```bash
npm publish
```

### GitHub Packages

```bash
NPM_REGISTRY=https://npm.pkg.github.com/ npm publish
```

### Verdaccio (Local Testing)

```bash
NPM_REGISTRY=http://localhost:4873/ npm publish
```

### Custom Registry

```bash
NPM_REGISTRY=https://your-registry.com/ npm publish
```

## Security Considerations

### Binary Integrity

- All binaries are checksummed with SHA256
- Checksums are included in the package and GitHub Release
- Users can verify: `sha256sum deepscanbot`

### Supply Chain Security

- Binaries are built in CI (GitHub Actions)
- Build process is transparent and reproducible
- GoReleaser ensures deterministic builds
- Source code is available for audit

### npm Package Security

- Package name is scoped: `@mindfiredigital/deep-scan-bot`
- Prevents package squatting
- Requires npm account ownership verification

## Performance Characteristics

### Package Size

- Total package size: ~15-20 MB (all platforms)
- Individual platform install: ~5-7 MB
- Installation time: 2-5 seconds (depends on network)

### Runtime Performance

- Native Go binary performance
- No runtime overhead
- Single static binary
- Fast startup time (<100ms)

## Comparison with Other Tools

DeepScanBot uses the same distribution model as:

- **Prisma**: `npm install @prisma/cli`
- **Supabase CLI**: `npm install @supabase/cli`
- **Vercel CLI**: `npm install vercel`
- **Turbo**: `npm install turbo`
- **AWS CDK**: `npm install aws-cdk`
- **Wrangler**: `npm install wrangler`

All these tools distribute native binaries via npm for easy installation and management.

## Future Improvements

Potential enhancements:

1. **Delta Updates**: Only download binary for current platform
2. **Lazy Download**: Download binary on first run instead of install
3. **Binary Caching**: Cache binaries across npm versions
4. **Progressive Download**: Download platforms as needed
5. **Signature Verification**: GPG signature verification for binaries

## References

- [npm Package Documentation](https://docs.npmjs.com/)
- [GoReleaser Documentation](https://goreleaser.com/)
- [npm postinstall Scripts](https://docs.npmjs.com/cli/v10/using-npm/scripts#postinstall)
- [Node.js process.platform](https://nodejs.org/api/os.html#osplatform)
- [Node.js process.arch](https://nodejs.org/api/os.html#osarch)