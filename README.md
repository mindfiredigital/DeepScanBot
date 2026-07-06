# DeepScanBot CLI

[![CI](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/ci.yml/badge.svg)](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/ci.yml)
[![Release](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/release.yml/badge.svg)](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/release.yml)
[![npm version](https://img.shields.io/npm/v/@mindfiredigital/deep-scan-bot)](https://www.npmjs.com/package/@mindfiredigital/deep-scan-bot)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mindfiredigital/DeepScanBot)](https://golang.org/)
[![License](https://img.shields.io/github/license/mindfiredigital/DeepScanBot)](LICENSE.md)

A high-performance web crawler and scanner built in Go, distributed as a pre-built binary via npm.

## Features

- **Web Crawling**: Recursively crawl websites with configurable depth
- **Concurrent Requests**: High-performance concurrent crawling
- **Cross-Platform**: macOS (Intel + Apple Silicon), Linux (amd64 + arm64), Windows (amd64)
- **Multiple Output Formats**: Text and JSON output
- **Proxy Support**: HTTP/HTTPS proxy support
- **Robots.txt**: Optional robots.txt compliance
- **Sitemap Support**: Discover URLs from sitemap.xml
- **Resume Mode**: Resume interrupted crawls
- **Retry Logic**: Configurable retry with backoff
- **Crawl Delay**: Politeness delay between requests

## Installation

### Via npm (Recommended)

```bash
npm install -g @mindfiredigital/deep-scan-bot
```

After installation, the `deepscanbot` command will be available globally:

```bash
deepscanbot --help
```

### Via Go (Development)

```bash
go install github.com/mindfiredigital/DeepScanBot/apps/cli@latest
```

### Via Binary Download

Download the appropriate binary for your platform from the [GitHub Releases](https://github.com/mindfiredigital/DeepScanBot/releases) page.

## Quick Start

```bash
# Crawl a website with default settings
deepscanbot scan https://example.com

# Crawl with custom depth and concurrency
deepscanbot scan https://example.com depth=3 concurrency=10

# Output as JSON
deepscanbot scan https://example.com json=true

# Use a proxy
deepscanbot scan https://example.com proxy=http://127.0.0.1:8080

# Enable sitemap discovery
deepscanbot scan https://example.com sitemap=true

# Resume a previous crawl
deepscanbot scan https://example.com resume=true output=crawler_results

# Show version
deepscanbot version

# Verify installation
deepscanbot doctor

# Generate shell completion
deepscanbot completion bash
```

## Usage

DeepScanBot uses a modern command-based CLI structure similar to git, docker, and kubectl.

### Commands

```bash
deepscanbot <command> [options]
```

| Command      | Description                            |
|--------------|----------------------------------------|
| `scan`       | Crawl and analyze a website            |
| `version`    | Show installed version                 |
| `doctor`     | Verify installation and environment    |
| `config`     | Manage CLI configuration               |
| `completion` | Generate shell completion script       |
| `help`       | Show help for any command              |

### Scan Command

```bash
deepscanbot scan <url> [key=value ...]
```

Options are specified as `key=value` pairs after the URL.

| Option                 | Description                                              | Default             |
|------------------------|----------------------------------------------------------|---------------------|
| `depth`                | Maximum crawl depth                                      | `2`                 |
| `timeout`              | Request timeout in seconds                               | `2`                 |
| `proxy`                | Proxy URL (e.g. `http://127.0.0.1:8080`)                 | `""`                |
| `json`                 | Output as JSON                                           | `false`             |
| `size`                 | Page size limit in KB (-1 = no limit)                    | `-1`                |
| `disable-redirects`    | Disable following redirects                              | `false`             |
| `show-source`          | Show source of each URL                                  | `false`             |
| `insecure`             | Disable TLS verification                                 | `false`             |
| `unique`               | Ensure unique URLs                                       | `false`             |
| `concurrency`          | Maximum concurrent requests (0 = CPU count)              | `0`                 |
| `host-concurrency`     | Max concurrent requests per host (0 = use concurrency)   | `0`                 |
| `content-types`        | MIME types to download (quoted, space/comma separated)   | `"text/html"`       |
| `output`               | Output filename without extension                        | `"crawler_results"` |
| `ignore-robots`        | Ignore robots.txt restrictions                           | `false`             |
| `cross-domain`         | Follow links to other hosts                              | `false`             |
| `retries`              | Number of retry attempts                                 | `0`                 |
| `retry-backoff`        | Base retry backoff duration (e.g. `500ms`, `2s`)         | `1s`                |
| `delay`                | Politeness delay between requests to same host           | `0`                 |
| `sitemap`              | Discover URLs from /sitemap.xml                          | `false`             |
| `resume`               | Load existing output and avoid recrawling                | `false`             |

### Examples

#### Basic Crawl

```bash
deepscanbot scan https://example.com
```

#### Advanced Crawl with All Options

```bash
deepscanbot scan https://docs.example.com \
  depth=5 \
  concurrency=20 \
  host-concurrency=5 \
  timeout=10 \
  delay=200ms \
  retries=3 \
  retry-backoff=1s \
  json=true \
  sitemap=true \
  cross-domain=true \
  content-types="text/html application/pdf" \
  output=scan_results
```

## Development

### Prerequisites

- Go 1.22.4+
- Node.js 18+ (for npm package development)
- GoReleaser (for building releases)

### Local Setup

```bash
# Clone the repository
git clone https://github.com/mindfiredigital/DeepScanBot.git
cd DeepScanBot

# Install Go dependencies
go mod download

# Build the CLI
go build -o deepscanbot ./apps/cli

# Run tests
go test ./...

# Run linting
golangci-lint run ./...
```

### Building Manually

```bash
# Build for current platform
go build -o deepscanbot ./apps/cli

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o deepscanbot-linux-amd64 ./apps/cli
GOOS=darwin GOARCH=amd64 go build -o deepscanbot-darwin-amd64 ./apps/cli
GOOS=windows GOARCH=amd64 go build -o deepscanbot-windows-amd64.exe ./apps/cli
```

### Building with GoReleaser

```bash
# Install GoReleaser
# See: https://goreleaser.com/install/

# Snapshot build (local testing)
goreleaser release --snapshot --clean

# Release build
goreleaser release --clean
```

## Release Flow

The release process is fully automated via GitHub Actions.

```
Git Tag (v1.0.0)
      ↓
GitHub Action (release.yml)
      ↓
GoReleaser Build
      ↓
dist/ (binaries + checksums)
      ↓
Copy to npm package
      ↓
Update package.json version
      ↓
npm publish
      ↓
npm install -g @mindfiredigital/deep-scan-bot
```

### Creating a Release

1. **Create a Git tag**:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

2. **The GitHub Action will automatically**:
   - Build binaries for all platforms via GoReleaser
   - Generate SHA256 checksums
   - Copy binaries into the npm package
   - Update package.json version from the Git tag
   - Publish to the npm registry
   - Create a GitHub Release with artifacts

3. **Users can then install**:

```bash
npm install -g @mindfiredigital/deep-scan-bot
```

### Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.x.x): Breaking changes
- **MINOR** (x.1.x): New features, backward compatible
- **PATCH** (x.x.1): Bug fixes, backward compatible

Pre-release versions use the `-next` tag on npm:

```bash
npm install -g @mindfiredigital/deep-scan-bot@next
```

## Publishing

### Configuration

The npm publish configuration comes from environment variables:

| Variable       | Description              | Default                       |
| -------------- | ------------------------ | ----------------------------- |
| `NPM_REGISTRY` | npm registry URL         | `https://registry.npmjs.org/` |
| `NPM_TOKEN`    | npm authentication token | (required)                    |

### Supported Registries

| Registry            | URL                                                                | Authentication |
| ------------------- | ------------------------------------------------------------------ | -------------- |
| **npmjs**           | `https://registry.npmjs.org/`                                      | `NPM_TOKEN`    |
| **GitHub Packages** | `https://npm.pkg.github.com/`                                      | `GITHUB_TOKEN` |
| **Verdaccio**       | `http://localhost:4873/`                                           | `NPM_TOKEN`    |
| **Nexus**           | `https://nexus.example.com/repository/npm-private/`                | `NPM_TOKEN`    |
| **Artifactory**     | `https://artifactory.example.com/artifactory/api/npm/npm-private/` | `NPM_TOKEN`    |

### Publishing to Different Registries

```bash
# npmjs (default)
npm publish

# GitHub Packages
NPM_REGISTRY=https://npm.pkg.github.com/ npm publish

# Verdaccio (local)
NPM_REGISTRY=http://localhost:4873/ npm publish

# Dry run (verify without publishing)
npm publish --dry-run
```

### Snapshot Releases

For testing purposes, you can create snapshot releases:

```bash
# Create a snapshot build
goreleaser release --snapshot --clean

# Copy binaries and update version
bash scripts/copy-to-npm.sh
VERSION=0.0.0-snapshot node scripts/sync-version.js

# Dry run publish
npm publish --dry-run
```

## Project Structure

```
project/
│
├── apps/
│   └── cli/              # Go CLI application
│       ├── main.go
│       └── tests/
│
├── packages/
│   ├── crawler/          # Web crawling logic
│   ├── fetcher/          # HTTP fetching
│   ├── logger/           # Logging utilities
│   ├── parser/           # HTML parsing
│   ├── storage/          # Output storage
│   └── types/            # Shared types
│
├── dist/                 # Pre-built binaries (generated)
├── bin/                  # Installed binary (generated by postinstall)
├── scripts/              # Helper scripts
│   ├── copy-to-npm.sh
│   ├── sync-version.js
│   ├── verify-binary.js
│   ├── generate-checksums.sh
│   └── prepublish-check.js
│
├── .github/
│   └── workflows/
│       ├── ci.yml        # CI workflow
│       └── release.yml   # Release workflow
│
├── .goreleaser.yml       # GoReleaser configuration
├── package.json          # npm package configuration
├── postinstall.js        # npm post-install script
└── README.md             # This file
```

## Troubleshooting

### Installation Issues

**Problem**: `npm install -g @mindfiredigital/deep-scan-bot` fails

**Solutions**:

- Ensure you have Node.js 18+ installed: `node --version`
- Check npm permissions: `npm config get prefix`
- Try with sudo (Unix): `sudo npm install -g @mindfiredigital/deep-scan-bot`
- Clear npm cache: `npm cache clean --force`

**Problem**: "Unsupported platform" error during installation

**Solutions**:

- Verify your OS and architecture: `node -e "console.log(process.platform, process.arch)"`
- Supported platforms: macOS (x64, arm64), Linux (x64, arm64), Windows (x64)
- Ensure you're using a 64-bit version of Node.js

**Problem**: "Command not found" after installation

**Solutions**:

- Check npm global bin directory is in your PATH: `npm bin -g`
- Add to PATH: `export PATH=$(npm bin -g):$PATH`
- On Windows, restart your terminal after installation

### Binary Issues

**Problem**: Binary won't execute

**Solutions**:

- Check file permissions: `ls -la $(which deepscanbot)`
- Reapply permissions: `chmod +x $(which deepscanbot)`
- Verify binary integrity: `node scripts/verify-binary.js`

**Problem**: Checksum verification fails

**Solutions**:

- Reinstall the package: `npm install -g @mindfiredigital/deep-scan-bot`
- Check the checksums file: `cat dist/checksums.txt`
- Verify against the GitHub Release checksums

### Crawling Issues

**Problem**: No results or empty output

**Solutions**:

- Verify the URL is accessible: `curl -I https://example.com`
- Increase timeout: `deepscanbot scan <url> timeout=10`
- Disable TLS verification for testing: `deepscanbot scan <url> insecure=true`
- Check if robots.txt is blocking: `deepscanbot scan <url> ignore-robots=true`

**Problem**: Too many requests or being blocked

**Solutions**:

- Reduce concurrency: `deepscanbot scan <url> concurrency=2`
- Add crawl delay: `deepscanbot scan <url> delay=1s`
- Use a proxy: `deepscanbot scan <url> proxy=http://proxy.example.com:8080`

## CI/CD Workflows

### CI Workflow (ci.yml)

Runs on every push and pull request to `main`:

- Builds Go code
- Runs tests with race detection
- Runs golangci-lint
- Checks Go formatting
- Verifies GoReleaser configuration
- Cross-platform build check
- npm package validation

### Release Workflow (release.yml)

Triggered by pushing a Git tag matching `v*`:

- Builds binaries via GoReleaser
- Copies binaries to npm package
- Updates package.json version
- Generates SHA256 checksums
- Verifies binaries
- Publishes to npm registry
- Uploads release artifacts

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

This project is licensed under the MIT License - see [LICENSE.md](LICENSE.md) for details.
</parameter>
<task_progress>
- [x] Implement key=value parsing for scan command
- [x] Rebuild successfully
- [x] Test new key=value syntax
- [x] Update README.md with key=value syntax
- [ ] Update all other documentation files
- [ ] Final validation
</task_progress>
</write_to_file>