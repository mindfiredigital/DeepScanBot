# Architecture

This document describes the architecture and design of DeepScanBot.

## Overview

DeepScanBot follows a modular architecture with clear separation of concerns. The project is organized into several packages, each responsible for a specific aspect of the crawling process.

## Package Structure

### Package: `main`

The entry point of the application. It processes command-line arguments, initializes the Crawler instance, and starts the crawling process.

**File:** `main.go`

**Responsibilities:**
- Parse CLI flags and arguments using Go's `flag` package
- Initialize crawler configuration with user-provided options
- Start the crawling process
- Output results to console or files
- Handle graceful shutdown and error reporting

**Key Functions:**
- `main()` - Application entry point
- Flag parsing and validation
- Crawler initialization
- Result display and file output

### Package: `crawler`

The core crawling logic. Manages the state of the crawling process, including visited URLs, crawl depth, and other configurations.

**File:** `crawler/crawler.go`

**Responsibilities:**
- Manage crawl state and visited URLs using thread-safe data structures
- Coordinate between fetcher and parser packages
- Enforce robots.txt compliance via robots.go
- Control concurrency and rate limiting with semaphores
- Handle retries and backoff via retry.go
- Discover sitemaps via sitemap.go
- Support resume mode by loading existing results

**Key Components:**
- `Crawler` struct: Main crawler state containing visited URLs, results queue, and configuration
- `Options` struct: Configuration options for crawl behavior
- `crawl()` method: Main orchestration loop
- `processURL()` method: Individual URL processing logic
- URL deduplication using map-based visited set
- Depth tracking with priority queue

**Data Structures:**
```go
type Crawler struct {
    visited     map[string]bool      // Visited URLs
    results     []types.URLEntry     // Crawl results
    skipped     []types.URLEntry     // Skipped URLs
    mu          sync.Mutex            // Thread safety
    wg          sync.WaitGroup        // Goroutine coordination
    sem         chan struct{}         // Concurrency limiter
    hostLimits  map[string]chan struct{} // Per-host limits
    options     Options               // Configuration
}

type Options struct {
    URL              string        // Starting URL
    Depth            int           // Max crawl depth
    Concurrency      int           // Max workers
    HostConcurrency  int           // Max per host
    Delay            time.Duration // Politeness delay
    // ... other options
}
```

### Package: `fetcher`

HTTP fetch layer. Handles all HTTP requests to fetch web page content.

**File:** `fetcher/fetcher.go`

**Responsibilities:**
- Execute HTTP GET requests with Go's `net/http` client
- Handle proxy configuration via `http.ProxyURL`
- Manage TLS/SSL settings with custom `tls.Config`
- Process redirects with custom `CheckRedirect` function
- Parse `Retry-After` headers for rate-limit handling
- Enforce page size limits by reading response body incrementally
- Filter by content type using `Content-Type` header
- Track request metrics (status code, size, duration, etc.)

**Key Features:**
- Configurable timeouts with `http.Client.Timeout`
- Proxy support for HTTP/HTTPS/SOCKS5
- TLS verification toggle with `InsecureSkipVerify`
- Redirect handling (follow, limit, or disable)
- Content-Type filtering with wildcard support
- Retry with exponential backoff via retry.go
- Rate-limit awareness with `Retry-After` header parsing

**Implementation Details:**
```go
type Fetcher struct {
    client      *http.Client       // HTTP client
    timeout     time.Duration      // Request timeout
    proxy       *url.URL           // Proxy URL
    insecure    bool               // Skip TLS verification
    maxSize     int64              // Max page size in bytes
    retries     int                // Max retry attempts
    retryBackoff time.Duration     // Base backoff duration
    userAgent   string             // User-Agent header
}

func (f *Fetcher) Fetch(url string) (*http.Response, error) {
    // 1. Create request with context and timeout
    // 2. Apply proxy if configured
    // 3. Execute request with retry logic
    // 4. Check content type
    // 5. Enforce size limits
    // 6. Parse Retry-After if needed
    // 7. Return response or error
}
```

