# DeepScanBot: Web Crawler

DeepScanBot is a customizable, feature-rich web crawler written in Go. It recursively crawls web pages, respects robots.txt, handles rate-limiting, supports retries, sitemap discovery, resume mode, and produces detailed JSON or text reports.

## Features

- **Customizable Crawl Depth**: Set the maximum depth to crawl web pages.
- **Timeout Management**: Set a timeout for each HTTP request.
- **Proxy Support**: Specify a proxy server for the HTTP requests.
- **Output Options**: Choose between plain text or JSON output with detailed reports.
- **Page Size Limit**: Skip pages exceeding a certain size.
- **Disable Redirects**: Option to disable HTTP redirects.
- **TLS Verification**: Option to disable TLS verification for HTTPS requests.
- **Unique URL Tracking**: Ensures URLs are crawled only once if enabled.
- **Show URL Source**: Display where each URL was found (e.g., in `<a>` tags, `<script>` tags).
- **CPU-aware concurrency**: By default, request concurrency uses the available Go CPU capacity; it can be overridden.
- **Content-Type filtering**: Download only configured MIME types; HTML remains the default.
- **Per-page outcomes**: Results retain crawl depth, HTTP status, and fetch errors for successful and failed pages.
- **Retry Support**: Automatic retry with exponential backoff for transient failures (timeouts, 429, 5xx).
- **Rate-Limit/Backoff Handling**: Respects `Retry-After` headers from servers like Goodreads; extra backoff for 429 responses.
- **Crawl Delay / Politeness**: Configurable delay between requests to the same host.
- **Per-Host Concurrency**: Separate concurrency limit per host, especially useful with `-cross-domain`.
- **Sitemap Support**: Discover and crawl URLs from the starting host's `/sitemap.xml`, including nested sitemaps.
- **Resume Mode**: Load existing output file and avoid recrawling URLs already present.
- **Skipped Links Output**: Skipped URLs are tracked separately with reasons (robots.txt, domain scope, duplicates, etc.).
- **Enhanced JSON Schema**: Detailed summary with status code distribution, skip reason breakdown, and retry distribution.

## Installation

```bash
# Clone the repository
git clone https://github.com/mindfiredigital/DeepScanBot.git
cd DeepScanBot

# Install dependencies
go mod download

# Build the crawler (optional)
go build -o deepscanbot ./apps/cli

# Or run directly (no build required)
go run ./apps/cli -url <starting_url> [options]
```

## Usage

### Flags

| Flag                | Type     | Default           | Description                                                               |
| ------------------- | -------- | ----------------- | ------------------------------------------------------------------------- |
| `-url`              | string   | (required)        | The starting URL for the crawler                                          |
| `-depth`            | int      | 2                 | Maximum depth to crawl                                                    |
| `-timeout`          | int      | 2                 | Timeout for each HTTP request in seconds                                  |
| `-proxy`            | string   | ""                | Proxy URL. Example: `http://127.0.0.1:8080`                               |
| `-json`             | bool     | false             | Output results in JSON format                                             |
| `-size`             | int      | -1                | Page size limit in KB (-1 = no limit)                                     |
| `-dr`               | bool     | false             | Disable following HTTP redirects                                          |
| `-s`                | bool     | false             | Show the source of the URL (e.g., href, script, img)                      |
| `-insecure`         | bool     | false             | Disable TLS verification                                                  |
| `-u`                | bool     | false             | Ensure unique URLs are crawled                                            |
| `-concurrency`      | int      | 0                 | Maximum concurrent request workers (0 = CPU cores)                        |
| `-host-concurrency` | int      | 0                 | Maximum concurrent requests per host (0 = uses -concurrency)              |
| `-content-types`    | string   | "text/html"       | Comma-separated MIME types to download. Supports wildcards like `image/*` |
| `-output`           | string   | "crawler_results" | Output filename without extension                                         |
| `-ignore-robots`    | bool     | false             | Ignore robots.txt crawl restrictions                                      |
| `-cross-domain`     | bool     | false             | Follow links to hosts other than the starting URL                         |
| `-retries`          | int      | 0                 | Number of retry attempts for transient fetch failures                     |
| `-retry-backoff`    | duration | 1s                | Base retry backoff duration (e.g., 500ms, 2s)                             |
| `-delay`            | duration | 0                 | Politeness delay between requests to the same host (e.g., 500ms)          |
| `-sitemap`          | bool     | false             | Discover and crawl URLs from the starting host's `/sitemap.xml`           |
| `-resume`           | bool     | false             | Load existing output file and avoid recrawling URLs already present       |
| `-h`                | bool     | false             | Show help message                                                         |

