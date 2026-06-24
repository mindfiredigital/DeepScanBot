# Getting Started

Welcome to DeepScanBot! This guide will help you get up and running quickly.

## What is DeepScanBot?

DeepScanBot is a powerful, feature-rich web crawler built with Go. It recursively crawls web pages, respects robots.txt, handles rate-limiting, supports retries, sitemap discovery, resume mode, and produces detailed JSON or text reports.

## Key Features

- 🤖 **Intelligent Crawling**: Respects robots.txt, handles redirects, and follows best practices
- ⚡ **High Performance**: CPU-aware concurrency with per-host rate limiting
- 🔄 **Resilient**: Automatic retries with exponential backoff and rate-limit handling
- 📊 **Detailed Reports**: JSON or text output with comprehensive statistics
- 🎯 **Flexible**: Content-type filtering, sitemap discovery, and resume mode
- 🛡️ **Production-Ready**: TLS support, proxy configuration, and comprehensive error handling

## Installation

### Prerequisites

- **Go** (version 1.21 or higher)
- **Git**

### Steps

1. **Clone the repository:**
   ```bash
   git clone https://github.com/mindfiredigital/DeepScanBot.git
   cd DeepScanBot
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the crawler:**
   ```bash
   go build -o deepscanbot
   ```

4. **Verify installation:**
   ```bash
   ./deepscanbot -h
   ```

## Quick Start

### Basic Crawl

Crawl a website up to 2 levels deep:

```bash
deepscanbot -url https://example.com -depth 2
```

### JSON Output with Details

Get detailed JSON output with URL source tracking:

```bash
deepscanbot -url https://example.com -depth 3 -json -s -u -output my_results
```

### Polite Crawl with Retries

Crawl with politeness delays and retry logic:

```bash
deepscanbot -url https://example.com -depth 2 -delay 1s -retries 3 -retry-backoff 2s
```

## Output Schema

### JSON Output

When using the `-json` flag, DeepScanBot produces a detailed JSON report:

```json
{
  "start_url": "https://example.com",
  "summary": {
    "total": 45,
    "passed": 30,
    "failed": 2,
    "skipped": 8,
    "discovered": 5
  },
  "urls": [...],
  "skipped": [...]
}
```

### Text Output (Default)

Text output shows one URL per line:

```
[https://example.com [status=200] [result=passed]
[https://example.com/about [status=200] [result=passed]
[https://example.com/not-found [status=404] [result=failed]
```

## System Requirements

### Minimum Requirements

- **OS**: Linux, macOS, or Windows (with WSL2 recommended for Windows)
- **Go**: Version 1.21 or higher
- **RAM**: 512 MB minimum, 2 GB recommended for large crawls
- **Disk**: 100 MB for application, additional space for output files
- **Network**: Stable internet connection for web crawling

### Recommended Configuration

- **Go**: Version 1.22 or higher for best performance
- **RAM**: 4 GB or more for concurrent crawling
- **CPU**: Multi-core processor for optimal concurrency
- **Network**: High-bandwidth connection for faster crawling

## Troubleshooting

### Common Issues

#### "Connection timeout" errors

Increase the timeout value:
```bash
deepscanbot -url https://example.com -timeout 10
```

#### "Too many requests" (429) errors

Add delays and retries:
```bash
deepscanbot -url https://example.com -delay 2s -retries 5 -retry-backoff 2s
```

#### "Certificate verification failed" errors

Disable TLS verification (use with caution):
```bash
deepscanbot -url https://example.com -insecure
```

#### "Blocked by robots.txt" warnings

If you need to ignore robots.txt (ensure you have permission):
```bash
deepscanbot -url https://example.com -ignore-robots
```

#### High memory usage

Reduce concurrency:
```bash
deepscanbot -url https://example.com -concurrency 2
```

### Getting Help

- 📚 Check the [Usage](usage) guide for detailed flag documentation
- 🏗️ Check [Architecture](architecture) to understand the internals
- 🛠️ See [Development Tools](development-tools) for contributing
- 🤝 See [Contributing](contributing) to contribute to the project
- 🐛 Report bugs via [GitHub Issues](https://github.com/mindfiredigital/DeepScanBot/issues)
- 💬 Join discussions in [GitHub Discussions](https://github.com/mindfiredigital/DeepScanBot/discussions)

## Next Steps

- 📖 Read the [Usage](usage) guide for detailed flag documentation
- ✨ Explore [Features](features) to learn about advanced capabilities
- 🏗️ Check [Architecture](architecture) to understand the internals
- 🛠️ Set up [Development Tools](development-tools) for contributing
- 🤝 See [Contributing](contributing) to contribute to the project