**Error Handling:**
- Timeout errors trigger retry with backoff
- 429 responses respect `Retry-After` header
- 5xx errors trigger exponential backoff retry
- Network errors are retried with increasing delays

### Package: `parser`

HTML parsing and link extraction.

**File:** `parser/parser.go`

**Responsibilities:**
- Parse HTML content using `golang.org/x/net/html` tokenizer
- Extract links from various HTML elements
- Track URL sources (where links were found) for source reporting
- Handle relative and absolute URLs with `url.Parse` and `url.ResolveReference`
- Filter by content type before parsing

**Supported Elements:**
- `<a>` tags (href attribute)
- `<script>` tags (src attribute)
- `<img>` tags (src attribute)
- `<link>` tags (href attribute)
- `<iframe>` tags (src attribute)
- `<form>` tags (action attribute)
- `<source>` tags (src attribute)
- `<video>` tags (src attribute)
- `<audio>` tags (src attribute)

**Implementation Details:**
```go
type Parser struct {
    contentTypes []string           // Allowed content types
    baseURL      *url.URL           // Base URL for resolving relative links
}

func (p *Parser) Parse(htmlContent []byte, sourceURL string) ([]types.URLEntry, error) {
    // 1. Tokenize HTML using html.NewTokenizer
    // 2. Iterate through tokens
    // 3. Extract URLs from supported elements
    // 4. Resolve relative URLs to absolute
    // 5. Filter by content type
    // 6. Return list of discovered URLs with source info
}
```

**URL Resolution:**
- Converts relative URLs (e.g., `/about`) to absolute (e.g., `https://example.com/about`)
- Handles protocol-relative URLs (e.g., `//cdn.example.com`)
- Removes fragments and normalizes URLs
- Filters out javascript: and data: URLs

### Package: `storage`

Result storage and report generation.

**Files:** `storage/storage.go`, `storage/io.go`, `storage/report.go`

**Responsibilities:**
- Store crawl results in memory using thread-safe slices
- Write results to files in text or JSON format
- Load existing results for resume mode
- Generate summary statistics with detailed metrics
- Track skipped URLs with categorized reasons

**Output Formats:**
- **Plain Text** (human-readable, one URL per line)
- **JSON** (machine-readable with full schema and statistics)

**Implementation Details:**

**Storage (storage.go):**
```go
type Storage struct {
    results []types.URLEntry      // Crawled URLs
    skipped []types.URLEntry      // Skipped URLs
    mu      sync.RWMutex          // Read/write lock
}

func (s *Storage) AddResult(entry types.URLEntry)
func (s *Storage) AddSkipped(entry types.URLEntry)
func (s *Storage) GetResults() []types.URLEntry
func (s *Storage) GetSkipped() []types.URLEntry
```

**I/O (io.go):**
```go
func SaveToText(filename string, results []types.URLEntry, summary types.Summary) error
func SaveToJSON(filename string, results []types.URLEntry, skipped []types.URLEntry, summary types.Summary) error
func LoadFromJSON(filename string) ([]types.URLEntry, []types.URLEntry, error)
```

**Report Generation (report.go):**
```go
func GenerateSummary(results []types.URLEntry, skipped []types.URLEntry) types.Summary
func CalculateStats(results []types.URLEntry, skipped []types.URLEntry) types.Summary
```

**Summary Statistics:**
- Total URLs processed
- Passed/failed/skipped counts
- Status code distribution
- Skip reason breakdown
- Retry distribution
- Maximum depth reached

### Package: `logger`

Logging utilities.

**File:** `logger/logger.go`

**Responsibilities:**
- Provide structured logging with consistent formatting
- Support different log levels (Info, Warn, Error, Debug)
- Format log messages with timestamps and context
- Thread-safe logging for concurrent operations

**Implementation Details:**
```go
type Logger struct {
    level   logLevel
    mu      sync.Mutex
    output  io.Writer
}

func (l *Logger) Info(format string, args ...interface{})
func (l *Logger) Warn(format string, args ...interface{})
func (l *Logger) Error(format string, args ...interface{})
func (l *Logger) Debug(format string, args ...interface{})
```

