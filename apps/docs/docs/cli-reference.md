---
sidebar_position: 3
---

# CLI Reference

This section provides complete reference documentation for all DeepScanBot CLI commands and options.

## Commands

| Command | Description |
|---------|-------------|
| `scan` | Crawl and analyze a website |
| `version` | Show installed version |
| `doctor` | Verify installation and environment |
| `config` | Manage CLI configuration |
| `completion` | Generate shell completion script |
| `help` | Show help for any command |

## Global Flags

These flags work with any command (scan, version, doctor, etc.):

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | `false` | Output results in JSON format |
| `--no-input` | `false` | Disable all interactive prompts; fail if required input is missing |
| `--verbose` | `false` | Display additional informational messages |
| `--debug` | `false` | Display detailed debugging information |
| `--quiet` | `false` | Suppress non-essential output (only show warnings and errors) |
| `--dry-run` | `false` | Preview actions that would be performed without making changes |

## Scan Options

These options are specific to the `scan` command and can be provided as flags (`--option=value`) or key=value pairs:

| Option | Default | Description |
|--------|---------|-------------|
| `--depth` | `2` | Maximum crawl depth |
| `--timeout` | `2` | Request timeout in seconds |
| `--concurrency` | `8` | Max concurrent workers |
| `--host-concurrency` | `2` | Max concurrent requests per host (0 = use effective concurrency) |
| `--content-types` | `"text/html"` | Allowed MIME types |
| `--output` | `"crawler_results"` | Output filename without extension |
| `--json` | `false` | JSON output (scan-specific flag) |
| `--size` | `-1` | Page size limit in bytes (-1 = unlimited) |
| `--proxy` | `""` | Proxy URL |
| `--unique` | `false` | Unique URLs only |
| `--show-source` | `false` | Show URL source in output |
| `--disable-redirects` | `false` | Disable following HTTP redirects |
| `--insecure` | `false` | Disable TLS verification |
| `--retries` | `0` | Retry attempts for failed requests |
| `--retry-backoff` | `"1s"` | Retry backoff duration |
| `--delay` | `"0s"` | Politeness delay between requests |
| `--sitemap` | `false` | Discover URLs from sitemap.xml |
| `--resume` | `false` | Resume from previous crawl results |
| `--ignore-robots` | `false` | Ignore robots.txt rules |
| `--cross-domain` | `false` | Follow external links |
| `--force` | `false` | Overwrite existing output file without prompting |
| `--yes` | `false` | Auto-confirm all destructive operations |
| `--input-file` | `""` | Read URLs from a file (one per line) |
| `--stdin` | `false` | Read URLs from standard input (one per line) |

> **Note on flag scopes:**
> - **Global flags** (`--json`, `--no-input`, etc.) work with any command
> - **Scan options** (`--depth`, `--timeout`, etc.) only work with the `scan` command
> - Scan options support both flag syntax (`--depth=3`) and key=value syntax (`depth=3`) for backward compatibility

> **Note:** For detailed examples and usage instructions, see the [Usage Guide](/docs/guide/usage).
