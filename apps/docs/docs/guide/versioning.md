---
sidebar_position: 3
---

# Versioning Strategy

DeepScanBot follows Semantic Versioning (SemVer) to ensure predictable version numbers and clear compatibility expectations.

## Semantic Versioning

DeepScanBot uses [Semantic Versioning](https://semver.org/) for all releases. Version numbers follow the format:

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

Examples:
- `1.0.0` - Stable release
- `1.2.3` - Patch release
- `2.0.0` - Major release with breaking changes
- `1.1.0-beta.1` - Pre-release version

## Version Components

### Major Version (MAJOR)

Increment the major version when you make incompatible API/CLI changes:

- **Breaking changes to CLI flags or commands**
- **Changes to command behavior that break existing scripts**
- **Changes to JSON output structure**
- **Removal of deprecated features**
- **Changes to exit codes**

Examples:
- Renaming a flag (e.g., `--depth` → `--max-depth`)
- Removing a command
- Changing the JSON response format
- Changing default behavior in a non-backward-compatible way

### Minor Version (MINOR)

Increment the minor version when you add functionality in a backward-compatible manner:

- **New CLI flags or commands**
- **New features that don't break existing behavior**
- **New output formats or options**
- **Performance improvements**

Examples:
- Adding a new flag (e.g., `--sitemap`)
- Adding a new command (e.g., `config`)
- Adding support for new content types
- Improving error messages

### Patch Version (PATCH)

Increment the patch version for backward-compatible bug fixes:

- **Bug fixes**
- **Security patches**
- **Documentation updates**
- **Performance optimizations that don't change behavior**

Examples:
- Fixing a crash when processing malformed URLs
- Fixing incorrect error messages
- Fixing memory leaks
- Updating dependencies for security

## Pre-Release Versions

Pre-release versions indicate unstable or testing releases:

- **Alpha** (`1.0.0-alpha.1`): Early testing, may have breaking changes
- **Beta** (`1.0.0-beta.1`): Feature-complete, testing phase
- **RC** (`1.0.0-rc.1`): Release candidate, stable but not final

Pre-release versions are sorted by their pre-release identifier and are considered lower precedence than the stable version.

## Build Metadata

Build metadata provides additional information about the build:

- **Git commit hash**: Identifies the exact source code version
- **Build date**: When the binary was built
- **Go version**: Version of Go used to compile
- **Platform**: Target OS/architecture

Build metadata is displayed in the version output and JSON response.

## Version Command

The `version` command displays the current version and build information:

### Human-Readable Output

```bash
$ deepscanbot version
DeepScanBot CLI v1.0.0 (abc1234) built on 2024-01-15T10:30:00Z
```

### JSON Output

```bash
$ deepscanbot version --json
{
  "status": "success",
  "data": {
    "version": "1.0.0",
    "name": "DeepScanBot CLI",
    "git_commit": "abc1234",
    "build_date": "2024-01-15T10:30:00Z",
    "go_version": "go1.21.0",
    "platform": "linux/amd64"
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "command": "version",
    "duration_ms": 0
  }
}
```

## Build-Time Version Embedding

Version information is embedded at build time using Go's `-ldflags` mechanism. This ensures that each binary contains accurate version information without requiring external files.

### Build Command

```bash
go build -ldflags "\
  -X main.cliVersion=1.0.0 \
  -X main.gitCommit=$(git rev-parse --short HEAD) \
  -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o deepscanbot ./apps/cli
```

### Makefile Example

```makefile
VERSION := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -ldflags "\
		-X main.cliVersion=$(VERSION) \
		-X main.gitCommit=$(GIT_COMMIT) \
		-X main.buildDate=$(BUILD_DATE)" \
		-o deepscanbot ./apps/cli
```

### GitHub Actions Example

```yaml
- name: Build
  run: |
    VERSION=${{ github.ref_name }}
    GIT_COMMIT=${{ github.sha }}
    BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    go build -ldflags "\
      -X main.cliVersion=${VERSION} \
      -X main.gitCommit=${GIT_COMMIT} \
      -X main.buildDate=${BUILD_DATE}" \
      -o deepscanbot ./apps/cli
```

## Release Process

### 1. Version Bump

Update the version in the codebase:

```bash
# Update version in main.go
const cliVersion = "1.2.0"
```

### 2. Create Git Tag

Create an annotated tag for the release:

```bash
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0
```

### 3. Build Release Binary

Build the binary with embedded version information:

```bash
VERSION="1.2.0"
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build -ldflags "\
  -X main.cliVersion=${VERSION} \
  -X main.gitCommit=${GIT_COMMIT} \
  -X main.buildDate=${BUILD_DATE}" \
  -o deepscanbot ./apps/cli
```

### 4. Verify Version

Verify the binary has the correct version:

```bash
./deepscanbot version
# Expected: DeepScanBot CLI v1.2.0 (abc1234) built on 2024-01-15T10:30:00Z
```

## Compatibility Guarantees

### Stable Releases (MAJOR.MINOR.PATCH)

- **CLI flags**: Flags are not removed or renamed in patch/minor releases
- **Commands**: Commands are not removed or renamed in patch/minor releases
- **JSON output**: Structure is stable within a major version
- **Exit codes**: Exit codes are consistent within a major version
- **Behavior**: Default behavior doesn't change in patch/minor releases

### Pre-Release Versions

- No compatibility guarantees
- Breaking changes may occur between pre-releases
- Intended for testing and feedback only

## Deprecation Policy

When deprecating features:

1. **Announce deprecation** in release notes
2. **Add deprecation warning** to the feature
3. **Maintain functionality** for at least 2 minor releases
4. **Remove in next major version**

Example deprecation notice:

```go
scanCmd.Flags().String("old-flag", "", 
    "DEPRECATED: Use --new-flag instead. Will be removed in v2.0.")
```

## Version Checking in Scripts

Scripts can check the version to ensure compatibility:

```bash
#!/bin/bash
# Check if DeepScanBot version is compatible

REQUIRED_VERSION="1.2.0"
CURRENT_VERSION=$(deepscanbot version --json | jq -r '.data.version')

# Simple version comparison (for major version checks)
MAJOR=$(echo $REQUIRED_VERSION | cut -d. -f1)
CURRENT_MAJOR=$(echo $CURRENT_VERSION | cut -d. -f1)

if [ "$CURRENT_MAJOR" -lt "$MAJOR" ]; then
    echo "Error: DeepScanBot v${REQUIRED_VERSION} or higher is required"
    echo "Current version: ${CURRENT_VERSION}"
    exit 1
fi

echo "DeepScanBot version ${CURRENT_VERSION} is compatible"
```

## Questions?

If you have questions about versioning or need to propose a versioning change, please open an issue or discussion on GitHub.