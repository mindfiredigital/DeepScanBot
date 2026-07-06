# Migration Guide: From Flags to Commands

This guide helps existing users transition from the old flag-based CLI interface to the new command-based interface.

## Overview

DeepScanBot has been redesigned with a modern command-based CLI structure, similar to git, docker, kubectl, and other popular tools. This provides a more intuitive and organized user experience.

## What Changed

### Old Interface (Flag-Based)

```bash
# Old syntax - all flags at the root level
deepscanbot -url https://example.com -depth 3 -json -output results
```

### New Interface (Command-Based)

```bash
# New syntax - commands with options
deepscanbot scan https://example.com --depth 3 --json --output results
```

## Command Structure

### Old: Single command with many flags
```bash
deepscanbot [flags]
```

### New: Subcommands for different operations
```bash
deepscanbot <command> [options]

Commands:
  scan         Crawl and analyze a website
  version      Show installed version
  doctor       Verify installation and environment
  config       Manage CLI configuration
  completion   Generate shell completion
  help         Show help for commands
```

## Flag Migration

### URL Flag → Scan Command Argument

**Old:**
```bash
deepscanbot -url https://example.com
```

**New:**
```bash
deepscanbot scan https://example.com
```

The URL is now a positional argument to the `scan` command instead of a flag.

### All Other Flags → Scan Command Options

All crawler-related flags remain the same but are now options under the `scan` command:

| Old Flag | New Option | Notes |
|----------|-----------|-------|
| `-url <url>` | `scan <url>` | URL is now a positional argument |
| `-depth <n>` | `--depth <n>` | Same flag, now under scan command |
| `-timeout <n>` | `--timeout <n>` | Same flag, now under scan command |
| `-proxy <url>` | `--proxy <url>` | Same flag, now under scan command |
| `-json` | `--json` | Same flag, now under scan command |
| `-size <n>` | `--size <n>` | Same flag, now under scan command |
| `-dr` | `--disable-redirects` | Long form only |
| `-s` | `--show-source` | Long form only |
| `-insecure` | `--insecure` | Same flag, now under scan command |
| `-u` | `--unique` | Short form changed |
| `-concurrency <n>` | `--concurrency <n>` | Same flag, now under scan command |
| `-host-concurrency <n>` | `--host-concurrency <n>` | Same flag, now under scan command |
| `-content-types <types>` | `--content-types <types>` | Same flag, now under scan command |
| `-output <name>` | `--output <name>` | Same flag, now under scan command |
| `-ignore-robots` | `--ignore-robots` | Same flag, now under scan command |
| `-cross-domain` | `--cross-domain` | Same flag, now under scan command |
| `-retries <n>` | `--retries <n>` | Same flag, now under scan command |
| `-retry-backoff <d>` | `--retry-backoff <d>` | Same flag, now under scan command |
| `-delay <d>` | `--delay <d>` | Same flag, now under scan command |
| `-sitemap` | `--sitemap` | Same flag, now under scan command |
| `-resume` | `--resume` | Same flag, now under scan command |

## Migration Examples

### Example 1: Basic Scan

**Old:**
```bash
deepscanbot -url https://example.com
```

**New:**
```bash
deepscanbot scan https://example.com
```

### Example 2: Scan with Depth and JSON Output

**Old:**
```bash
deepscanbot -url https://example.com -depth 3 -json
```

**New:**
```bash
deepscanbot scan https://example.com --depth 3 --json
```

### Example 3: Advanced Scan with All Options

**Old:**
```bash
deepscanbot \
  -url https://docs.example.com \
  -depth 5 \
  -concurrency 20 \
  -host-concurrency 5 \
  -timeout 10 \
  -delay 200ms \
  -retries 3 \
  -retry-backoff 1s \
  -json \
  -sitemap \
  -cross-domain \
  -content-types "text/html application/pdf" \
  -output scan_results
```

**New:**
```bash
deepscanbot scan https://docs.example.com \
  --depth 5 \
  --concurrency 20 \
  --host-concurrency 5 \
  --timeout 10 \
  --delay 200ms \
  --retries 3 \
  --retry-backoff 1s \
  --json \
  --sitemap \
  --cross-domain \
  --content-types "text/html application/pdf" \
  --output scan_results
```

### Example 4: Scan with Proxy

**Old:**
```bash
deepscanbot -url https://example.com -proxy http://127.0.0.1:8080
```

**New:**
```bash
deepscanbot scan https://example.com --proxy http://127.0.0.1:8080
```

### Example 5: Resume Crawl

**Old:**
```bash
deepscanbot -url https://example.com -resume -output crawler_results
```

**New:**
```bash
deepscanbot scan https://example.com --resume --output crawler_results
```

## New Commands

The new CLI structure introduces several new commands that didn't exist before:

### Version Command

```bash
# Show version
deepscanbot version
```

Previously, there was no way to check the version from the CLI.

### Doctor Command

```bash
# Verify installation
deepscanbot doctor
```

Checks that DeepScanBot is properly installed and the environment is configured correctly.

### Config Command

```bash
# View configuration
deepscanbot config

# Set configuration
deepscanbot config set <key> <value>

# Get configuration
deepscanbot config get <key>
```

Manage CLI configuration settings (placeholder for future implementation).

### Completion Command

```bash
# Generate shell completion
deepscanbot completion bash
deepscanbot completion zsh
deepscanbot completion fish
deepscanbot completion powershell
```

Generate shell completion scripts for better CLI experience.

## Global Flags

The new CLI also supports global flags that apply to all commands:

