# Features

DeepScanBot is a powerful, feature-rich web crawler built with Go. Here are its key features:

## Core Features

### Customizable Crawl Depth

Control how deep the crawler goes into a website's link structure. Set the maximum depth to crawl web pages, preventing infinite crawls and focusing on relevant content.

**Example:**
```bash
deepscanbot -url https://example.com -depth 3
```

**Use Cases:**
- Shallow crawls (depth 1-2) for link discovery
- Medium crawls (depth 3-4) for content mapping
- Deep crawls (depth 5+) for comprehensive site analysis

### Timeout Management

Set a timeout for each HTTP request to prevent hanging on slow or unresponsive servers. Default is 2 seconds, configurable based on your needs.

**Example:**
```bash
deepscanbot -url https://example.com -timeout 10
```

**Use Cases:**
- Fast internal networks: 2-5 seconds
- Public websites: 5-10 seconds
- Slow APIs or legacy systems: 15-30 seconds

### Proxy Support

Route HTTP requests through a proxy server for anonymity, load balancing, or accessing geo-restricted content.

**Example:**
```bash
deepscanbot -url https://example.com -proxy http://127.0.0.1:8080
```

**Use Cases:**
- Corporate networks requiring proxy access
- Load testing with proxy servers
- Geographic content testing

### Output Options

Choose between plain text or JSON output with detailed reports. JSON output includes comprehensive statistics, error details, and metadata.

**Example:**
```bash
deepscanbot -url https://example.com -json -output results
```

**Output Files:**
- Text: `results.txt` (human-readable)
- JSON: `results.json` (machine-readable with full schema)

### Page Size Limit

Skip pages exceeding a certain size to avoid downloading large files unnecessarily. Set in kilobytes, with -1 for no limit.

**Example:**
```bash
deepscanbot -url https://example.com -size 1024
```

**Use Cases:**
- Skip large media files
- Focus on text content
- Reduce bandwidth usage

### Disable Redirects

Option to disable following HTTP redirects, useful for analyzing redirect chains or testing specific URLs.

**Example:**
```bash
deepscanbot -url https://example.com -dr
```

**Use Cases:**
- Testing redirect configurations
- Analyzing redirect chains
- Crawling specific URL versions

### TLS Verification

Option to disable TLS/SSL certificate verification for HTTPS requests. Use with caution in production environments.

**Example:**
```bash
deepscanbot -url https://example.com -insecure
```

**Use Cases:**
- Testing with self-signed certificates
- Development environments
- Legacy systems with outdated certificates

### Unique URL Tracking

Ensures each URL is crawled only once, preventing duplicate requests and reducing crawl time.

**Example:**
```bash
deepscanbot -url https://example.com -u
```

**Benefits:**
- Faster crawls
- Reduced server load
- More accurate statistics

### Show URL Source

Display where each URL was found (e.g., in `<a>` tags, `<script>` tags, `<img>` tags). Useful for understanding link structure.

**Example:**
```bash
deepscanbot -url https://example.com -s
```

**Output Example:**
```
[href] https://example.com/about
[script] https://example.com/app.js
[img] https://example.com/logo.png
```

## Advanced Features

### CPU-Aware Concurrency

By default, DeepScanBot uses the available Go CPU capacity for concurrent requests. This can be overridden for specific use cases.

**Example:**
```bash
# Use 10 concurrent workers
deepscanbot -url https://example.com -concurrency 10

# Use default (CPU cores)
deepscanbot -url https://example.com
```

**Benefits:**
- Optimal resource utilization
- Prevents system overload
- Configurable for different environments

### Content-Type Filtering

Download only configured MIME types. HTML remains the default, but you can specify additional types like PDFs, images, or APIs.

**Example:**
```bash
# Download HTML, PDFs, and images
deepscanbot -url https://example.com -content-types "text/html,application/pdf,image/jpeg,image/png"
```

**Supported Types:**
- `text/html` (default)
- `application/pdf`
- `image/*` (wildcards supported)
- `application/json`
- Custom MIME types

**Use Cases:**
- Document crawling (PDFs, DOCX)
- Image harvesting
- API endpoint discovery

### Per-Page Outcomes

Results retain crawl depth, HTTP status, and fetch errors for both successful and failed pages. This provides complete visibility into crawl results.

**Example Output:**
```json
{
  "url": "https://example.com/page",
  "depth": 2,
  "status_code": 200,
  "result": "passed",
  "attempts": 1
}
```

**Benefits:**
- Complete audit trail
- Error analysis
- Performance metrics

### Retry Support