**Log Levels:**
- **Info**: General information (crawl progress, URL discovery)
- **Warn**: Warnings (retries, skipped URLs)
- **Error**: Errors (fetch failures, parse errors)
- **Debug**: Detailed debugging (HTTP headers, timing)

**Example Output:**
```
[INFO] 2024-01-15T10:30:45Z Starting crawl of https://example.com
[INFO] 2024-01-15T10:30:45Z Discovered 15 URLs from https://example.com
[WARN] 2024-01-15T10:30:46Z Retrying https://example.com/slow (attempt 2/3)
[ERROR] 2024-01-15T10:30:47Z Failed to fetch https://example.com/404: Not Found
[INFO] 2024-01-15T10:31:00Z Crawl completed: 45 total, 30 passed, 2 failed, 8 skipped
```

### Package: `types`

Type definitions and structures.

**File:** `types/types.go`

**Responsibilities:**
- Define core data structures shared across all packages
- Provide type definitions for URLs, results, and configuration
- Ensure type safety with strong typing
- Define constants for result states and skip reasons

**Core Types:**

**URLEntry:**
```go
type URLEntry struct {
    URL           string    `json:"url"`
    Source        string    `json:"source"`        // href, script, img, etc.
    Depth         int       `json:"depth"`
    StatusCode    int       `json:"status_code"`   // 0 if not fetched
    ContentType   string    `json:"content_type"`  // empty if not fetched
    Result        string    `json:"result"`        // passed, failed, skipped, discovered
    Error         string    `json:"error"`         // error message if failed
    SkippedReason string    `json:"skipped_reason"` // reason if skipped
    Attempts      int       `json:"attempts"`      // number of fetch attempts
}
```

**Summary:**
```go
type Summary struct {
    Total               int            `json:"total"`
    Passed              int            `json:"passed"`
    Failed              int            `json:"failed"`
    Skipped             int            `json:"skipped"`
    Discovered          int            `json:"discovered"`
    SkippedByRobots     int            `json:"skipped_by_robots"`
    SkippedByDomain     int            `json:"skipped_by_domain"`
    SkippedByDuplicate  int            `json:"skipped_by_duplicate"`
    SkippedByContentType int           `json:"skipped_by_content_type"`
    SkippedByDepth      int            `json:"skipped_by_depth"`
    SkippedByOther      int            `json:"skipped_by_other"`
    RetriedRequests     int            `json:"retried_requests"`
    MaxDepth            int            `json:"max_depth"`
    URLsByStatusCode    map[string]int `json:"urls_by_status_code"`
    SkippedByReason     map[string]int `json:"skipped_by_reason"`
    RetryDistribution   map[int]int    `json:"retry_distribution"`
}
```

**Constants:**
```go
const (
    ResultPassed     = "passed"
    ResultFailed     = "failed"
    ResultSkipped    = "skipped"
    ResultDiscovered = "discovered"
)

const (
    SkipReasonRobots        = "skipped_by_robots"
    SkipReasonDomain        = "skipped_by_domain"
    SkipReasonDuplicate     = "skipped_by_duplicate"
    SkipReasonContentType   = "skipped_by_content_type"
    SkipReasonDepth         = "skipped_by_depth"
    SkipReasonOther         = "skipped_by_other"
)
```

## Data Flow

### High-Level Flow

```
┌─────────────┐
│   main.go   │
│  (CLI Args) │
└──────┬──────┘
       │
       │ Initialize
       ▼
┌─────────────┐
│   crawler   │◄───── Options (config)
│  (Orchestrate)│
└──────┬──────┘
       │
       │ Concurrent Processing
       ├──────────────────┐
       ▼                  ▼
┌─────────────┐    ┌─────────────┐
│   fetcher   │    │   parser    │
│  (HTTP GET) │    │ (HTML Parse) │
└──────┬──────┘    └──────┬──────┘
       │                  │
       │            ┌──────┴──────┐
       │            │   storage   │
       │            │  (Results)  │
       │            └──────┬──────┘
       │                   │
       └───────────────────┘
                         │
                         │ Write Output
                         ▼
                  ┌─────────────┐
                  │   Output    │
                  │ (JSON/Text) │
                  └─────────────┘
```

