# DeepScanBot: Web Crawler

DeepScanBot is a customizable web crawler written in Go.

## Overview

DeepScanBot allows you to crawl websites with various configurations, including crawl depth, timeout settings, proxy support, and output options.

## Features

- **Customizable Crawl Depth**: Set the maximum depth to crawl web pages.
- **Timeout Management**: Set a timeout for each HTTP request.
- **Proxy Support**: Specify a proxy server for the HTTP requests.
- **Output Options**: Choose between plain text or JSON output.
- **Page Size Limit**: Skip pages exceeding a certain size.
- **Disable Redirects**: Option to disable HTTP redirects.
- **TLS Verification**: Option to disable TLS verification for HTTPS requests.
- **Unique URL Tracking**: Ensures URLs are crawled only once if enabled.
- **Show URL Source**: Display where each URL was found (e.g., in `<a>` tags, `<script>` tags).
- **Concurrency control**: Limit maximum concurrent requests to avoid overloading target servers.
- **Content-Type filtering**: Download only configured MIME types; HTML remains the default.

## Usage

To run the web crawler, use the following commands:

### Install Dependencies

```bash
go mod download

# Run the Crawler

go run main.go -url <starting_url> [options]

# Build the Crawler

go build

# Flags
-url <string>: Required. The starting URL for the crawler.
-depth <int>: Maximum depth to crawl. Default: 2.
-timeout <int>: Timeout for each HTTP request in seconds. Default: 2.
-proxy <string>: Proxy URL for HTTP requests. Example: http://127.0.0.1:8080.
-json: Output results in JSON format. Default: false.
-size <int>: Limit page size in KB. Default: -1 (no limit).
-dr: Disable following HTTP redirects. Default: false.
-s: Show the source of the URL based on where it was found. Default: false.
-insecure: Disable TLS verification. Default: false.
-u: Ensure unique URLs are crawled. Default: false.
-concurrency <int>: Limit maximum concurrent request workers. Default: 10.
-content-types <string>: Comma-separated MIME types to download. Default: text/html. Supports wildcards such as image/*.
-output <string>: Output filename without an extension. Default: crawler_results.
-h: Show help message.

# Example
To start crawling from https://example.com with a maximum depth of 3, run:

go run main.go -url https://example.com -depth 3
```
