---
sidebar_position: 5
---

# Automation & CI/CD

DeepScanBot is designed to run reliably in CI/CD pipelines, scripts, and AI agents without ever hanging on user input.

## Non-Interactive Mode

### `--no-input` Flag

Add `--no-input` to any command to disable all interactive prompts. If required input is missing, the CLI fails immediately with a clear error message instead of waiting.

```bash
# Run scan in non-interactive mode
deepscanbot scan https://example.com --no-input --force

# Check version non-interactively
deepscanbot --no-input --version --json

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

## CI/CD Examples

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

DeepScanBot uses standardized exit codes to make CLI failures predictable for scripts, CI/CD pipelines, and AI agents.

| Code | Constant | Description | Example Scenarios |
|------|----------|-------------|-------------------|
| `0` | `Success` | Command completed successfully | Scan finished, version shown |
| `1` | `InvalidInput` | Invalid argument or option value | Malformed URL, unknown flag |
| `2` | `ValidationError` | Semantic validation failure | Empty output filename |
| `30` | `NetworkFailure` | Network request failed | DNS resolution failure |
| `31` | `Timeout` | Operation exceeded deadline | Request timed out |
| `70` | `InternalError` | Unexpected internal error | Failed to write output |

> **Note:** Exit codes 3 (AuthFailure), 10 (AuthzFailure), and 20 (NotFound) are defined in the codebase for future use but are not currently returned by any runtime errors.

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
- **Error message** — a clear description of what went wrong
- **Hint (optional)** — an actionable suggestion with an example when available

```bash
$ deepscanbot scan ftp://example.com
Error: Invalid URL: "ftp://example.com" must be an absolute http:// or https:// URL.
Hint: Example: https://example.com

$ deepscanbot scan http://example.com output=
Error: Output filename must not be empty.
Hint: Use output=<filename> with a non-empty value.
```