### Detailed Crawl Flow

```
1. Initialization
   ├─ Parse CLI flags
   ├─ Create Options struct
   ├─ Initialize Crawler with options
   └─ Load existing results if -resume enabled

2. Seed URL Processing
   ├─ Add starting URL to queue
   ├─ Mark as discovered
   └─ Begin crawl loop

3. Crawl Loop (for each URL)
   ├─ Dequeue URL from priority queue (by depth)
   ├─ Check if already visited (deduplication)
   ├─ Check depth limit
   ├─ Check domain scope (-cross-domain)
   │
   ├─ robots.txt Check
   │  ├─ Fetch robots.txt for domain
   │  ├─ Parse with robotstxt library
   │  └─ Allow/deny based on rules
   │
   ├─ Fetch Page (fetcher package)
   │  ├─ Create HTTP request with timeout
   │  ├─ Apply proxy if configured
   │  ├─ Execute with retry logic
   │  │  ├─ Attempt 1: Direct request
   │  │  ├─ On failure: Wait (backoff)
   │  │  ├─ Attempt 2: Retry
   │  │  └─ ... up to max retries
   │  ├─ Check content type
   │  ├─ Enforce size limit
   │  └─ Return response or error
   │
   ├─ Parse HTML (parser package)
   │  ├─ Tokenize HTML content
   │  ├─ Extract links from elements
   │  ├─ Resolve relative URLs
   │  └─ Return discovered URLs
   │
   ├─ Store Results (storage package)
   │  ├─ Add URL entry with status
   │  ├─ Track errors if failed
   │  └─ Update summary statistics
   │
   └─ Enqueue Discovered URLs
      ├─ Filter by depth
      ├─ Deduplicate
      └─ Add to queue for processing

4. Completion
   ├─ Wait for all goroutines to finish
   ├─ Generate final summary
   ├─ Write output file (JSON or text)
   └─ Display summary to user
```

### Concurrent Processing Flow

```
Main Goroutine
│
├─ Create worker pool (N goroutines)
│  │
│  ├─ Worker 1: fetch → parse → store → repeat
│  ├─ Worker 2: fetch → parse → store → repeat
│  ├─ Worker 3: fetch → parse → store → repeat
│  └─ ...
│
├─ URL Queue (thread-safe)
│  ├─ Priority queue sorted by depth
│  └─ Mutex-protected access
│
├─ Shared State (thread-safe)
│  ├─ Visited URLs map (sync.Map or mutex)
│  ├─ Results slice (mutex)
│  └─ Skipped URLs slice (mutex)
│
└─ Semaphores
   ├─ Global concurrency limiter
   └─ Per-host concurrency limiters
```

## Concurrency Model

DeepScanBot uses Go's concurrency primitives for efficient, polite crawling:

### Core Concurrency Patterns

- **Worker Pool Pattern**: Multiple goroutines fetch pages concurrently
- **Semaphore-based Limiting**: Controls maximum concurrent requests using channels
- **Per-host Limiting**: Separate limits for each domain to prevent overwhelming single servers
- **Thread-safe Storage**: Mutex-protected shared state for concurrent access

### Concurrency Controls

1. **Global Concurrency** (`-concurrency`): Total worker count (default: CPU cores)
   - Controls overall system resource usage
   - Prevents overwhelming the local system
   - Default: `runtime.NumCPU()`

2. **Per-host Concurrency** (`-host-concurrency`): Limit per domain
   - Prevents overwhelming individual servers
   - Essential for polite crawling
   - Default: Same as global concurrency