### Examples

#### 1. Basic Crawl

```bash
go run ./apps/cli -url https://example.com -depth 2
```

Crawls `https://example.com` up to 2 levels deep and outputs results to `crawler_results.txt`.

#### 2. JSON Output with Details

```bash
go run ./apps/cli -url https://example.com -depth 3 -json -s -u -output my_results
```

Outputs JSON to `my_results.json` with URL source tracking and deduplication.

#### 3. Crawl with Retry and Delay

```bash
go run ./apps/cli -url https://docs.example.com -depth 2 -retries 3 -retry-backoff 2s -delay 1s -host-concurrency 1
```

Retries failed requests up to 3 times with exponential backoff, waits 1 second between requests to the same host, and allows only 1 concurrent request per host.

#### 4. Cross-Domain Crawl with Sitemap

```bash
go run ./apps/cli -url https://example.com -depth 3 -cross-domain -sitemap -concurrency 10 -host-concurrency 2 -json
```

Discovers URLs from sitemap.xml, follows links to any domain, with 10 total workers and 2 per host.

#### 5. Resume an Interrupted Crawl

```bash
# First run (interrupted)
go run ./apps/cli -url https://example.com -depth 3 -json -output my_results

# Resume
go run ./apps/cli -url https://example.com -depth 3 -json -output my_results -resume
```

Loaded existing results from `my_results.json` and skips already-crawled URLs.

#### 6. Crawl Goodreads with Rate-Limit Handling

```bash
go run ./apps/cli -url https://www.goodreads.com -depth 2 -delay 2s -retries 5 -retry-backoff 2s -concurrency 2 -host-concurrency 1 -json -output goodreads_results
```

Uses 2-second politeness delay, 5 retries with exponential backoff, limited concurrency to handle Goodreads rate limits gracefully.

#### 7. Crawl PDF and Images

```bash
go run ./apps/cli -url https://example.com -depth 2 -content-types "text/html,application/pdf,image/jpeg,image/png" -json
```

Downloads HTML, PDF, JPEG, and PNG files while still parsing HTML for links.

## Output Schema

### JSON Output (`-json`)

The JSON report contains a detailed summary and two URL lists:

