# Architecture Overview

This document describes the architecture of the DeepScanBot npm distribution system.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         DeepScanBot Architecture                     │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────────┐         ┌──────────────────┐         ┌──────────────────┐
│   npm Registry   │────────▶│   npm Package    │────────▶│  User's Machine  │
│  (npmjs/GitHub)  │         │  (@mindfiredigital│         │                  │
│                  │         │   /deep-scan-bot) │         │                  │
└──────────────────┘         └──────────────────┘         └──────────────────┘
                                                                    │
                                                                    │ npm install -g
                                                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Installation Process                         │
│                                                                     │
│  1. npm downloads package (~15-20 MB)                              │
│  2. postinstall.js executes                                        │
│  3. Platform detection (OS + arch)                                 │
│  4. Binary selection from dist/                                    │
│  5. Copy to bin/ directory                                          │
│  6. Set executable permissions                                      │
│  7. Verify binary works                                            │
└─────────────────────────────────────────────────────────────────────┘
                                                                    │
                                                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Runtime Execution                              │
│                                                                     │
│  User runs: deepscanbot scan https://example.com                   │
│                            │                                        │
│                            ▼                                        │
│  ┌──────────────────────────────────────────┐                      │
│  │  Native Go Binary (no runtime needed)    │                      │
│  │  • Fast startup (<100ms)                 │                      │
│  │  • Low memory footprint                  │                      │
│  │  • High performance                      │                      │
│  └──────────────────────────────────────────┘                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. Build System

```
┌─────────────────────────────────────────────────────────────┐
│                    Build System                              │
└─────────────────────────────────────────────────────────────┘

Source Code (Go)
    │
    │ go build / goreleaser
    ▼
Platform Binaries
    │
    │ 5 platforms × 1 binary each
    ▼
dist/
    ├── deepscanbot_linux_amd64_v1/deepscanbot
    ├── deepscanbot_linux_arm64_v8.0/deepscanbot
    ├── deepscanbot_darwin_amd64_v1/deepscanbot
    ├── deepscanbot_darwin_arm64_v8.0/deepscanbot
    ├── deepscanbot_windows_amd64_v1/deepscanbot.exe
    └── checksums.txt
```

**Key Components**:
- **GoReleaser**: Cross-compilation for all platforms
- **Build Flags**: `show-source=true -w` (strip debug info), version injection
- **CGO_ENABLED=0**: Static binary, no dependencies
- **Deterministic Builds**: Reproducible binaries

### 2. Package System

```
┌─────────────────────────────────────────────────────────────┐
│                    npm Package Structure                     │
└─────────────────────────────────────────────────────────────┘

@mindfiredigital/deep-scan-bot/
    │
    ├── dist/                    # Pre-built binaries (all platforms)
    │   ├── deepscanbot_linux_amd64_v1/
    │   │   └── deepscanbot      # ~5-7 MB
    │   ├── deepscanbot_linux_arm64_v8.0/
    │   │   └── deepscanbot      # ~5-7 MB
    │   ├── deepscanbot_darwin_amd64_v1/
    │   │   └── deepscanbot      # ~5-7 MB
    │   ├── deepscanbot_darwin_arm64_v8.0/
    │   │   └── deepscanbot      # ~5-7 MB
    │   ├── deepscanbot_windows_amd64_v1/
    │   │   └── deepscanbot.exe  # ~5-7 MB
    │   └── checksums.txt        # SHA256 hashes
    │
    ├── bin/                     # Installed binary (created by postinstall)
    │   └── deepscanbot          # User's platform binary
    │
    ├── postinstall.js           # Platform detection & installation
    ├── package.json             # Package metadata
    ├── README.md                # Documentation
    └── LICENSE.md               # MIT License
```

**Package Size**: ~15-20 MB total (all platforms)

### 3. Installation System

```
┌─────────────────────────────────────────────────────────────┐
│              postinstall.js Architecture                     │
└─────────────────────────────────────────────────────────────┘

Input: npm install -g @mindfiredigital/deep-scan-bot
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Platform Detection                                  │
│   • Read process.platform (darwin/linux/win32)              │
│   • Read process.arch (x64/arm64)                           │
│   • Map to Go OS/arch naming                                │
│   • Validate platform is supported                          │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Binary Discovery                                    │
│   • Scan dist/ directory                                    │
│   • Find directory matching:                                │
│     - Contains "deepscanbot"                                │
│     - Contains OS name                                      │
│     - Contains arch name                                    │
│   • Ignore version suffix (_v1, _v8.0)                      │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Binary Installation                                 │
│   • Copy binary from dist/ to bin/                          │
│   • Set executable permissions (755 on Unix)                │
│   • Use .exe extension on Windows                           │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Binary Verification                                 │
│   • Try: deepscanbot --version                              │
│   • Fallback: deepscanbot --help                            │
│   • Log success or warning                                   │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
Output: deepscanbot command available globally
```

