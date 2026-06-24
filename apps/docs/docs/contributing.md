# Contributing

Thank you for your interest in contributing to DeepScanBot! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by the [CODE_OF_CONDUCT.md](https://github.com/mindfiredigital/DeepScanBot/blob/main/CODE_OF_CONDUCT.md). Please be respectful and constructive in all interactions.

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

- **Go** (version 1.21 or higher)
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
   go build -o deepscanbot .
   ```

5. Verify the installation:
   ```bash
   ./deepscanbot -h
   ```

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue using the [bug report template](https://github.com/mindfiredigital/DeepScanBot/blob/main/.github/ISSUE_TEMPLATE/bug_report.md). Include:

- Clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, DeepScanBot version)
- Command used and any relevant logs

### Suggesting Features

Feature requests are welcome! Please use the [feature request template](https://github.com/mindfiredigital/DeepScanBot/blob/main/.github/ISSUE_TEMPLATE/feature_request.md) and include:

- Clear description of the feature
- Problem statement
- Proposed solution
- Use cases

### Contributing Code

1. Create a new branch from `development`:

   ```bash
   git checkout development
   git pull upstream development
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

7. Create a Pull Request using the [PR template](https://github.com/mindfiredigital/DeepScanBot/blob/main/.github/PULL_REQUEST_TEMPLATE.md)

## Pull Request Process

### Before Submitting

1. **Ensure your branch is up to date:**
   ```bash
   git checkout development
   git pull upstream development
   git checkout feat/your-feature-name
   git rebase development
   ```

2. **Run all checks:**
   ```bash
   # Format code
   gofumpt -w .
   
   # Organize imports
   gci write -s standard -s default -s "prefix(github.com/mindfiredigital/DeepScanBot)" .
   
   # Run linter
   golangci-lint run
   
   # Run tests
   go test ./...
   ```

3. **Update documentation:**
   - Update README.md if needed
   - Add/update code comments
   - Update docs/ if adding features

### Submitting a PR

1. **Push your branch:**
   ```bash
   git push origin feat/your-feature-name
   ```

2. **Create Pull Request:**
   - Use the PR template
   - Provide clear description of changes
   - Link related issues (e.g., "Closes #123")
   - Add screenshots/examples if applicable

3. **PR Description Should Include:**
   - Summary of changes
   - Motivation and context
   - Testing performed
   - Screenshots (if UI changes)
   - Breaking changes (if any)

### PR Title Convention

Use conventional commits format:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks
- `perf:` - Performance improvements
- `ci:` - CI/CD changes

**Examples:**
```
feat(crawler): add support for custom headers
fix(parser): handle malformed HTML gracefully
docs(usage): add proxy configuration examples
refactor(fetcher): simplify retry logic
test(crawler): add unit tests for URL deduplication
perf(parser): optimize HTML tokenization
```

### Review Process

1. **Automated Checks:**
   - CI/CD pipeline runs tests
   - Linting checks pass
   - Build succeeds

2. **Code Review:**
   - At least one maintainer must approve
   - Address review comments
   - Update PR as needed

3. **Final Approval:**
   - All checks pass
   - Approvals obtained
   - No unresolved discussions

4. **Merge:**
   - PR will be squashed and merged into `development`
   - Branch will be deleted after merge

## Coding Standards

### Go Code Style

Follow these guidelines to maintain consistent, high-quality code:

#### Formatting

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofumpt` (strict gofmt) to format your code:
  ```bash
  gofumpt -w .
  ```
- Use `gci` for import management:
  ```bash
  gci write -s standard -s default -s "prefix(github.com/mindfiredigital/DeepScanBot)" .
  ```
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

#### Naming Conventions

**Packages:**
- Use lowercase, single-word names
- Avoid underscores or mixed caps
- Examples: `crawler`, `parser`, `storage`

**Functions:**
- Use `MixedCaps` or `mixedCaps` (not `mixed_caps`)
- Be descriptive and verb-based: `FetchURL`, `ParseHTML`, `SaveResults`
- Avoid generic names: `doit`, `process`, `handle`

**Variables:**
- Use `mixedCaps` for local variables
- Use descriptive names: `maxRetries` not `mr`
- Use single letters only for loop indices: `i`, `j`, `k`

**Constants:**
- Use `MixedCaps` for exported constants
- Use `mixedCaps` for unexported constants
- Group related constants together

**Interfaces:**
- Use `-er` suffix for single-method interfaces: `Reader`, `Writer`, `Fetcher`
- Name interfaces after their behavior, not implementation

#### Code Organization

**File Structure:**
- One package per directory
- Group related functionality
- Keep files focused and under 500 lines when possible

**Function Size:**
- Keep functions under 50 lines
- Single responsibility principle
- Extract complex logic into helper functions

**Error Handling:**
- Always check errors explicitly
- Provide context in error messages
- Use `fmt.Errorf` with `%w` for wrapping
- Don't ignore errors with `_`

**Example:**
```go
// Good
func FetchURL(url string) (*http.Response, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
    }
    return resp, nil
}

// Bad
func FetchURL(url string) (*http.Response, error) {
    resp, _ := http.Get(url)  // Ignoring error
    return resp, nil
}
```

### General Guidelines

- Write clear, self-documenting code
- Add comments for complex logic (explain "why", not "what")
- Keep functions small and focused
- Use meaningful variable and function names
- Handle errors explicitly with context
- Write unit tests for new functionality
- Avoid magic numbers (use named constants)
- Prefer composition over inheritance

### Code Review Guidelines

#### As a Reviewer

**What to Look For:**
- **Correctness**: Does the code work as intended?
- **Design**: Is the solution well-designed and maintainable?
- **Performance**: Are there any performance issues?
- **Security**: Are there any security vulnerabilities?
- **Tests**: Are there adequate tests?
- **Documentation**: Is the code well-documented?

**Review Checklist:**
- [ ] Code follows project style guidelines
- [ ] Functions are small and focused
- [ ] Error handling is comprehensive
- [ ] Tests cover success and failure cases
- [ ] No hardcoded values or magic numbers
- [ ] No commented-out code
- [ ] No debug prints or logging leftovers
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced

**Providing Feedback:**
- Be respectful and constructive
- Explain the "why" behind suggestions
- Provide code examples for complex changes
- Distinguish between "must fix" and "nice to have"
- Acknowledge good solutions

**Example Review Comments:**
```
✅ Good: "Consider extracting this logic into a separate function to improve testability"

❌ Bad: "This is wrong"

✅ Good: "This could be optimized by using a map instead of a slice for O(1) lookups"

❌ Bad: "This is slow"
```

#### As an Author

**Before Requesting Review:**
- [ ] Self-review your code
- [ ] Run all tests: `go test ./...`
- [ ] Run linter: `golangci-lint run`
- [ ] Format code: `gofumpt -w .`
- [ ] Organize imports: `gci write ...`
- [ ] Update documentation
- [ ] Add/update tests
- [ ] Test edge cases

**Responding to Feedback:**
- Be open to suggestions
- Ask clarifying questions
- Explain your reasoning if disagreeing
- Make requested changes promptly
- Mark conversations as resolved

**Common Review Feedback:**

**Performance:**
```
Reviewer: "This loop could be O(n²). Consider using a map for lookups."
Author: "Good point! I'll refactor to use a map for O(1) lookups."
```

**Error Handling:**
```
Reviewer: "This error is not being handled. What happens if the file doesn't exist?"
Author: "You're right. I'll add proper error handling with a descriptive error message."
```

**Testing:**
```
Reviewer: "Please add a test case for the error path."
Author: "Added test case `TestFetchURL_InvalidURL` to cover the error scenario."
```

**Design:**
```
Reviewer: "This function is doing too much. Can we split it into smaller functions?"
Author: "Agreed. I'll extract the validation logic into `validateURL()` and the fetching into `fetchWithRetry()`."
```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. The git hooks will automatically validate your commit messages:

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
- `perf` - Performance improvements
- `ci` - CI/CD changes
- `build` - Build system changes
- `revert` - Revert previous commit

**Subject Line:**
- Use imperative mood: "add" not "added" or "adds"
- Don't capitalize first letter
- No period at the end
- Keep under 50 characters if possible

**Body (optional):**
- Explain what and why, not how
- Wrap at 72 characters
- Can include motivation and context

**Footer (optional):**
- Reference issues: `Closes #123`, `Fixes #456`
- Breaking changes: `BREAKING CHANGE: description`

**Example:**

```
feat(crawler): add support for custom user agents

- Add UserAgent option to crawler configuration
- Update HTTP request headers to use custom user agent
- Add tests for custom user agent functionality
- Update documentation with usage examples

Closes #42
```

**More Examples:**

```
fix(parser): handle relative URLs correctly

Previously, relative URLs were not being resolved correctly
when the base URL had a path component. This fix ensures
proper URL resolution using url.ResolveReference.

Fixes #89
```

```
refactor(fetcher): simplify retry logic

Extract retry logic into separate function to improve
readability and testability. No functional changes.

BREAKING CHANGE: RetryBackoff field renamed to BaseBackoff
```

```
docs(usage): add proxy configuration examples

Add examples for HTTP, HTTPS, and SOCKS5 proxy usage.
Include best practices for proxy configuration.
```

### Testing Standards

- Write tests for all new functionality
- Use table-driven tests for multiple scenarios
- Test both success and failure cases
- Mock external dependencies
- Aim for >80% code coverage
- Tests should be independent and idempotent

**Example Test Structure:**
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    InputType{...},
            expected: OutputType{...},
            wantErr:  false,
        },
        {
            name:     "invalid input",
            input:    InputType{...},
            expected: OutputType{},
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("Function() = %v, want %v", result, tt.expected)
            }
        })
    }
}
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

## Development Tools

For detailed information about the development toolchain, see the [Development Tools](development-tools) documentation.

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Reach out to the maintainers

Thank you for contributing to DeepScanBot! 🚀