```json
{
  "start_url": "https://example.com",
  "output_file": "my_results.json",
  "started_at": "2026-06-23T12:00:00Z",
  "finished_at": "2026-06-23T12:00:05Z",
  "duration_ms": 5234,
  "summary": {
    "total": 45,
    "passed": 30,
    "failed": 2,
    "skipped": 8,
    "discovered": 5,
    "skipped_by_robots": 3,
    "skipped_by_domain": 2,
    "skipped_by_duplicate": 2,
    "skipped_by_content_type": 1,
    "skipped_by_depth": 0,
    "skipped_by_other": 0,
    "retried_requests": 2,
    "max_depth": 3,
    "urls_by_status_code": {
      "200": 30,
      "404": 1,
      "500": 1
    },
    "skipped_by_reason": {
      "disallowed by robots.txt": 3,
      "outside domain scope": 2,
      "duplicate": 2,
      "content type not allowed": 1
    },
    "retry_distribution": {
      "2": 1,
      "3": 1
    }
  },
  "urls": [
    {
      "url": "https://example.com",
      "source": "href",
      "depth": 0,
      "status_code": 200,
      "content_type": "text/html; charset=utf-8",
      "result": "passed",
      "attempts": 1
    },
    {
      "url": "https://example.com/about",
      "source": "href",
      "depth": 1,
      "status_code": 200,
      "content_type": "text/html",
      "result": "passed",
      "attempts": 1
    },
    {
      "url": "https://example.com/not-found",
      "source": "href",
      "depth": 2,
      "status_code": 404,
      "content_type": "text/html",
      "result": "failed",
      "error": "bad status code: 404",
      "attempts": 1
    },
    {
      "url": "https://example.com/discovered-link",
      "source": "href",
      "depth": 2,
      "result": "discovered"
    }
  ],
  "skipped": [
    {
      "url": "https://external.com/page",
      "source": "href",
      "depth": 1,
      "result": "skipped",
      "skipped_reason": "outside domain scope"
    },
    {
      "url": "https://example.com/admin",
      "source": "href",
      "depth": 1,
      "result": "skipped",
      "skipped_reason": "disallowed by robots.txt"
    }
  ]
}
```

### URLEntry Fields

| Field            | Type   | Description                                                                                                                            |
| ---------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| `url`            | string | The discovered or crawled URL                                                                                                          |
| `source`         | string | Where the URL was found: `href`, `script`, `img`, `link`, `iframe`, `form`, `sitemap`                                                  |
| `depth`          | int    | Crawl depth at which the URL was found                                                                                                 |
| `status_code`    | int    | HTTP status code (only for fetched URLs)                                                                                               |
| `content_type`   | string | MIME type from the response (only for fetched URLs)                                                                                    |
| `result`         | string | `passed` (fetched successfully), `failed` (fetch error), `skipped` (not fetched), `discovered` (found but not fetched)                 |
| `error`          | string | Error message if the fetch failed                                                                                                      |
| `skipped_reason` | string | Reason for skipping: `disallowed by robots.txt`, `outside domain scope`, `duplicate`, `content type not allowed`, `max depth exceeded` |
| `attempts`       | int    | Number of fetch attempts (1+; >1 means retries were used)                                                                              |

### Text Output (default)

Text output shows one URL per line with optional metadata in brackets:

```
[https://example.com [status=200] [result=passed]
[https://example.com/about [status=200] [result=passed]
[https://example.com/not-found [status=404] [result=failed] [error=bad status code: 404]
[https://external.com/page [result=skipped] [skipped=outside domain scope]

// With -s flag:
[href] https://example.com
```

## Architecture

The project is divided into the following packages:

### Package: `main`

The entry point. Processes command-line arguments, initializes the Crawler, and outputs results.

### Package: `crawler`

Core crawling logic: manages state, visited URLs, crawl depth, robots.txt compliance, concurrency control, retries, sitemap discovery, and rate-limit handling.

### Package: `fetcher`

HTTP fetch layer. Handles requests with proxy, TLS, redirect options, and parses `Retry-After` headers for rate-limit backoff.

### Package: `parser`

HTML parsing and link extraction using `golang.org/x/net/html`. Extracts links from `<a>`, `<script>`, `<img>`, `<link>`, `<iframe>`, and `<form>` elements.

### Package: `storage`

Result storage and report generation. Supports JSON and text output, resume file loading, and comprehensive summary statistics.

## Error Handling

The crawler handles:

- Invalid URLs or unsupported protocols
- HTTP errors (4xx, 5xx)
- Timeout and connection errors
- TLS verification errors
- Page size limits
- Robots.txt restrictions
- Rate-limiting (429) with Retry-After header support
- Content-Type mismatch

## Dependencies

- `golang.org/x/net/html` — HTML parsing
- `github.com/temoto/robotstxt` — robots.txt parsing
