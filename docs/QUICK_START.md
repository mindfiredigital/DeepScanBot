# Quick Start Guide

Get DeepScanBot up and running in 5 minutes.

## For End Users

### Installation

```bash
npm install -g @mindfiredigital/deep-scan-bot
```

### Verify Installation

```bash
deepscanbot --version
deepscanbot -h
```

### First Crawl

```bash
deepscanbot scan https://example.com depth=2
```

That's it! No Go installation required.

## For Developers

### Prerequisites

- Go 1.22.4+
- Node.js 18+
- Git

### Clone and Build

```bash
# Clone repository
git clone https://github.com/mindfiredigital/DeepScanBot.git
cd DeepScanBot

# Install dependencies
go mod download

# Build binary
go build -o deepscanbot ./apps/cli

# Test
./deepscanbot --version
```

### Run Without Building

```bash
go run ./apps/cli -url https://example.com depth=2
```

### Run Tests

```bash
# Go tests
go test ./...

# Lint
golangci-lint run ./...
```

## For Contributors

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/mindfiredigital/DeepScanBot.git
cd DeepScanBot

# Install development tools
go install mvdan.cc/gofumpt@latest
go install github.com/daixiang0/gci@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/evilmartians/lefthook@latest && lefthook install

# Install dependencies
go mod download

# Install git hooks
lefthook install
```

### Development Workflow

```bash
# 1. Create feature branch
git checkout -b feat/your-feature

# 2. Make changes and format
gofumpt -l -w .
gci write --skip-vendor show-source=true standard show-source=true default show-source=true "prefix(github.com/mindfiredigital/DeepScanBot)" .

# 3. Run tests
go test ./...
golangci-lint run ./...

# 4. Test npm package locally
goreleaser build --snapshot --clean
node scripts/verify-binary.js
npm pack
npm install -g ./deep-scan-bot-*.tgz
deepscanbot --version
npm uninstall -g @mindfiredigital/deep-scan-bot

# 5. Commit and push
git add .
git commit -m "feat: add new feature"
git push origin feat/your-feature
```

### Pre-Release Checklist

Before submitting a PR:

- [ ] All tests pass: `go test ./...`
- [ ] Linting passes: `golangci-lint run ./...`
- [ ] Code is formatted: `gofumpt -l -w .`
- [ ] Imports organized: `gci write ...`
- [ ] npm package builds: `goreleaser build --snapshot --clean`
- [ ] Binary verification passes: `node scripts/verify-binary.js`
- [ ] Documentation updated

## For Release Managers

### Creating a Release

```bash
# 1. Ensure you're on main branch with latest code
git checkout main
git pull origin main

# 2. Run all tests
go test ./...
golangci-lint run ./...

# 3. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 4. Monitor GitHub Actions
# Visit: https://github.com/mindfiredigital/DeepScanBot/actions

# 5. Verify release
npm view @mindfiredigital/deep-scan-bot version
npm install -g @mindfiredigital/deep-scan-bot
deepscanbot --version
```

### What Happens Automatically

When you push a tag, GitHub Actions:

1. ✅ Builds binaries for all platforms (Linux, macOS, Windows)
2. ✅ Generates SHA256 checksums
3. ✅ Copies binaries to npm package structure
4. ✅ Updates package.json version
5. ✅ Publishes to npm registry
6. ✅ Creates GitHub Release with binaries
7. ✅ Deploys documentation (if configured)

### Post-Release Verification

```bash
# Check npm package
npm view @mindfiredigital/deep-scan-bot

# Install and test
npm install -g @mindfiredigital/deep-scan-bot
deepscanbot --version
deepscanbot -h

# Check GitHub Release
# Visit: https://github.com/mindfiredigital/DeepScanBot/releases
```

## Common Tasks

### Update Dependencies

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Update npm dependencies (if any)
npm update
```

### Build for Specific Platform

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o deepscanbot-linux-amd64 ./apps/cli

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o deepscanbot-darwin-arm64 ./apps/cli

# Windows
GOOS=windows GOARCH=amd64 go build -o deepscanbot-windows-amd64.exe ./apps/cli
```

### Test Local Changes

```bash
# Quick test with go run
go run ./apps/cli -url https://example.com depth=1

# Build and test
go build -o deepscanbot ./apps/cli
./deepscanbot scan https://example.com depth=1

# Test via npm (full integration test)
goreleaser build --snapshot --clean
npm pack
npm install -g ./deep-scan-bot-*.tgz
deepscanbot scan https://example.com depth=1
npm uninstall -g @mindfiredigital/deep-scan-bot
```

### Debug Binary Issues

```bash
# Check binary format
file deepscanbot

# Check binary permissions (Unix)
ls -l deepscanbot

# Make executable (Unix)
chmod +x deepscanbot

# Test binary directly
./deepscanbot --version
./deepscanbot -h

# Run with debug output
./deepscanbot scan https://example.com depth=1 -v
```

## Troubleshooting

### "Command not found" after npm install

```bash
# Find npm global bin
npm bin -g

# Add to PATH (bash/zsh)
echo 'export PATH=$(npm bin -g):$PATH' >> ~/.bashrc
source ~/.bashrc

# Add to PATH (fish)
set -U fish_user_paths (npm bin -g) $fish_user_paths
```

### "Unsupported platform" error

Check supported platforms:
- macOS (Intel x64, Apple Silicon arm64)
- Linux (amd64, arm64)
- Windows (amd64)

If you're on a supported platform, the package may not have been built correctly. Report an issue.

### Permission denied

```bash
# Find binary
which deepscanbot

# Set executable permissions
chmod +x $(which deepscanbot)
```

### Binary won't execute

```bash
# Reinstall package
npm uninstall -g @mindfiredigital/deep-scan-bot
npm install -g @mindfiredigital/deep-scan-bot

# Check binary integrity
node scripts/verify-binary.js
```

## Getting Help

- 📖 [Documentation](https://mindfiredigital.github.io/DeepScanBot/)
- 🐛 [Issue Tracker](https://github.com/mindfiredigital/DeepScanBot/issues)
- 💬 [Discussions](https://github.com/mindfiredigital/DeepScanBot/discussions)
- 📧 [Email](mailto:deepscanbot@mindfiresolutions.com)

## Next Steps

- Read the [Usage Guide](/docs/guide/usage) for detailed CLI options
- Explore [Features](/docs/guide/features) to learn about capabilities
- Check [Architecture](/docs/architecture) to understand the codebase
- Review [Contributing Guide](/docs/contribution-guide/how-to-contribute) to contribute