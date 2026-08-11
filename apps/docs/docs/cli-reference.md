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

## Scan Options

| Option | Default | Description |
|--------|---------|-------------|
| `depth` | `2` | Maximum crawl depth |
| `timeout` | `2` | Request timeout in seconds |
| `concurrency` | `0` | Max concurrent workers (0 = CPU count) |
| `host-concurrency` | `0` | Max concurrent requests per host |
| `content-types` | `"text/html"` | Allowed MIME types |
| `output` | `"crawler_results"` | Output filename without extension |
| `json` | `false` | JSON output (scan option) |
| `size` | `-1` | Page size limit in KB |
| `proxy` | `""` | Proxy URL |
| `unique` | `false` | Unique URLs only |
| `show-source` | `false` | Show URL source |
| `disable-redirects` | `false` | Disable redirects |
| `insecure` | `false` | Disable TLS verification |
| `retries` | `0` | Retry attempts |
| `retry-backoff` | `"1s"` | Retry backoff duration |
| `delay` | `"0s"` | Politeness delay |
| `sitemap` | `false` | Discover sitemap.xml |
| `resume` | `false` | Resume mode |
| `ignore-robots` | `false` | Ignore robots.txt |
| `cross-domain` | `false` | Follow external links |

> **Note:** For detailed examples and usage instructions, see the [Usage Guide](/docs/guide/usage).
