# Usage

This guide shows how to use DeepScanBot with various examples and configurations.

## Installation

```bash
# Clone the repository
git clone https://github.com/mindfiredigital/DeepScanBot.git
cd DeepScanBot

# Install dependencies
go mod download

# Build the crawler
go build -o deepscanbot

# Or run directly
go run main.go -url <starting_url> [options]
```

## Command-Line Flags

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

## Examples

### 1. Basic Crawl

```bash
deepscanbot -url https://example.com -depth 2
```

Crawls `https://example.com` up to 2 levels deep and outputs results to `crawler_results.txt`.

### 2. JSON Output with Details

```bash
deepscanbot -url https://example.com -depth 3 -json -s -u -output my_results
```

Outputs JSON to `my_results.json` with URL source tracking and deduplication.

### 3. Crawl with Retry and Delay

```bash
deepscanbot -url https://docs.example.com -depth 2 -retries 3 -retry-backoff 2s -delay 1s -host-concurrency 1
```

Retries failed requests up to 3 times with exponential backoff, waits 1 second between requests to the same host, and allows only 1 concurrent request per host.

### 4. Cross-Domain Crawl with Sitemap

```bash
deepscanbot -url https://example.com -depth 3 -cross-domain -sitemap -concurrency 10 -host-concurrency 2 -json
```

Discovers URLs from sitemap.xml, follows links to any domain, with 10 total workers and 2 per host.

### 5. Resume an Interrupted Crawl

```bash
# First run (interrupted)
deepscanbot -url https://example.com -depth 3 -json -output my_results

# Resume
deepscanbot -url https://example.com -depth 3 -json -output my_results -resume
```

Loaded existing results from `my_results.json` and skips already-crawled URLs.

### 6. Crawl Goodreads with Rate-Limit Handling

```bash
deepscanbot -url https://www.goodreads.com -depth 2 -delay 2s -retries 5 -retry-backoff 2s -concurrency 2 -host-concurrency 1 -json -output goodreads_results
```

Uses 2-second politeness delay, 5 retries with exponential backoff, limited concurrency to handle Goodreads rate limits gracefully.

### 7. Crawl PDF and Images

```bash
deepscanbot -url https://example.com -depth 2 -content-types "text/html,application/pdf,image/jpeg,image/png" -json
```

Downloads HTML, PDF, JPEG, and PNG files while still parsing HTML for links.

### 8. Crawl Through a Proxy

```bash
deepscanbot -url https://example.com -depth 2 -proxy http://127.0.0.1:8080 -json
```

Routes all HTTP requests through a proxy server. Useful for corporate networks or anonymization.

**Proxy Types:**
- HTTP proxy: `http://proxy.example.com:8080`
- HTTPS proxy: `https://proxy.example.com:8443`
- SOCKS5 proxy: `socks5://proxy.example.com:1080`

### 9. Crawl with Custom Timeout

```bash
deepscanbot -url https://slow-api.example.com -depth 2 -timeout 30 -retries 2
```

Increases timeout to 30 seconds for slow-responding servers. Useful for APIs with high latency.

**Timeout Recommendations:**
- Fast internal services: 2-5 seconds
- Public websites: 5-10 seconds
- Slow APIs: 15-30 seconds
- Very slow services: 30-60 seconds

### 10. Disable TLS Verification

```bash
deepscanbot -url https://self-signed.example.com -depth 2 -insecure
```

Disables TLS/SSL certificate verification. **Use with caution** - only for testing with self-signed certificates.

**When to Use:**
- Development environments
- Testing with self-signed certificates
- Legacy systems with outdated certificates

**Security Note:** Never use `-insecure` in production or when crawling untrusted sites.

### 11. High-Performance Crawl

```bash
deepscanbot -url https://example.com -depth 3 -concurrency 20 -host-concurrency 5 -json -output fast_crawl
```

Maximizes performance with 20 total workers and 5 per host. Ideal for fast, internal networks.

**Performance Tips:**
- Monitor system resources (CPU, memory)
- Adjust concurrency based on target server capacity
- Use `-delay` to avoid overwhelming servers
- Consider `-host-concurrency` for multi-domain crawls

### 12. Minimal Crawl (Text Only)

```bash
deepscanbot -url https://example.com -depth 1 -size 100 -content-types "text/html"
```

Minimal crawl with depth 1, 100KB page size limit, and HTML-only content type. Fast and focused.

### 13. Crawl with No Redirects

```bash
deepscanbot -url https://example.com -dr -json
```

Disables following HTTP redirects. Useful for analyzing redirect chains or testing specific URLs.

**Use Cases:**
- Testing redirect configurations
- Analyzing redirect chains
- Crawling specific URL versions

### 14. Large-Scale Site Audit

```bash
deepscanbot -url https://example.com -depth 5 -concurrency 15 -host-concurrency 3 -delay 500ms -retries 3 -retry-backoff 1s -json -output site_audit
```

Comprehensive site audit with deep crawling, controlled concurrency, and retry logic.

**Best Practices:**
- Start with lower depth and increase gradually
- Monitor server response times
- Adjust delays based on server behavior
- Use `-resume` for very large crawls

### 15. Monitor Crawl Progress

```bash
# Start crawl in background
deepscanbot -url https://example.com -depth 3 -json -output crawl &

# Monitor output file growth
watch -n 5 'wc -l crawl.json'

# Check specific URLs
grep '"result": "failed"' crawl.json
```

Monitor long-running crawls by watching output file growth and checking for errors.

## Output Formats

### Text Output (Default)

Text output shows one URL per line with optional metadata in brackets:

```
[https://example.com [status=200] [result=passed]
[https://example.com/about [status=200] [result=passed]
[https://example.com/not-found [status=404] [result=failed] [error=bad status code: 404]
[https://external.com/page [result=skipped] [skipped=outside domain scope]

// With -s flag:
[href] https://example.com
```

### JSON Output (`-json`)

The JSON report contains a detailed summary and two URL lists. See the [Output Schema](intro#output-schema) section for details.