3. **Politeness Delay** (`-delay`): Wait time between requests to same host
   - Ensures respectful crawling behavior
   - Prevents IP bans
   - Default: 0 (no delay)

### Implementation Details

**Worker Pool:**
```go
func (c *Crawler) startWorkers() {
    for i := 0; i < c.options.Concurrency; i++ {
        go c.worker()
    }
}

func (c *Crawler) worker() {
    for url := range c.queue {
        c.processURL(url)
    }
}
```

**Semaphore-based Limiting:**
```go
// Global concurrency limiter
c.sem = make(chan struct{}, c.options.Concurrency)

// Per-host limiter
c.hostLimits = make(map[string]chan struct{})

func (c *Crawler) acquireHostPermit(host string) {
    if _, exists := c.hostLimits[host]; !exists {
        c.hostLimits[host] = make(chan struct{}, c.options.HostConcurrency)
    }
    c.hostLimits[host] <- struct{}{}
}
```

**Thread-safe Storage:**
```go
func (c *Crawler) addResult(entry types.URLEntry) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.results = append(c.results, entry)
}
```

### Performance Considerations

- **Optimal Concurrency**: Balance between speed and politeness
- **Memory Management**: Limit queue size to prevent memory exhaustion
- **CPU Utilization**: Match concurrency to available CPU cores
- **Network I/O**: Concurrent requests maximize bandwidth utilization

## Error Handling

DeepScanBot implements comprehensive error handling to ensure resilient crawling:

### Error Handling Strategies

- **Retry Logic**: Exponential backoff for transient failures
- **Rate-limit Awareness**: Respects `Retry-After` headers
- **Graceful Degradation**: Continues crawling despite individual failures
- **Detailed Reporting**: Tracks errors per URL in output
- **Context Cancellation**: Supports graceful shutdown via context

### Retry Strategy

**Retryable Errors:**
- Network timeouts
- 429 (Too Many Requests)
- 5xx server errors (500, 502, 503, 504)
- Connection resets
- Temporary network failures

**Non-retryable Errors:**
- 4xx client errors (except 429)
- DNS resolution failures
- Invalid URLs
- SSL/TLS errors (unless -insecure)

**Backoff Algorithm:**
```
Attempt 1: Immediate
Attempt 2: base_duration * 2^1 = 2s
Attempt 3: base_duration * 2^2 = 4s
Attempt 4: base_duration * 2^3 = 8s
...
```

**Configuration:**
- `-retries`: Maximum retry attempts (default: 0)
- `-retry-backoff`: Base backoff duration (default: 1s)

**Example:**
```bash
deepscanbot -url https://example.com -retries 3 -retry-backoff 2s
```
Results in retry delays: 2s, 4s, 8s

### Rate-Limit Handling

**Retry-After Header Support:**
- Parses `Retry-After` header (seconds or HTTP-date)
- Waits specified duration before retry
- Falls back to exponential backoff if header missing

**429 Response Handling:**
- Extra backoff multiplier for 429 responses
- Prevents immediate retry after rate limit
- Respects server-imposed limits

**Example Flow:**
```
1. Request → 429 Too Many Requests
2. Parse Retry-After: 120 seconds
3. Wait 120 seconds
4. Retry request
5. Success or next retry
```

### Error Reporting

**In JSON Output:**
```json
{
  "url": "https://example.com/slow",
  "result": "failed",
  "error": "fetch failed after 3 attempts: timeout",
  "attempts": 3
}
```

**In Text Output:**
```
[https://example.com/slow [status=0] [result=failed] [error=timeout after 3 attempts]
```

**Skip Reasons:**
- `skipped_by_robots` - Blocked by robots.txt
- `skipped_by_domain` - Outside domain scope
- `skipped_by_duplicate` - Duplicate URL
- `skipped_by_content_type` - Wrong MIME type
- `skipped_by_depth` - Exceeded max depth
- `skipped_by_other` - Other reasons

### Graceful Degradation

- Individual URL failures don't stop the crawl
- Errors are logged and tracked
- Crawl continues with remaining URLs
- Final report includes all errors for analysis