```bash
# Enable debug logging
deepscanbot --debug scan https://example.com

# Suppress non-essential output
deepscanbot --quiet scan https://example.com

# Combine global flags
deepscanbot --debug --quiet scan https://example.com
```

## Help System

The new CLI has an improved help system:

```bash
# Show main help
deepscanbot --help
deepscanbot help

# Show help for scan command
deepscanbot scan --help
deepscanbot help scan

# Show help for specific command
deepscanbot help version
```

### Example Help Output

```bash
$ deepscanbot --help

DeepScanBot CLI

Usage:
  deepscanbot <command> [options]

Commands:
  scan         Crawl and analyze a website
  version      Show installed version
  doctor       Verify installation and environment
  config       Manage CLI configuration
  completion   Generate shell completion
  help         Show help for commands

Options:
  -d, --debug     Enable debug logging
  -q, --quiet     Suppress non-essential output
  -h, --help      Show this help message

Examples:
  deepscanbot scan https://example.com
  deepscanbot scan https://example.com --depth 3
  deepscanbot scan https://example.com --json
  deepscanbot version
  deepscanbot doctor
```

## Breaking Changes

### 1. URL is Now a Positional Argument

**Impact:** Scripts that use `-url` flag will break.

**Migration:**
```bash
# Old
deepscanbot -url https://example.com

# New
deepscanbot scan https://example.com
```

### 2. Some Short Flags Changed

**Impact:** Scripts using `-dr`, `-s`, or `-u` flags need updates.

**Migration:**
```bash
# Old
deepscanbot -url https://example.com -dr -s -u

# New
deepscanbot scan https://example.com --disable-redirects --show-source --unique
```

### 3. Help Flag Behavior

**Impact:** `-h` flag behavior changed slightly.

**Migration:**
```bash
# Old
deepscanbot -h  # Show help

# New
deepscanbot --help  # Show help
deepscanbot help    # Show help
deepscanbot scan --help  # Show scan command help
```

## Automated Migration

### Using sed (Linux/macOS)

Convert old-style commands to new-style:

```bash
# Replace -url with scan command
sed -i 's/-url \([^ ]*\)/scan \1/g' script.sh

# Replace flag names
sed -i 's/-dr/--disable-redirects/g' script.sh
sed -i 's/-s\b/--show-source/g' script.sh
sed -i 's/-u\b/--unique/g' script.sh
```

### Using PowerShell (Windows)

```powershell
# Replace -url with scan command
(Get-Content script.sh) -replace '-url (\S+)', 'scan $1' | Set-Content script.sh

# Replace flag names
(Get-Content script.sh) -replace '-dr', '--disable-redirects' | Set-Content script.sh
(Get-Content script.sh) -replace '-s\b', '--show-source' | Set-Content script.sh
(Get-Content script.sh) -replace '-u\b', '--unique' | Set-Content script.sh
```

## Compatibility Mode (Temporary)

For a transitional period, you can use the old syntax with a compatibility wrapper:

### Create an Alias

```bash
# Add to ~/.bashrc or ~/.zshrc
alias deepscanbot-legacy='deepscanbot scan'
```

### Usage

```bash
# Old syntax still works with alias
deepscanbot-legacy -url https://example.com -depth 3
```

**Note:** This is a temporary workaround. The old flag-based interface is deprecated and will be removed in a future version.

## Testing Your Migration

After updating your scripts, test them:

```bash
# Test basic scan
deepscanbot scan https://example.com --depth 1

# Test with all your usual flags
deepscanbot scan https://example.com --depth 3 --json --output results

# Verify output
ls -la results.json
```

## Common Migration Patterns

### Pattern 1: Simple URL Scan

**Before:**
```bash
deepscanbot -url https://example.com
```

**After:**
```bash
deepscanbot scan https://example.com
```

### Pattern 2: Scan with Options

**Before:**
```bash
deepscanbot -url https://example.com -depth 3 -json -output results
```

**After:**
```bash
deepscanbot scan https://example.com --depth 3 --json --output results
```

### Pattern 3: Complex Scan

**Before:**
```bash
deepscanbot \
  -url https://example.com \
  -depth 5 \
  -concurrency 10 \
  -proxy http://proxy:8080 \
  -json \
  -sitemap \
  -output results
```

**After:**
```bash
deepscanbot scan https://example.com \
  --depth 5 \
  --concurrency 10 \
  --proxy http://proxy:8080 \
  --json \
  --sitemap \
  --output results
```

## Getting Help

If you encounter issues during migration:

1. Check the [Usage Guide](/docs/guide/usage) for new syntax
2. Run `deepscanbot scan --help` to see all available options
3. Run `deepscanbot --help` to see all commands
4. Open an issue on [GitHub](https://github.com/mindfiredigital/DeepScanBot/issues)

## Timeline

- **v1.0.0**: New command-based CLI introduced
- **v1.x.x**: Both old and new syntax supported (with deprecation warnings)
- **v2.0.0**: Old flag-based syntax removed

## Summary

The migration from flags to commands is straightforward:

1. Add `scan` command before the URL
2. Replace `-url <url>` with just `<url>` as a positional argument
3. Keep all other flags mostly the same (just add `--` prefix for long forms)
4. Update any scripts or documentation

The new command-based interface provides:
- Better organization
- Easier discoverability
- More intuitive usage
- Room for future expansion
- Better help system

</parameter>
<parameter=task_progress>
- [x] Examine current CLI implementation
- [x] Add Cobra CLI library
- [x] Redesign CLI with subcommands
- [x] Update Go CLI code (complete rewrite with Cobra)
- [ ] Update all documentation
- [x] Create migration guide
- [ ] Update examples
- [ ] Final validation
</parameter>
</write_to_file>