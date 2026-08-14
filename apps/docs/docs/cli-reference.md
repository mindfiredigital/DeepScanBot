---
sidebar_position: 3
---

# CLI Reference

This section provides complete reference documentation for all DeepScanBot CLI commands and options.

## Commands

DeepScanBot is built around a set of subcommands, each with a focused purpose.

### `scan`

Crawl and analyze one or more websites, following links up to a configurable depth and producing a report.

```bash
deepscanbot scan https://example.com
deepscanbot scan https://example.com https://another.com --depth=3
deepscanbot scan --input-file=urls.txt --depth=3

# Scan with JSON output
deepscanbot scan https://example.com --json
```

See the [Scan Options](#scan-options) section below for all available options, or the [Usage Guide](/docs/guide/usage) for detailed examples.

### `version`

Show the installed version. The short flag `--version` is also supported.

```bash
deepscanbot version
deepscanbot version --json
```

### `doctor`

Verify the installation and environment. Useful for checking that everything is set up correctly.

```bash
deepscanbot doctor
deepscanbot doctor --json
```

### `config`

Manage CLI configuration. Currently configuration is provided via command-line flags; running this command reports the current state.

```bash
deepscanbot config
```

### `completion`

Generate a shell completion script for Bash, Zsh, Fish, or PowerShell. See [Shell Completion](#shell-completion) below for setup instructions.

```bash
deepscanbot completion bash
deepscanbot completion zsh
deepscanbot completion fish
deepscanbot completion powershell
```

### `help`

Show help for any command.

```bash
deepscanbot --help
deepscanbot scan --help
deepscanbot help scan
```

## Shell Completion

Install shell completion so that `deepscanbot`, its subcommands, and options auto-complete in your terminal.

### Bash

```bash
deepscanbot completion bash | sudo tee /etc/bash_completion.d/deepscanbot
```

To load it in the current shell:

```bash
source /etc/bash_completion.d/deepscanbot
```

### Zsh

```bash
deepscanbot completion zsh > "${fpath[1]}/_deepscanbot"
```

Then start a new shell (or run `compinit`) so zsh picks up the new completion script.

### Fish

```bash
deepscanbot completion fish > ~/.config/fish/completions/deepscanbot.fish
```

### PowerShell

```powershell
deepscanbot completion powershell | Out-String | Invoke-Expression
```

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
| `--output` | `"crawler_results"` | Output filename base (automatically gets .txt or .json extension) |
| `--size` | `-1` | Page size limit in KB (-1 = unlimited) |
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
> - **Global flags** (`--json`, `--no-input`, `--verbose`, `--debug`, `--quiet`, `--dry-run`) work with any command and use `--flag` syntax only
> - **Scan options** (`--force`, `--yes`, `--input-file`, `--stdin`, `--depth`, `--timeout`, etc.) only work with the `scan` command
> - Scan options accept both flag syntax (`--depth=3`) and key=value syntax (`depth=3`); a few, such as `--force` and `--yes`, are flag-only
> - `--json` is a global flag that produces JSON-formatted output for supported commands

> **Note:** For detailed examples and usage instructions, see the [Usage Guide](/docs/guide/usage).