## State Management

### In-Memory State

DeepScanBot maintains several in-memory data structures during crawling:

**Visited URLs:**
```go
visited map[string]bool
```
- Set of already-crawled URLs
- O(1) lookup for duplicate detection
- Thread-safe access with mutex

**URL Queue:**
```go
queue []string  // or priority queue
```
- Priority queue for pending URLs (sorted by depth)
- FIFO within same depth level
- Thread-safe access with mutex

**Results:**
```go
results []types.URLEntry
```
- Slice of successfully crawled URLs
- Includes status, errors, metadata
- Thread-safe append with mutex

**Skipped URLs:**
```go
skipped []types.URLEntry
```
- Separate list for skipped URLs
- Includes skip reason for each URL
- Thread-safe append with mutex

**Per-host Limits:**
```go
hostLimits map[string]chan struct{}
```
- Semaphore channels for each domain
- Enforces per-host concurrency limits
- Dynamically created as new hosts are encountered

### State Persistence

**In-Memory Only:**
- All state is maintained in memory during crawl
- No intermediate writes to disk
- Final output written on completion

**Memory Optimization:**
- Results stored as compact structs
- URLs normalized to reduce memory usage
- Large crawls may require increased RAM

### Resume Mode

When `-resume` is enabled, DeepScanBot can continue interrupted crawls:

**How It Works:**

1. **Load Existing Output:**
   ```go
   results, skipped, err := storage.LoadFromJSON("crawl_results.json")
   ```
   - Reads existing JSON output file
   - Parses URLs and metadata
   - Populates visited URLs set

2. **Populate Visited Set:**
   ```go
   for _, entry := range results {
       c.visited[entry.URL] = true
   }
   for _, entry := range skipped {
       c.visited[entry.URL] = true
   }
   ```
   - Marks all previous URLs as visited
   - Prevents recrawling completed URLs

3. **Skip Existing URLs:**
   ```go
   if c.visited[discoveredURL] {
       continue  // Skip already crawled
   }
   ```
   - Checks each discovered URL
   - Skips if already in results

4. **Continue Crawling:**
   - Processes remaining URLs in queue
   - Appends new results to existing data
   - Overwrites output file on completion

**Resume Requirements:**
- Original output file must exist
- File must be valid JSON format
- Starting URL should be the same
- Depth and options can be adjusted

**Benefits:**
- Saves time on large crawls
- Reduces server load
- Enables incremental crawling
- Recovers from interruptions

## Output Schema

### URLEntry Structure

The `URLEntry` struct represents a single URL in the crawl results:

