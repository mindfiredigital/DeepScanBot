# DeepScanBot CLI

[![CI](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/ci.yml/badge.svg)](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/ci.yml)
[![Release](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/release.yml/badge.svg)](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/release.yml)
[![npm version](https://img.shields.io/npm/v/@mindfiredigital/deep-scan-bot)](https://www.npmjs.com/package/@mindfiredigital/deep-scan-bot)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mindfiredigital/DeepScanBot)](https://golang.org/)
[![License](https://img.shields.io/github/license/mindfiredigital/DeepScanBot)](LICENSE.md)

**A high-performance web crawler and scanner** — recursively crawls websites, respects robots.txt, handles rate-limiting, and produces comprehensive JSON or text reports. Available as a single, self-contained binary via npm.

## What is DeepScanBot?

DeepScanBot is a **feature-rich, concurrent web crawler** that makes website crawling fast, reliable, and developer-friendly. It handles the complexities of web crawling automatically:

- **Discovers all pages** on a website by following internal links up to a configurable depth
- **Respects crawl rules** defined in robots.txt with an option to bypass when needed
- **Handles rate limits** with automatic retry and exponential backoff
- **Filters content** by MIME type for targeted crawling
- **Exports results** in structured JSON or readable text formats with detailed analytics

### Who Is It For?

- **Security Researchers** - Audit websites for exposed endpoints
- **SEO Specialists** - Discover all pages, find broken links, analyze site structure
- **QA Engineers** - Verify page availability, test website coverage
- **Data Analysts** - Collect structured data from websites for analysis
- **DevOps Engineers** - Monitor website health, validate deployments

## Features

- **⚡ Concurrent Crawling**: Multi-threaded architecture with configurable concurrency, per-host rate limiting, and CPU-aware auto-scaling
- **🛡️ Robots.txt Compliance**: Automatically respects robots.txt rules with optional bypass
- **🔄 Retry & Rate-Limit Handling**: Automatic retry with exponential backoff and Retry-After header parsing
- **📊 Rich Output**: JSON or text reports with detailed summaries, status code distribution, and skip reason breakdowns
- **🗺️ Sitemap Discovery**: Auto-discover and crawl URLs from sitemap.xml, including nested sitemap indexes
- **📁 Content-Type Filtering**: Filter downloads by MIME type, enforce page size limits
- **🌐 Proxy Support**: Route traffic through HTTP/HTTPS proxy servers
- **▶️ Resume Mode**: Resume interrupted crawls without recrawling already visited URLs
- **🌍 Cross-Domain Crawling**: Optionally follow links to external domains
- **🔒 TLS Options**: Disable TLS verification for self-signed certificates

## Installation

### Via npm (Recommended)

```bash
npm install -g @mindfiredigital/deepscanbot
```

After installation, the `deepscanbot` command will be available globally:

```bash
deepscanbot --help
```

No additional runtime dependencies required. The npm package automatically installs the correct binary for your platform.

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

## JSON Output Mode

DeepScanBot supports a consistent `--json` flag across all commands that return structured data. When enabled, all output is written as valid JSON to `stdout`, while progress messages and diagnostics are sent to `stderr`.

### Usage

```bash
# Scan with JSON output
deepscanbot scan https://example.com --json

# Version with JSON output
deepscanbot version --json

# Doctor with JSON output
deepscanbot doctor --json
```

### Sample JSON Output

#### Scan Command

```bash
$ deepscanbot scan https://example.com depth=0 --json
{
  "status": "success",
  "data": {
    "start_url": "https://example.com",
    "output_file": "crawler_results.json",
    "started_at": "2024-01-15T10:30:00Z",
    "finished_at": "2024-01-15T10:30:05Z",
    "duration_ms": 5000,
    "summary": {
      "total": 1,
      "passed": 1,
      "failed": 0,
      "skipped": 0,
      "discovered": 0,
      "max_depth": 0
    },
    "urls": [
      {
        "url": "https://example.com",
        "depth": 0,
        "status_code": 200,
        "content_type": "text/html",
        "result": "passed"
      }
    ],
    "skipped": null
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:05Z",
    "command": "scan",
    "duration_ms": 5000
  }
}
```

#### Version Command

```bash
$ deepscanbot version --json
{
  "status": "success",
  "data": {
    "version": "1.0.0",
    "name": "DeepScanBot CLI"
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "command": "version",
    "duration_ms": 0
  }
}
```

#### Doctor Command

```bash
$ deepscanbot doctor --json
{
  "status": "success",
  "data": {
    "installed": true,
    "executable": true,
    "configured": true,
    "checks_passed": 3,
    "message": "All checks passed!"
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "command": "doctor",
    "duration_ms": 0
  }
}
```

#### Error Response

```bash
$ deepscanbot scan invalid-url --json
{
  "status": "error",
  "error": {
    "message": "invalid URL \"invalid-url\": must be an absolute http:// or https:// URL",
    "code": "invalid_url"
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "command": "scan",
    "duration_ms": 0
  }
}
```

### Key Features

- **Valid JSON Only**: `stdout` contains only valid JSON when `--json` is enabled
- **Separate Streams**: Progress messages, warnings, and logs are written to `stderr`
- **Consistent Format**: All commands use the same JSON response structure
- **Backward Compatible**: Existing behavior is preserved when `--json` is not specified
- **Extensible**: The centralized output formatting makes it easy to add new output formats in the future

### Response Format

All JSON responses follow this structure:

```json
{
  "status": "success | error",
  "data": { ... },        // Present on success
  "error": {              // Present on error
    "message": "Error description",
    "code": "error_code"
  },
  "meta": {
    "timestamp": "ISO8601 timestamp",
    "command": "command_name",
    "duration_ms": 1234
  }
}
```

### Piping and Automation

The JSON output mode is designed for easy integration with scripts and tools:

```bash
# Parse with jq
deepscanbot scan https://example.com --json | jq '.data.summary'

# Extract specific fields
deepscanbot version --json | jq -r '.data.version'

# Check command success
if deepscanbot scan https://example.com --json | jq -e '.status == "success"'; then
  echo "Scan completed successfully"
fi
```

## Usage

DeepScanBot uses a modern command-based CLI structure similar to git, docker, and kubectl.

### Commands

```bash
deepscanbot <command> [options]
```

| Command      | Description                         |
| ------------ | ----------------------------------- |
| `scan`       | Crawl and analyze a website         |
| `version`    | Show installed version              |
| `doctor`     | Verify installation and environment |
| `config`     | Manage CLI configuration            |
| `completion` | Generate shell completion script    |
| `help`       | Show help for any command           |

### Scan Command

```bash
deepscanbot scan <url> [key=value ...]
```

Options are specified as `key=value` pairs after the URL.

| Option              | Description                                            | Default             |
| ------------------- | ------------------------------------------------------ | ------------------- |
| `depth`             | Maximum crawl depth                                    | `2`                 |
| `timeout`           | Request timeout in seconds                             | `2`                 |
| `proxy`             | Proxy URL (e.g. `http://127.0.0.1:8080`)               | `""`                |
| `json`              | Output as JSON                                         | `false`             |
| `size`              | Page size limit in KB (-1 = no limit)                  | `-1`                |
| `disable-redirects` | Disable following redirects                            | `false`             |
| `show-source`       | Show source of each URL                                | `false`             |
| `insecure`          | Disable TLS verification                               | `false`             |
| `unique`            | Ensure unique URLs                                     | `false`             |
| `concurrency`       | Maximum concurrent requests (0 = CPU count)            | `0`                 |
| `host-concurrency`  | Max concurrent requests per host (0 = use concurrency) | `0`                 |
| `content-types`     | MIME types to download (quoted, space/comma separated) | `"text/html"`       |
| `output`            | Output filename without extension                      | `"crawler_results"` |
| `ignore-robots`     | Ignore robots.txt restrictions                         | `false`             |
| `cross-domain`      | Follow links to other hosts                            | `false`             |
| `retries`           | Number of retry attempts                               | `0`                 |
| `retry-backoff`     | Base retry backoff duration (e.g. `500ms`, `2s`)       | `1s`                |
| `delay`             | Politeness delay between requests to same host         | `0`                 |
| `sitemap`           | Discover URLs from /sitemap.xml                        | `false`             |
| `resume`            | Load existing output and avoid recrawling              | `false`             |

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
npm install -g @mindfiredigital/deepscanbot
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
npm install -g @mindfiredigital/deepscanbot
```

### Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.x.x): Breaking changes
- **MINOR** (x.1.x): New features, backward compatible
- **PATCH** (x.x.1): Bug fixes, backward compatible

Pre-release versions use the `-next` tag on npm:

```bash
npm install -g @mindfiredigital/deepscanbot@next
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
│   └── prepublish-check.js
│
├── .github/
│   └── workflows/
│       ├── ci.yml             # CI workflow
│       └── release.yml        # Release workflow
│       └── release-docs.yml   # Release docs workflow
│
├── .goreleaser.yml       # GoReleaser configuration
├── package.json          # npm package configuration
├── postinstall.js        # npm post-install script
└── README.md             # This file
```

## Troubleshooting

### Installation Issues

**Problem**: `npm install -g @mindfiredigital/deepscanbot` fails

**Solutions**:

- Ensure you have Node.js 18+ installed: `node --version`
- Check npm permissions: `npm config get prefix`
- Try with sudo (Unix): `sudo npm install -g @mindfiredigital/deepscanbot`
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