Automatic retry with exponential backoff for transient failures including timeouts, 429 (Too Many Requests), and 5xx server errors.

**Example:**
```bash
deepscanbot -url https://example.com -retries 3 -retry-backoff 2s
```

**Retry Strategy:**
- Exponential backoff: 2s, 4s, 8s...
- Respects `Retry-After` headers
- Maximum attempts configurable
- Only retries on transient errors

**Use Cases:**
- Unstable networks
- Rate-limited APIs
- Temporary server issues

### Rate-Limit/Backoff Handling

Intelligently respects `Retry-After` headers from servers like Goodreads, with extra backoff for 429 responses.

**Example:**
```bash
deepscanbot -url https://www.goodreads.com -delay 2s -retries 5 -retry-backoff 2s
```

**Features:**
- Parses `Retry-After` headers
- Automatic backoff adjustment
- Polite crawling behavior
- Prevents IP bans

### Crawl Delay / Politeness

Configurable delay between requests to the same host to avoid overwhelming servers.

**Example:**
```bash
# Wait 1 second between requests to the same host
deepscanbot -url https://example.com -delay 1s

# Wait 2 seconds (more polite)
deepscanbot -url https://example.com -delay 2s
```

**Best Practices:**
- 1-2 seconds for most sites
- 3-5 seconds for rate-sensitive sites
- Respect server resources

### Per-Host Concurrency

Separate concurrency limit per host, especially useful when crawling multiple domains with `-cross-domain`.

**Example:**
```bash
# 10 total workers, max 2 per host
deepscanbot -url https://example.com -concurrency 10 -host-concurrency 2 -cross-domain
```

**Benefits:**
- Prevents overwhelming single domains
- Enables efficient multi-domain crawling
- Fine-grained control

### Sitemap Support

Discover and crawl URLs from the starting host's `/sitemap.xml`, including nested sitemaps and sitemap indexes.

**Example:**
```bash
deepscanbot -url https://example.com -sitemap
```

**Features:**
- Automatic sitemap.xml discovery
- Nested sitemap support
- Sitemap index parsing
- Priority and changefreq awareness

**Use Cases:**
- Complete site discovery
- SEO analysis
- Content inventory

### Resume Mode

Load existing output file and avoid recrawling URLs already present. Perfect for large crawls that get interrupted.

**Example:**
```bash
# First run (interrupted)
deepscanbot -url https://example.com -depth 5 -json -output crawl_results

# Resume from where you left off
deepscanbot -url https://example.com -depth 5 -json -output crawl_results -resume
```

**How It Works:**
1. Loads existing output file
2. Populates visited URLs set
3. Skips already-crawled URLs
4. Continues from remaining queue

**Benefits:**
- Saves time on large crawls
- Reduces server load
- Enables incremental crawling

### Skipped Links Output

Skipped URLs are tracked separately with detailed reasons (robots.txt, domain scope, duplicates, content type, depth limits, etc.).

**Example Output:**
```json
{
  "url": "https://external.com/page",
  "result": "skipped",
  "skipped_reason": "outside domain scope"
}
```

**Skip Reasons:**
- `skipped_by_robots` - Blocked by robots.txt
- `skipped_by_domain` - Outside domain scope
- `skipped_by_duplicate` - Duplicate URL
- `skipped_by_content_type` - Wrong MIME type
- `skipped_by_depth` - Exceeded max depth
- `skipped_by_other` - Other reasons

### Enhanced JSON Schema

Detailed summary with status code distribution, skip reason breakdown, and retry distribution for comprehensive analysis.

**Example Summary:**
```json
{
  "summary": {
    "total": 150,
    "passed": 120,
    "failed": 15,
    "skipped": 10,
    "discovered": 5,
    "skipped_by_robots": 3,
    "skipped_by_domain": 4,
    "urls_by_status_code": {
      "200": 115,
      "404": 3,
      "500": 2
    },
    "retry_distribution": {
      "1": 8,
      "2": 5,
      "3": 2
    }
  }
}
```

**Benefits:**
- Comprehensive analytics
- Error pattern identification
- Performance insights

## Feature Comparison

| Feature | DeepScanBot | Basic Crawlers |
|---------|-------------|----------------|
| robots.txt compliance | ✅ | ❌ |
| Retry with backoff | ✅ | ❌ |
| Sitemap discovery | ✅ | ❌ |
| Resume mode | ✅ | ❌ |
| Rate-limit handling | ✅ | ❌ |
| Per-host concurrency | ✅ | ❌ |
| Content-type filtering | ✅ | ❌ |
| Detailed JSON reports | ✅ | ❌ |
| Politeness delays | ✅ | ❌ |