```go
type URLEntry struct {
    URL           string    `json:"url"`            // The URL that was crawled
    Source        string    `json:"source"`         // Where URL was found (href, script, img, etc.)
    Depth         int       `json:"depth"`          // Crawl depth at which URL was discovered
    StatusCode    int       `json:"status_code"`    // HTTP status code (0 if not fetched)
    ContentType   string    `json:"content_type"`   // MIME type of response (empty if not fetched)
    Result        string    `json:"result"`         // Outcome: passed, failed, skipped, discovered
    Error         string    `json:"error"`          // Error message if fetch/parse failed
    SkippedReason string    `json:"skipped_reason"` // Reason if URL was skipped
    Attempts      int       `json:"attempts"`       // Number of fetch attempts made
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `url` | string | The complete URL that was processed |
| `source` | string | HTML element where link was found (href, script, img, etc.) |
| `depth` | int | Depth level in crawl hierarchy (0 = starting URL) |
| `status_code` | int | HTTP status code (200, 404, etc.), 0 if not fetched |
| `content_type` | string | MIME type from Content-Type header |
| `result` | string | One of: `passed`, `failed`, `skipped`, `discovered` |
| `error` | string | Error message if result is `failed` |
| `skipped_reason` | string | Category explaining why URL was skipped |
| `attempts` | int | Number of times fetch was attempted |

**Result Values:**
- `passed` - Successfully fetched and parsed
- `failed` - Failed to fetch or parse after retries
- `skipped` - Skipped due to robots.txt, domain, duplicate, etc.
- `discovered` - Found but not yet processed

### Summary Structure

The `Summary` struct provides aggregate statistics for the entire crawl:

```go
type Summary struct {
    Total               int            `json:"total"`                // Total URLs processed
    Passed              int            `json:"passed"`               // Successfully crawled
    Failed              int            `json:"failed"`               // Failed after retries
    Skipped             int            `json:"skipped"`              // Skipped for various reasons
    Discovered          int            `json:"discovered"`           // Found but not crawled
    SkippedByRobots     int            `json:"skipped_by_robots"`    // Blocked by robots.txt
    SkippedByDomain     int            `json:"skipped_by_domain"`    // Outside domain scope
    SkippedByDuplicate  int            `json:"skipped_by_duplicate"` // Duplicate URL
    SkippedByContentType int           `json:"skipped_by_content_type"` // Wrong MIME type
    SkippedByDepth      int            `json:"skipped_by_depth"`     // Exceeded max depth
    SkippedByOther      int            `json:"skipped_by_other"`     // Other reasons
    RetriedRequests     int            `json:"retried_requests"`     // Total retry attempts
    MaxDepth            int            `json:"max_depth"`            // Deepest crawled URL
    URLsByStatusCode    map[string]int `json:"urls_by_status_code"`  // Status code distribution
    SkippedByReason     map[string]int `json:"skipped_by_reason"`    // Skip reason distribution
    RetryDistribution   map[int]int    `json:"retry_distribution"`   // Retry count distribution
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `total` | int | Total number of URLs processed (passed + failed + skipped) |
| `passed` | int | Successfully fetched and parsed URLs |
| `failed` | int | URLs that failed after all retry attempts |
| `skipped` | int | URLs skipped without attempting fetch |
| `discovered` | int | URLs found but not yet processed |
| `skipped_by_robots` | int | Blocked by robots.txt rules |
| `skipped_by_domain` | int | Outside domain scope (without -cross-domain) |
| `skipped_by_duplicate` | int | Duplicate URL already in results |
| `skipped_by_content_type` | int | MIME type not in allowed list |
| `skipped_by_depth` | int | Exceeded maximum crawl depth |
| `skipped_by_other` | int | Other skip reasons |
| `retried_requests` | int | Total number of retry attempts across all URLs |
| `max_depth` | int | Maximum depth reached during crawl |
| `urls_by_status_code` | map | Distribution of HTTP status codes (e.g., \{"200": 115, "404": 3\}) |
| `skipped_by_reason` | map | Distribution of skip reasons |
| `retry_distribution` | map | Distribution of retry counts (e.g., \{1: 8, 2: 5, 3: 2\}) |

**Example Summary:**
```json
{
  "total": 150,
  "passed": 120,
  "failed": 15,
  "skipped": 10,
  "discovered": 5,
  "skipped_by_robots": 3,
  "skipped_by_domain": 4,
  "skipped_by_duplicate": 2,
  "skipped_by_content_type": 0,
  "skipped_by_depth": 1,
  "skipped_by_other": 0,
  "retried_requests": 12,
  "max_depth": 3,
  "urls_by_status_code": {
    "200": 115,
    "404": 3,
    "500": 2
  },
  "skipped_by_reason": {
    "skipped_by_robots": 3,
    "skipped_by_domain": 4,
    "skipped_by_duplicate": 2,
    "skipped_by_depth": 1
  },
  "retry_distribution": {
    "1": 8,
    "2": 3,
    "3": 1
  }
}
```

## Dependencies

### External Libraries

- [`golang.org/x/net/html`](https://pkg.go.dev/golang.org/x/net/html) - HTML tokenization and parsing
- [`github.com/temoto/robotstxt`](https://github.com/temoto/robotstxt) - robots.txt parsing and compliance
- [`golang.org/x/net`](https://pkg.go.dev/golang.org/x/net) - Additional networking utilities (context, trace)

**Why These Libraries?**

**golang.org/x/net/html:**
- Production-grade HTML parser
- Handles malformed HTML gracefully
- Token-based parsing for efficiency
- No external C dependencies

**github.com/temoto/robotstxt:**
- Full robots.txt specification support
- Handles wildcards and crawl-delay
- Actively maintained
- Simple API

### Standard Library

DeepScanBot leverages Go's standard library extensively:

- **`net/http`** - HTTP client with connection pooling
- **`net/url`** - URL parsing, resolution, and normalization
- **`sync`** - Concurrency primitives (Mutex, WaitGroup, Once)
- **`encoding/json`** - JSON serialization/deserialization
- **`bufio`** - Buffered I/O for efficient reading
- **`context`** - Request cancellation and timeouts
- **`time`** - Timeouts, delays, and backoff
- **`flag`** - Command-line argument parsing
- **`os`** - File I/O and environment interaction
- **`strings`** - String manipulation and parsing
- **`strconv`** - String to primitive conversions
- **`errors`** - Error creation and handling

**Standard Library Benefits:**
- No external dependencies for core functionality
- Well-tested and optimized
- Consistent API across platforms
- Small binary size

## Design Patterns

DeepScanBot employs several proven design patterns for maintainability and extensibility:

### Worker Pool Pattern

**Purpose:** Concurrent fetching with controlled parallelism

**Implementation:**
- Fixed number of worker goroutines
- Shared job queue
- Semaphore-based concurrency control

**Benefits:**
- Predictable resource usage
- Prevents system overload
- Easy to scale

**Example:**
```go
func (c *Crawler) startWorkers() {
    for i := 0; i < c.options.Concurrency; i++ {
        go c.worker()
    }
}
```

### Strategy Pattern

**Purpose:** Pluggable output formats (text/JSON)

**Implementation:**
- `OutputStrategy` interface
- `TextOutput` and `JSONOutput` implementations
- Runtime selection based on flags

**Benefits:**
- Easy to add new output formats
- Separation of concerns
- Testable output logic

**Example:**
```go
type OutputStrategy interface {
    Save(filename string, data interface{}) error
}

type TextOutput struct{}
type JSONOutput struct{}
```

### Observer Pattern

**Purpose:** Event-driven URL discovery and processing

**Implementation:**
- URL discovery events (found, skipped, failed)
- Callback functions for each event type
- Decoupled event producers and consumers

**Benefits:**
- Extensible event handling
- Easy to add logging/metrics
- Loose coupling

**Example:**
```go
type CrawlEvent struct {
    URL   string
    Event string // "discovered", "crawled", "skipped"
}

type EventHandler func(event CrawlEvent)
```

### Repository Pattern

**Purpose:** Storage abstraction for different output formats

**Implementation:**
- `Storage` interface for data access
- `FileStorage` implementation
- `MemoryStorage` for testing

**Benefits:**
- Abstract data persistence
- Easy to swap storage backends
- Testable with mock storage

**Example:**
```go
type Storage interface {
    Save(results []URLEntry) error
    Load() ([]URLEntry, error)
}
```

### Builder Pattern

**Purpose:** Fluent configuration via `NewCrawlerWithOptions`

**Implementation:**
- `CrawlerBuilder` struct
- Chainable configuration methods
- Single `Build()` method to create instance

**Benefits:**
- Readable configuration
- Type-safe options
- Default value handling

**Example:**
```go
crawler := NewCrawlerWithOptions().
    WithURL("https://example.com").
    WithDepth(3).
    WithConcurrency(10).
    Build()
```

### Additional Patterns

**Semaphore Pattern:**
- Controls concurrent access to resources
- Implements per-host rate limiting
- Prevents resource exhaustion

**Retry Pattern:**
- Exponential backoff strategy
- Configurable retry policies
- Respects server signals (Retry-After)

**Pipeline Pattern:**
- Fetch → Parse → Store
- Each stage runs concurrently
- Backpressure handling
