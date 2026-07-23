# Contributing to DeepScanBot

Thank you for your interest in contributing to DeepScanBot! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)

## Code of Conduct

By participating in this project, you agree to abide by the [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Please be respectful and constructive in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/DeepScanBot.git
   cd DeepScanBot
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/mindfiredigital/DeepScanBot.git
   ```

## Development Setup

### Prerequisites

- **Go** (version 1.22 or higher)
- **Git**

### Installation

1. Install dependencies:

   ```bash
   go mod download
   ```

2. Install development tools:

   ```bash
   go install mvdan.cc/gofumpt@latest
   go install github.com/daixiang0/gci@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install github.com/evilmartians/lefthook@latest
   ```

3. Install git hooks:

   ```bash
   lefthook install
   ```

4. Build the project:

   ```bash
   go build -o deepscanbot ./apps/cli
   ```

5. Verify the installation:
   ```bash
   go run ./apps/cli -h
   ```

### Project Structure

```
DeepScanBot/
├── apps/
│   └── cli/              # Go CLI application
│       ├── main.go
│       └── tests/
├── packages/
│   ├── crawler/          # Web crawling logic
│   ├── exitcode/         # Standardized exit codes and error handling
│   ├── fetcher/          # HTTP fetching
│   ├── logger/           # Logging utilities
│   ├── noinput/          # Non-interactive mode and TTY detection
│   ├── output/           # Output formatting (JSON, human-readable, command tree)
│   ├── parser/           # HTML parsing
│   ├── storage/          # Output storage
│   └── types/            # Shared types
├── scripts/              # Helper scripts
├── .github/
│   └── workflows/        # CI/CD workflows
├── .goreleaser.yml       # GoReleaser configuration
├── package.json          # npm package configuration
├── postinstall.js        # npm post-install script
├── go.mod                # Go module definition
└── go.sum                # Go module checksums
```

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue using the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include:

- Clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, DeepScanBot version)
- Command used and any relevant logs

### Suggesting Features

Feature requests are welcome! Please use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- Clear description of the feature
- Problem statement
- Proposed solution
- Use cases

### Contributing Code

1. Create a new branch from `main`:

   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feat/your-feature-name
   ```

2. Make your changes following the [coding standards](#coding-standards)

3. Add or update tests for your changes

4. Ensure all tests pass:

   ```bash
   go test ./...
   ```

5. Commit your changes with a clear message:

   ```bash
   git add .
   git commit -m "feat: add feature X"
   ```

6. Push to your fork:

   ```bash
   git push origin feat/your-feature-name
   ```

7. Create a Pull Request using the [PR template](.github/PULL_REQUEST_TEMPLATE.md)

## Pull Request Process

1. Ensure your PR description clearly describes the changes
2. Link any related issues (e.g., "Closes #123")
3. All tests must pass
4. Code must be linted and follow project standards
5. At least one maintainer must approve the PR
6. The PR will be squashed and merged into the `main` branch

### PR Title Convention

Use conventional commits format:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofumpt` (strict gofmt) to format your code:
  ```bash
  gofumpt -w .
  ```
- Use `goimports` or `gci` for import management:
  ```bash
  goimports -w .
  # or
  gci write .
  ```
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### General Guidelines

- Write clear, self-documenting code
- Add comments for complex logic
- Keep functions small and focused
- Use meaningful variable and function names
- Handle errors explicitly
- Write unit tests for new functionality

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. Use `cocogitto` or a custom commit hook to validate commit messages:

```bash
# Using cocogitto
cog commit

# Or use lefthook for pre-commit hooks
lefthook install
```

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

**Example:**

```
feat(crawler): add support for custom user agents

- Add UserAgent option to crawler configuration
- Update HTTP request headers
- Add tests for custom user agent functionality

Closes #42
```

## Testing

### Running Tests

Run all tests:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./crawler
```

Run tests with verbose output:

```bash
go test -v ./...
```

### Writing Tests

- Place test files alongside the code they test (e.g., `crawler_test.go` for `crawler.go`)
- Use table-driven tests where appropriate
- Test both success and failure cases
- Mock external dependencies when necessary

## Reporting Bugs

Found a bug? Please help us fix it by:

1. Searching existing issues to avoid duplicates
2. Creating a new issue using the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md)
3. Providing as much detail as possible
4. Including steps to reproduce

## Suggesting Features

Have an idea? We'd love to hear it:

1. Search existing feature requests
2. Create a new issue using the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md)
3. Explain the problem you're trying to solve
4. Describe your proposed solution

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Reach out to the maintainers

Thank you for contributing to DeepScanBot! 🚀
