# Development Tools

This document describes the development toolchain used in DeepScanBot, including code formatting, linting, import management, and git hooks.

## Toolchain Overview

| Tool | Purpose | Version |
|------|---------|---------|
| [gofumpt](https://github.com/mvdan/gofumpt) | Strict Go formatting | v0.10.0+ |
| [gci](https://github.com/daixiang0/gci) | Import management | v0.14.0+ |
| [golangci-lint](https://github.com/golangci/golangci-lint) | Linting and static analysis | v1.64.8+ |
| [lefthook](https://github.com/evilmartians/lefthook) | Git hooks manager | v1.13.6+ |

## Installation

Install all development tools with a single command:

```bash
go install mvdan.cc/gofumpt@latest
go install github.com/daixiang0/gci@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/evilmartians/lefthook@latest
```

## Git Hooks (Lefthook)

Lefthook manages git hooks to enforce code quality standards before commits and pushes.

### Installation

```bash
lefthook install
```

### Available Hooks

#### Pre-commit Hook

Runs automatically before each commit:

1. **Format** - Formats Go code with `gofumpt`
2. **Imports** - Organizes imports with `gci`
3. **Lint** - Runs `golangci-lint` with auto-fix
4. **Conventional Commit** - Validates commit message format

#### Commit-msg Hook

Validates that commit messages follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

**Format:**
```
<type>(<scope>): <subject>
```

**Valid Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks
- `perf` - Performance improvements
- `ci` - CI/CD changes
- `build` - Build system changes
- `revert` - Revert previous commit
- `merge` - Merge commit

**Example:**
```
feat(crawler): add support for custom user agents

- Add UserAgent option to crawler configuration
- Update HTTP request headers
- Add tests for custom user agent functionality

Closes #42
```

#### Pre-push Hook

Runs automatically before pushing to remote:

1. **Test** - Executes all tests with `go test ./...`
2. **Lint-all** - Runs full `golangci-lint` analysis

## Code Formatting (gofumpt)

[gofumpt](https://github.com/mvdan/gofumpt) is a stricter version of `gofmt` that enforces additional formatting rules.

### Usage

Format all Go files:
```bash
gofumpt -w .
```

Format specific files:
```bash
gofumpt -w ./crawler ./fetcher
```

### Features

- Enforces standard `gofmt` rules
- Groups standard library imports separately
- Adds blank lines after imports
- Aligns declarations
- And more strict formatting rules

## Import Management (gci)

[gci](https://github.com/daixiang0/gci) organizes Go imports into sections.

### Usage

Write imports for all files:
```bash
gci write -s standard -s default -s "prefix(github.com/mindfiredigital/DeepScanBot)" .
```

### Import Order

1. **Standard library** - Go standard library packages
2. **Default** - Third-party packages
3. **Project prefix** - Internal packages prefixed with `github.com/mindfiredigital/DeepScanBot`

### Example

Before:
```go
import (
    "fmt"
    "github.com/mindfiredigital/DeepScanBot/crawler"
    "net/http"
    "github.com/temoto/robotstxt"
)
```

After:
```go
import (
    "fmt"
    "net/http"

    "github.com/temoto/robotstxt"

    "github.com/mindfiredigital/DeepScanBot/crawler"
)
```

## Linting (golangci-lint)

[golangci-lint](https://github.com/golangci/golangci-lint) is a fast, parallel linter aggregator for Go.

### Usage

Run linter:
```bash
golangci-lint run
```

Run with auto-fix:
```bash
golangci-lint run --fix
```

Run specific linters:
```bash
golangci-lint run --disable-all --enable=gofmt,revive,gosec
```

### Enabled Linters

The project uses 30+ linters including:

- **gofmt** - Code formatting
- **goimports** - Import organization
- **govet** - Go vet checks
- **gosec** - Security vulnerabilities
- **revive** - Fast, configurable linter
- **staticcheck** - Static analysis
- **errcheck** - Error handling
- **ineffassign** - Ineffective assignments
- **unused** - Unused code detection
- **misspell** - Spelling mistakes
- **And many more...**

### Configuration

See the [golangci-lint configuration](https://github.com/mindfiredigital/DeepScanBot/blob/main/.golangci.yml) for the full configuration.

## Workflow

### Typical Development Workflow

1. **Create a branch:**
   ```bash
   git checkout development
   git pull upstream development
   git checkout -b feat/your-feature-name
   ```

2. **Make changes** to the code

3. **Stage changes:**
   ```bash
   git add .
   ```

4. **Commit** (hooks run automatically):
   ```bash
   git commit -m "feat(crawler): add new feature"
   ```
   
   The pre-commit hook will:
   - Format your code with gofumpt
   - Organize imports with gci
   - Run golangci-lint with auto-fix
   - Validate your commit message

5. **Push** (hooks run automatically):
   ```bash
   git push origin feat/your-feature-name
   ```
   
   The pre-push hook will:
   - Run all tests
   - Run full linting

### Bypassing Hooks (Not Recommended)

In case of emergency, you can bypass hooks:

```bash
git commit --no-verify -m "emergency fix"
git push --no-verify origin main
```

**Note:** This should only be used in exceptional circumstances.

## Troubleshooting

### Lefthook not running

Ensure lefthook is installed:
```bash
lefthook install
```

### golangci-lint takes too long

Increase timeout:
```bash
golangci-lint run --timeout 10m
```

Or run only on changed files:
```bash
golangci-lint run --timeout 5m $(git diff --name-only --diff-filter=d | grep '\.go$')
```

### Import order issues

Regenerate imports:
```bash
gci write -s standard -s default -s "prefix(github.com/mindfiredigital/DeepScanBot)" .
```

## IDE Setup

### Visual Studio Code

**Recommended Extensions:**
- **Go** (golang.go) - Official Go extension with IntelliSense, debugging, and testing
- **GitLens** (eamodio.gitlens) - Git integration and blame annotations
- **Error Lens** (usernamehw.errorlens) - Inline error highlighting
- **Git Graph** (mhutchie.git-graph) - Visualize Git history

**Settings (`.vscode/settings.json`):**
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "gofumpt",
  "go.formatOnSave": true,
  "go.importOrder": true,
  "go.testFlags": ["-v"],
  "editor.codeActionsOnSave": {
    "source.organizeImports": true,
    "source.fixAll": true
  }
}
```

**Tasks (`.vscode/tasks.json`):**
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Run Tests",
      "type": "shell",
      "command": "go test ./...",
      "group": {
        "kind": "test",
        "isDefault": true
      }
    },
    {
      "label": "Run Linter",
      "type": "shell",
      "command": "golangci-lint run",
      "group": "build"
    },
    {
      "label": "Format Code",
      "type": "shell",
      "command": "gofumpt -w .",
      "group": "build"
    }
  ]
}
```

### GoLand / IntelliJ IDEA

**Configuration:**
1. **Go Version**: Set to 1.21+
2. **Go Tools**: Enable gofumpt and gci
3. **Code Style**: Set to Go with gofumpt formatting
4. **Inspections**: Enable golangci-lint integration

**Settings:**
- **Go → Go Modules**: Enable Go Modules integration
- **Editor → Code Style → Go**: Set import layout to match gci
- **Tools → Go Tools**: Configure gofumpt as formatter

### Vim / Neovim

**With coc.nvim:**
```json
{
  "go.gopls": {
    "ui.semanticTokens": true
  },
  "go.formatTool": "gofumpt",
  "go.lintTool": "golangci-lint",
  "go.importOrder": true
}
```

**With nvim-lspconfig:**
```lua
require'lspconfig'.gopls.setup{
  settings = {
    gopls = {
      formatting = true,
      gofumpt = true,
    }
  }
}
```

## Debugging

### Using Delve (dlv)

**Installation:**
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

**Basic Debugging:**
```bash
# Debug main package
dlv debug

# Debug with arguments
dlv debug -- -url https://example.com -depth 2

# Set breakpoint
dlv debug
(dlv) break main.go:42
(dlv) continue
```

**VS Code Debug Configuration (`.vscode/launch.json`):**
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch DeepScanBot",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["-url", "https://example.com", "-depth", "2"]
    },
    {
      "name": "Test Current File",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${file}"
    }
  ]
}
```

**Common Debug Scenarios:**

**1. Debug URL Processing:**
```bash
dlv debug -- -url https://example.com -depth 2
(dlv) break crawler/crawler.go:processURL
(dlv) continue
```

**2. Debug HTTP Fetching:**
```bash
dlv debug -- -url https://example.com
(dlv) break fetcher/fetcher.go:Fetch
(dlv) continue
```

**3. Debug HTML Parsing:**
```bash
dlv debug -- -url https://example.com
(dlv) break parser/parser.go:Parse
(dlv) continue
```

### Logging for Debugging

**Enable Debug Logs:**
```go
// In main.go or specific package
logger.SetLevel(logger.DebugLevel)
```

**Add Strategic Log Points:**
```go
logger.Debug("Fetching URL: %s (depth: %d)", url, depth)
logger.Debug("Discovered %d URLs from %s", len(urls), url)
logger.Debug("Adding to queue: %s", discoveredURL)
```

### Profiling

**CPU Profiling:**
```bash
# Run with profiling enabled
go run main.go -url https://example.com -cpuprofile cpu.prof

# Analyze profile
go tool pprof cpu.prof
(pprof) top10
(pprof) list processURL
```

**Memory Profiling:**
```bash
# Run with memory profiling
go run main.go -url https://example.com -memprofile mem.prof

# Analyze profile
go tool pprof mem.prof
```

**Trace Profiling:**
```bash
# Run with execution trace
go run main.go -url https://example.com -trace trace.out

# View trace
go tool trace trace.out
```

## Testing

### Running Tests

**All Tests:**
```bash
go test ./...
```

**Specific Package:**
```bash
go test ./crawler
go test ./fetcher
go test ./parser
```

**With Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Verbose Output:**
```bash
go test -v ./...
```

**Benchmark Tests:**
```bash
go test -bench=. ./...
```

### Writing Tests

**Table-Driven Tests:**
```go
func TestParseHTML(t *testing.T) {
    tests := []struct {
        name     string
        html     string
        expected []string
    }{
        {
            name:     "simple link",
            html:     `<a href="/about">About</a>`,
            expected: []string{"https://example.com/about"},
        },
        {
            name:     "multiple links",
            html:     `<a href="/a">A</a><a href="/b">B</a>`,
            expected: []string{"https://example.com/a", "https://example.com/b"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser([]string{"text/html"}, baseURL)
            result, err := parser.Parse([]byte(tt.html), baseURL)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**Mocking HTTP Responses:**
```go
func TestFetchWithRetry(t *testing.T) {
    // Create test server
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("test content"))
    }))
    defer ts.Close()

    // Test fetch
    fetcher := NewFetcher(FetcherOptions{...})
    resp, err := fetcher.Fetch(ts.URL)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Troubleshooting

### Common Development Issues

#### Lefthook not running

Ensure lefthook is installed:
```bash
lefthook install
```

Check hook status:
```bash
lefthook run pre-commit --all-files
```

#### golangci-lint takes too long

Increase timeout:
```bash
golangci-lint run --timeout 10m
```

Or run only on changed files:
```bash
golangci-lint run --timeout 5m $(git diff --name-only --diff-filter=d | grep '\.go$')
```

#### Import order issues

Regenerate imports:
```bash
gci write -s standard -s default -s "prefix(github.com/mindfiredigital/DeepScanBot)" .
```

#### Build fails with "undefined" errors

Ensure dependencies are up to date:
```bash
go mod tidy
go mod download
```

#### Tests fail intermittently

Tests may be flaky due to:
- Network dependencies (mock external calls)
- Timing issues (use `time.Sleep` sparingly)
- Concurrent access (ensure proper synchronization)

Run tests sequentially:
```bash
go test -parallel 1 ./...
```

### Performance Profiling

**Identify Slow Code:**
```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./crawler
go tool pprof cpu.prof

# Look for hot paths
(pprof) top10
(pprof) list crawl
```

**Memory Leaks:**
```bash
# Memory profiling
go test -memprofile=mem.prof ./crawler
go tool pprof mem.prof

# Check for unexpected allocations
(pprof) list processURL
```

**Goroutine Leaks:**
```bash
# Enable debug logging
GODEBUG=gctrace=1 go run main.go -url https://example.com

# Or use runtime/pprof
import _ "net/http/pprof"
# Visit http://localhost:6060/debug/pprof/goroutine?debug=1
```

## Contributing

When contributing to DeepScanBot:

1. Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification
2. Ensure all tests pass: `go test ./...`
3. Run the linter: `golangci-lint run`
4. Format your code: `gofumpt -w .`
5. Organize imports: `gci write -s standard -s default -s "prefix(github.com/mindfiredigital/DeepScanBot)" .`

See the [Contributing Guide](contributing) for more details.