### 4. Release System

```
┌─────────────────────────────────────────────────────────────┐
│                    Release Workflow                          │
└─────────────────────────────────────────────────────────────┘

Git Tag (v1.0.0)
    │
    │ git push origin v1.0.0
    ▼
GitHub Actions (release.yml)
    │
    ├─▶ 1. Checkout & Setup
    │   • Go 1.22.4
    │   • Node.js 20
    │   • GoReleaser
    │
    ├─▶ 2. Build Phase
    │   • goreleaser release --clean
    │   • Builds 5 platform binaries
    │   • Generates checksums.txt
    │
    ├─▶ 3. Package Preparation
    │   • bash scripts/copy-to-npm.sh
    │   • node scripts/sync-version.js
    │   • bash scripts/generate-checksums.sh
    │   • node scripts/verify-binary.js
    │
    ├─▶ 4. Publish Phase
    │   • npm publish --access public
    │   • Creates GitHub Release
    │   • Uploads binary artifacts
    │
    └─▶ 5. Completion
        • npm package available
        • GitHub Release created
        • Users can install
```

## Data Flow

### Installation Data Flow

```
User System
    │
    │ 1. npm install -g @mindfiredigital/deep-scan-bot
    │
    ▼
npm Registry
    │
    │ 2. Download package.tgz
    │
    ▼
Local System
    │
    │ 3. Extract package
    │
    ▼
postinstall.js
    │
    │ 4. Read process.platform → "linux"
    │ 5. Read process.arch → "x64"
    │ 6. Map to: os="linux", arch="amd64"
    │ 7. Scan dist/ for "deepscanbot_linux_amd64"
    │ 8. Found: dist/deepscanbot_linux_amd64_v1/
    │ 9. Copy: dist/.../deepscanbot → bin/deepscanbot
    │ 10. chmod 755 bin/deepscanbot
    │ 11. Verify: deepscanbot --version
    │
    ▼
Global bin directory
    │
    │ 12. Binary installed
    │
    ▼
User can run: deepscanbot --version
```

### Version Synchronization Flow

```
Git Tag: v1.0.0
    │
    │ Pushed to GitHub
    ▼
GitHub Actions
    │
    │ Extracts version: 1.0.0
    ▼
GoReleaser
    │
    │ Uses version: 1.0.0
    ▼
package.json
    │
    │ Updated by sync-version.js: 1.0.0
    ▼
npm Publish
    │
    │ Package version: 1.0.0
    ▼
GitHub Release
    │
    │ Tag: v1.0.0
    ▼
All components synchronized
```

## Platform Support Matrix

| Component | Linux AMD64 | Linux ARM64 | macOS AMD64 | macOS ARM64 | Windows AMD64 |
|-----------|-------------|-------------|-------------|-------------|---------------|
| GoReleaser Build | ✅ | ✅ | ✅ | ✅ | ✅ |
| Binary in Package | ✅ | ✅ | ✅ | ✅ | ✅ |
| postinstall.js Detection | ✅ | ✅ | ✅ | ✅ | ✅ |
| Binary Installation | ✅ | ✅ | ✅ | ✅ | ✅ |
| Binary Execution | ✅ | ✅ | ✅ | ✅ | ✅ |

## Technology Stack

### Build Time
- **Go**: 1.22.4+ (compilation)
- **GoReleaser**: v2 (cross-compilation, packaging)
- **Node.js**: 20+ (npm package management)
- **npm**: 9+ (package publishing)

### Runtime
- **Go Binary**: Static, no dependencies
- **Node.js**: Only for postinstall.js (during npm install)
- **OS**: Linux, macOS, Windows

### CI/CD
- **GitHub Actions**: Automation
- **GoReleaser Action**: Build orchestration
- **npm CLI**: Package publishing

## Security Architecture

