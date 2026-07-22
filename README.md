# DeepScanBot CLI

[![CI](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/ci.yml/badge.svg)](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/ci.yml)
[![Release](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/release.yml/badge.svg)](https://github.com/mindfiredigital/DeepScanBot/actions/workflows/release.yml)
[![npm version](https://img.shields.io/npm/v/@mindfiredigital/deepscanbot)](https://www.npmjs.com/package/@mindfiredigital/deepscanbot)
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

DeepScanBot ships as a single, self-contained binary with **zero runtime dependencies**. Choose the method that fits your workflow.

| Method            | Platform              | Best for                            |
| ----------------- | --------------------- | ----------------------------------- |
| **Homebrew** 🍺   | macOS & Linux         | Developers, long-term maintenance   |
| **npm** 📦        | macOS, Linux, Windows | Node.js ecosystem users             |
| **curl** 🌐       | macOS & Linux         | CI/CD, scripting, one-liners        |
| **PowerShell** 🪟 | Windows               | Windows automation                  |
| **Go Install** 🔧 | Any (Go required)     | Go developers, building from source |
| **Manual** 📥     | Any                   | Air-gapped, offline environments    |

---

### 🍺 Homebrew (macOS & Linux)

```bash
# One-time tap setup
brew tap mindfiredigital/tap

# Install
brew install deepscanbot

# Upgrade later
brew upgrade deepscanbot
```

> **Prerequisite:** The [homebrew-tap](https://github.com/mindfiredigital/homebrew-tap) repository must be initialized on GitHub. See [Homebrew Tap Setup](#homebrew-tap-setup) below.

---

### 📦 npm (Cross-Platform)

```bash
npm install -g @mindfiredigital/deepscanbot
```

The npm package automatically detects your OS and architecture. No additional runtime dependencies are required.

```bash
deepscanbot --help
```

---

### 🌐 curl (macOS & Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.sh | bash -s -- -b /usr/local/bin -v v1.0.0
```

**What the script does:**

1. Detects your OS (`linux` / `darwin`) and architecture (`amd64` / `arm64`)
2. Fetches the latest release version from the GitHub API
3. Downloads the binary and `checksums.txt` from the release
4. Verifies the binary against its SHA256 checksum
5. Installs to `/usr/local/bin` (or custom path via `-b`)
6. Runs `deepscanbot version` to confirm the installation

---

### 🪟 PowerShell (Windows)

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.ps1'))
```

Install a specific version:

```powershell
.\install.ps1 -InstallDir "C:\tools" -Version "v1.0.0"
```

The script installs to `$env:ProgramFiles\DeepScanBot` by default and offers to add it to your PATH.

---

### 🔧 Go Install (From Source)

Requires Go 1.22+:

```bash
go install github.com/mindfiredigital/DeepScanBot/apps/cli@latest
```

The binary is placed in `$GOPATH/bin` (typically `~/go/bin`).

---

### 📥 Manual Download

Download the binary for your platform from the [latest release](https://github.com/mindfiredigital/DeepScanBot/releases/latest):

| Platform | Architecture  | Binary Name                     |
| -------- | ------------- | ------------------------------- |
| macOS    | x86_64        | `deepscanbot_darwin_amd64`      |
| macOS    | Apple Silicon | `deepscanbot_darwin_arm64`      |
| Linux    | x86_64        | `deepscanbot_linux_amd64`       |
| Linux    | ARM64         | `deepscanbot_linux_arm64`       |
| Windows  | x86_64        | `deepscanbot_windows_amd64.exe` |

**Linux / macOS:**

```bash
chmod +x deepscanbot_*
sudo mv deepscanbot_linux_amd64 /usr/local/bin/deepscanbot
```

**Windows (PowerShell as Administrator):**

```powershell
Move-Item .\deepscanbot_windows_amd64.exe C:\Windows\System32\deepscanbot.exe
```

---

### ✅ Verify Your Installation

Run these three commands to confirm DeepScanBot is installed correctly:

```bash
# 1. Check the version
deepscanbot version

# 2. Run the diagnostic checks
deepscanbot doctor

# 3. Verify the help output
deepscanbot --help
```

Expected output for `deepscanbot version`:

```
DeepScanBot CLI v1.0.0
```

---

### 🔐 Security Verification

**Checksums:** Every release includes a `checksums.txt` file with SHA256 hashes of all binaries.

```bash
# Linux / macOS (sha256sum)
sha256sum -c checksums.txt --ignore-missing

# macOS (alternative with shasum)
shasum -a 256 -c checksums.txt --ignore-missing

# Windows PowerShell
Get-FileHash .\deepscanbot_windows_amd64.exe -Algorithm SHA256
```

**Attestation:** GitHub signed attestations are generated for every release. Verify provenance with the GitHub CLI:

```bash
gh attestation verify deepscanbot_darwin_amd64 --owner mindfiredigital
```

---

### Homebrew Tap Setup

> **For maintainers:** These steps initialize the `homebrew-tap` repository so GoReleaser can publish Homebrew casks to it.

**1. Create the repository on GitHub:**

```bash
# Visit https://github.com/new
# Owner: mindfiredigital
# Repository name: homebrew-tap
# Visibility: Public
# Do NOT initialize with a README
```

**2. Initialize the repository with the required files:**

````bash
mkdir homebrew-tap && cd homebrew-tap
git init

# Create the tap migration file (redirects old formula users to the new cask)
cat > tap_migrations.json << 'EOF'
{"deepscanbot": "mindfiredigital/tap/deepscanbot"}
EOF

# Create the Casks directory (GoReleaser will populate it)
mkdir Casks
touch Casks/.gitkeep

# Create a README for the tap
cat > README.md << 'EOF'
# Mindfire Digital Homebrew Tap

This tap provides the [DeepScanBot](https://github.com/mindfiredigital/DeepScanBot) CLI tool.

## Usage

```bash
brew tap mindfiredigital/tap
brew install deepscanbot
````

EOF

git add .
git commit -m "chore: initialize homebrew-tap"
git remote add origin https://github.com/mindfiredigital/homebrew-tap.git
git push -u origin main

````

**3. Configure the release secret:**

```bash
# Create a GitHub Personal Access Token with `repo` scope
# Add it as a repository secret in DeepScanBot
gh secret set TAP_GITHUB_TOKEN --repo mindfiredigital/DeepScanBot
````

**4. _(After first release)_ — Migrate existing users:**

Once GoReleaser publishes the first cask, existing users will automatically be redirected by the `tap_migrations.json` file. The old `Formula/` directory can then be removed from `homebrew-tap`.

---

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

## Non-Interactive Mode

DeepScanBot is designed to run reliably in CI/CD pipelines, scripts, and AI agents without ever hanging on user input.

### `--no-input` Flag

Add `--no-input` to any command to disable all interactive prompts. If required input is missing, the CLI fails immediately with a clear error message instead of waiting.

```bash
# Run scan in non-interactive mode
deepscanbot scan https://example.com --no-input --force

# Check version non-interactively
deepscanbot --no-input version --json

# Run doctor in CI
deepscanbot --no-input doctor
```

### `--force` Flag

The `--force` flag allows overwriting existing output files without prompting. In non-interactive mode, the scan command refuses to overwrite an existing output file unless `--force` is explicitly passed.

```bash
# Safe for CI/CD — overwrites output if it exists
deepscanbot scan https://example.com --no-input --force

# Will fail in non-interactive mode without --force
deepscanbot scan https://example.com --no-input
# Error: Output file "crawler_results.txt" already exists.
# Hint: Pass --force to overwrite or use output=<filename>.
```

### TTY Detection

When stdin is not connected to a terminal (e.g., piped input, CI runners), the CLI automatically detects the non-TTY environment and never waits for user input. The `--no-input` flag provides an explicit override for cases where TTY detection is insufficient.

### Best Practices for CI/CD

```bash
# Always use --no-input and --force for automated runs
deepscanbot scan https://example.com --no-input --force --json

# Use exit codes to check success
if deepscanbot scan https://example.com --no-input --force depth=0; then
  echo "Scan completed"
else
  echo "Scan failed with exit code $?"
fi
```

## Exit Codes

DeepScanBot uses standardized exit codes to make CLI failures predictable for scripts, CI/CD pipelines, and AI agents. Every command returns a consistent exit code that tells you exactly what happened.

| Code | Constant          | Description                                  | Example Scenarios                              |
| ---- | ----------------- | -------------------------------------------- | ---------------------------------------------- |
| `0`  | `Success`         | Command completed successfully               | Scan finished, version shown                   |
| `1`  | `InvalidInput`    | Invalid argument or option value             | Malformed URL, unknown flag, missing argument  |
| `2`  | `ValidationError` | Semantic validation failure                  | Empty output filename, invalid option value    |
| `3`  | `AuthFailure`     | Authentication failure with a remote service | Invalid API token, missing credentials         |
| `10` | `AuthzFailure`    | Authenticated but lacking permission         | Insufficient rights to access resource         |
| `20` | `NotFound`        | Requested resource could not be located      | URL/file not found                             |
| `30` | `NetworkFailure`  | Network request failed (non-timeout)         | DNS resolution failure, connection refused     |
| `31` | `Timeout`         | Operation exceeded its configured deadline   | Request timed out, scan exceeded max duration  |
| `70` | `InternalError`   | Unexpected internal error (likely a bug)     | Failed to write output file, serialization bug |

### Checking Exit Codes

```bash
# Check exit code in a script
deepscanbot scan https://example.com
exit_code=$?
echo "Exit code: $exit_code"

# Conditional execution based on exit code
if deepscanbot scan https://example.com depth=0; then
  echo "Scan succeeded"
else
  echo "Scan failed with exit code $?"
fi
```

### Error Messages

All errors include:

- **What went wrong** — a clear description of the problem
- **Why it happened** — the root cause when possible
- **How to fix it** — an actionable hint with an example

```bash
$ deepscanbot scan ftp://example.com
Error: Invalid URL: "ftp://example.com" must be an absolute http:// or https:// URL.
Hint: Example: https://example.com

$ deepscanbot scan http://example.com output=
Error: Output filename must not be empty.
Hint: Use output=<filename> with a non-empty value.
```

## Usage

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
│   ├── exitcode/         # Standardized exit codes and error handling
│   ├── noinput/          # Non-interactive mode and TTY detection
│   ├── fetcher/          # HTTP fetching
│   ├── logger/           # Logging utilities
│   ├── output/           # Output formatting (JSON, human-readable, command tree)
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