### Binary Security
```
┌─────────────────────────────────────────────────────────────┐
│                    Binary Security                           │
└─────────────────────────────────────────────────────────────┘

1. Build Security
   • CI-built in GitHub Actions (trusted environment)
   • Reproducible builds (GoReleaser)
   • No local build artifacts used

2. Integrity Verification
   • SHA256 checksums for all binaries
   • Checksums included in package
   • Checksums in GitHub Release

3. Distribution Security
   • HTTPS only (npm registry)
   • Scoped package name (prevents squatting)
   • npm ownership verification

4. Runtime Security
   • Static binary (no dynamic linking)
   • No runtime dependencies
   • Minimal attack surface
```

### Supply Chain Security

```
┌─────────────────────────────────────────────────────────────┐
│                  Supply Chain Security                       │
└─────────────────────────────────────────────────────────────┘

Source Code (GitHub)
    │
    │ Public, auditable
    ▼
CI Build (GitHub Actions)
    │
    │ • Reproducible
    │ • Transparent
    │ • No secrets in build
    ▼
Binary Artifacts
    │
    │ • Checksummed
    │ • Signed (future)
    ▼
npm Package
    │
    │ • Scoped name
    │ • Verified publisher
    ▼
User Installation
    │
    │ • Verified checksums
    │ • postinstall verification
    ▼
Runtime
    │
    │ • Static binary
    │ • No network calls
    │ • Minimal permissions
```

## Performance Characteristics

### Build Performance
- **GoReleaser Build Time**: ~2-3 minutes (all platforms)
- **Parallel Compilation**: Yes (GoReleaser)
- **Caching**: Go modules cached in CI

### Package Performance
- **Package Size**: ~15-20 MB (all platforms)
- **Download Time**: 2-5 seconds (depends on network)
- **Installation Time**: 2-5 seconds
- **Postinstall Time**: <1 second

### Runtime Performance
- **Binary Size**: ~5-7 MB per platform
- **Startup Time**: <100ms
- **Memory Usage**: ~20-50 MB (depends on crawl)
- **CPU Usage**: Multi-threaded, configurable

## Scalability

### Current Capacity
- **Platforms**: 5 (Linux, macOS, Windows × 2 architectures each)
- **Package Size**: ~15-20 MB
- **Installation Time**: 2-5 seconds

### Scaling Strategies

#### Adding New Platforms
1. Update `.goreleaser.yml` with new GOOS/GOARCH
2. Update `postinstall.js` validation
3. Update `verify-binary.js` expected platforms
4. Update documentation
5. Test on new platform

#### Reducing Package Size
1. **Delta Updates**: Only include current platform
2. **Lazy Download**: Download binary on first run
3. **Compression**: Better compression algorithms
4. **Binary Stripping**: Remove unnecessary symbols

## Error Handling

### Build Errors
- Go compilation errors → CI fails, no package published
- GoReleaser errors → CI fails, no package published
- Missing dependencies → CI fails early

### Installation Errors
- Unsupported platform → Clear error message with supported platforms
- Missing binary → Error with directory listing
- Permission denied → Warning with chmod instructions
- Verification failure → Warning but installation continues

### Runtime Errors
- Invalid arguments → Exit code 1, error message
- Network errors → Retry with backoff
- File system errors → Exit code 1, error message

## Monitoring and Observability

### Build Metrics
- Build success/failure rate
- Build time per platform
- Binary size trends
- Test coverage

### Installation Metrics
- npm download statistics
- Installation success rate
- Platform distribution
- Installation time

### Runtime Metrics
- Command usage (via CLI flags)
- Error rates
- Performance metrics (via JSON output)

## Future Architecture Considerations

### Potential Improvements

1. **Delta Updates**
   - Only download binary for current platform
   - Reduces package size from ~20 MB to ~6 MB
   - Requires npm lazy-install or custom installer

2. **Binary Caching**
   - Cache binaries across npm versions
   - Reduce download time for updates
   - Requires custom npm script logic

3. **Signature Verification**
   - GPG signatures for binaries
   - Verify during postinstall
   - Enhanced security

4. **Auto-Updates**
   - Check for new versions
   - Prompt user to update
   - Similar to npm update -g

5. **Plugin System**
   - Extensible architecture
   - Custom crawlers
   - Output formatters

## References

- [npm Package Architecture](https://docs.npmjs.com/cli/v10/configuring-npm/package-json)
- [GoReleaser Architecture](https://goreleaser.com/architecture/)
- [GitHub Actions Architecture](https://docs.github.com/en/actions/learn-github-actions/understanding-github-actions)
- [Node.js postinstall](https://docs.npmjs.com/cli/v10/using-npm/scripts#postinstall)