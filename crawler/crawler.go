package crawler

import (
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/temoto/robotstxt"
	"web-crawler-assignment/logger"
	"web-crawler-assignment/parser"
	"web-crawler-assignment/storage"
	"web-crawler-assignment/types"
)

// Options alias for backward compatibility.
type Options = types.CrawlerOptions

// Crawler orchestrates the web crawling process with concurrency control,
// rate limiting, robots.txt compliance, and retry logic.
type Crawler struct {
	startURL         string
	maxDepth         int
	timeout          time.Duration
	proxyUrl         string
	maxSize          int
	disableRedirects bool
	insecure         bool
	uniqueUrls       bool
	contentTypes     []string
	ignoreRobots     bool
	crossDomain      bool
	retries          int
	retryBackoff     time.Duration
	crawlDelay       time.Duration
	perHostLimit     int
	includeSitemap   bool
	resumeEntries    []storage.URLEntry
	seedHost         string
	seedOrigin       string
	pageStorage      *storage.PageStorage
	robotsMu         sync.Mutex
	robotsCache      map[string]*robotstxt.RobotsData
	robotsLoaded     map[string]bool
	hostMu           sync.Mutex
	hostSemaphores   map[string]chan struct{}
	hostLastRequest  map[string]time.Time
	wg               sync.WaitGroup
	sem              chan struct{}
	fetched          atomic.Int64
	failed           atomic.Int64
	skipped          atomic.Int64
	deepestDepth     atomic.Int64
	log              *logger.Logger
}

// NewCrawler creates a Crawler with default options.
func NewCrawler(startURL string, maxDepth int, timeout time.Duration, proxyUrl string, maxSize int, disableRedirects bool, insecure bool, uniqueUrls bool, concurrency int, contentTypes []string, ignoreRobots bool, crossDomain bool) *Crawler {
	return NewCrawlerWithOptions(startURL, maxDepth, timeout, proxyUrl, maxSize, disableRedirects, insecure, uniqueUrls, concurrency, contentTypes, ignoreRobots, crossDomain, Options{})
}

// NewCrawlerWithOptions creates a Crawler with the given options.
func NewCrawlerWithOptions(startURL string, maxDepth int, timeout time.Duration, proxyUrl string, maxSize int, disableRedirects bool, insecure bool, uniqueUrls bool, concurrency int, contentTypes []string, ignoreRobots bool, crossDomain bool, options Options) *Crawler {
	if concurrency <= 0 {
		concurrency = runtime.GOMAXPROCS(0)
	}
	if options.PerHostConcurrency <= 0 {
		options.PerHostConcurrency = concurrency
	}
	if options.Retries < 0 {
		options.Retries = 0
	}
	if options.RetryBackoff <= 0 {
		options.RetryBackoff = time.Second
	}
	if maxDepth < 0 {
		maxDepth = 0
	}

	parsedStartURL, _ := url.Parse(startURL)
	seedOrigin := ""
	if parsedStartURL != nil && parsedStartURL.Scheme != "" && parsedStartURL.Host != "" {
		seedOrigin = parsedStartURL.Scheme + "://" + parsedStartURL.Host
	}
	ps := storage.NewPageStorage()
	ps.SeedEntries(options.ResumeEntries)

	return &Crawler{
		startURL:         startURL,
		maxDepth:         maxDepth,
		timeout:          timeout,
		proxyUrl:         proxyUrl,
		maxSize:          maxSize,
		disableRedirects: disableRedirects,
		insecure:         insecure,
		uniqueUrls:       uniqueUrls,
		contentTypes:     contentTypes,
		ignoreRobots:     ignoreRobots,
		crossDomain:      crossDomain,
		retries:          options.Retries,
		retryBackoff:     options.RetryBackoff,
		crawlDelay:       options.CrawlDelay,
		perHostLimit:     options.PerHostConcurrency,
		includeSitemap:   options.IncludeSitemap,
		resumeEntries:    options.ResumeEntries,
		seedHost:         parsedStartURL.Host,
		seedOrigin:       seedOrigin,
		pageStorage:      ps,
		robotsCache:      make(map[string]*robotstxt.RobotsData),
		robotsLoaded:     make(map[string]bool),
		hostSemaphores:   make(map[string]chan struct{}),
		hostLastRequest:  make(map[string]time.Time),
		sem:              make(chan struct{}, concurrency),
		log:              logger.New("info"),
	}
}

// Start runs the crawl and returns all URL entries.
func (c *Crawler) Start() ([]storage.URLEntry, error) {
	report, err := c.StartReport()
	if err != nil {
		return nil, err
	}
	return append(append([]storage.URLEntry{}, report.URLs...), report.Skipped...), nil
}

// StartReport runs the crawl and returns a detailed report.
func (c *Crawler) StartReport() (storage.CrawlReport, error) {
	startedAt := time.Now()
	c.log.Infof("Starting crawl: url=%s max-depth=%d concurrency=%d per-host-concurrency=%d retries=%d delay=%s sitemap=%t resumed=%d",
		c.startURL, c.maxDepth, cap(c.sem), c.perHostLimit, c.retries, c.crawlDelay, c.includeSitemap, len(c.resumeEntries))

	c.enqueueCrawl(c.startURL, "href", 0)
	if c.includeSitemap && c.maxDepth > 0 {
		c.enqueueSitemapURLs()
	}
	c.wg.Wait()

	finishedAt := time.Now()
	report := storage.NewCrawlReport(c.startURL, "", startedAt, finishedAt, c.pageStorage.Results())
	c.log.Infof("Crawl finished: url=%s total=%d passed=%d failed=%d skipped=%d duration=%s",
		c.startURL, report.Summary.Total, report.Summary.Passed, report.Summary.Failed, report.Summary.Skipped, finishedAt.Sub(startedAt).Round(time.Millisecond))
	return report, nil
}

// crawl processes a single URL: checks robots.txt, fetches, stores results, and parses links.
func (c *Crawler) crawl(url string, depth int) {
	defer c.wg.Done()
	c.recordDepth(depth)

	if !c.allowedByRobots(url) {
		c.skipped.Add(1)
		c.pageStorage.StoreEntry(storage.URLEntry{URL: url, Depth: depth, Result: "skipped", SkippedReason: "disallowed by robots.txt"})
		c.log.Infof("Skipping %s because robots.txt disallows it", url)
		return
	}

	release := c.acquireRequestSlots(url)
	defer release()

	data, size, contentType, statusCode, attempts, err := c.fetchWithRetry(url)
	if err != nil {
		c.failed.Add(1)
		c.pageStorage.StoreEntry(storage.URLEntry{URL: url, Depth: depth, StatusCode: statusCode, ContentType: contentType, Result: "failed", Error: err.Error(), Attempts: attempts})
		c.log.Errorf("Error fetching URL %s: %v", url, err)
		return
	}

	if c.maxSize > 0 && size > c.maxSize*1024 {
		c.failed.Add(1)
		c.pageStorage.StoreEntry(storage.URLEntry{URL: url, Depth: depth, StatusCode: statusCode, ContentType: contentType, Result: "failed", Error: "page exceeds configured size limit", Attempts: attempts})
		c.log.Infof("Skipping URL %s due to size limit (%d bytes > %d bytes)", url, size, c.maxSize*1024)
		return
	}

	if data == nil && contentType != "" {
		c.skipped.Add(1)
		c.pageStorage.StoreEntry(storage.URLEntry{URL: url, Depth: depth, StatusCode: statusCode, ContentType: contentType, Result: "skipped", SkippedReason: "content type not allowed", Attempts: attempts})
		c.log.Infof("Skipping %s because content type %q is not allowed", url, contentType)
		return
	}

	c.fetched.Add(1)
	c.pageStorage.StoreEntry(storage.URLEntry{URL: url, Depth: depth, StatusCode: statusCode, ContentType: contentType, Result: "passed", Attempts: attempts})

	if depth < c.maxDepth && strings.Contains(strings.ToLower(contentType), "text/html") {
		for link, source := range parser.Parse(data, url) {
			c.handleDiscoveredLink(link, source, depth+1)
		}
	}
}

// handleDiscoveredLink processes a discovered link from a parsed page.
func (c *Crawler) handleDiscoveredLink(targetURL, source string, depth int) {
	if source == "href" || source == "iframe" || source == "sitemap" || source == "script" || source == "img" || source == "link" || source == "form" {
		c.enqueueCrawl(targetURL, source, depth)
		return
	}
	if !c.shouldFollow(targetURL) {
		c.storeSkipped(targetURL, source, depth, "outside domain scope")
		return
	}
	if c.uniqueUrls && !c.pageStorage.MarkVisitedIfNew(targetURL) {
		c.storeSkipped(targetURL, source, depth, "duplicate")
		return
	}
	c.pageStorage.StoreSource(targetURL, source)
	c.pageStorage.StoreEntry(storage.URLEntry{URL: targetURL, Depth: depth, Result: "discovered"})
}

// enqueueCrawl queues a URL for crawling if it passes all checks.
func (c *Crawler) enqueueCrawl(targetURL, source string, depth int) {
	if depth > c.maxDepth {
		c.storeSkipped(targetURL, source, depth, "max depth exceeded")
		return
	}
	if !c.shouldFollow(targetURL) {
		c.storeSkipped(targetURL, source, depth, "outside domain scope")
		return
	}
	if c.uniqueUrls && !c.pageStorage.MarkVisitedIfNew(targetURL) {
		c.storeSkipped(targetURL, source, depth, "duplicate")
		return
	}
	c.pageStorage.StoreSource(targetURL, source)
	c.wg.Add(1)
	go c.crawl(targetURL, depth)
}

// storeSkipped records a skipped URL entry.
func (c *Crawler) storeSkipped(targetURL, source string, depth int, reason string) {
	c.skipped.Add(1)
	c.pageStorage.StoreSource(targetURL, source)
	c.pageStorage.StoreEntry(storage.URLEntry{URL: targetURL, Depth: depth, Result: "skipped", SkippedReason: reason})
}

// shouldFollow checks if a URL should be followed based on domain scope.
func (c *Crawler) shouldFollow(targetURL string) bool {
	if c.crossDomain {
		return true
	}
	parsedURL, err := url.Parse(targetURL)
	return err == nil && parsedURL.Host == c.seedHost
}

// hostKey extracts the host portion from a URL for rate-limiting.
func (c *Crawler) hostKey(targetURL string) string {
	parsedURL, err := url.Parse(targetURL)
	if err != nil || parsedURL.Host == "" {
		return targetURL
	}
	return parsedURL.Host
}

// acquireRequestSlots acquires concurrency and per-host slots before fetching.
func (c *Crawler) acquireRequestSlots(targetURL string) func() {
	hostSem := c.hostSemaphore(targetURL)
	hostSem <- struct{}{}
	c.sem <- struct{}{}
	return func() {
		<-c.sem
		<-hostSem
	}
}

// hostSemaphore returns or creates a per-host concurrency limiter.
func (c *Crawler) hostSemaphore(targetURL string) chan struct{} {
	host := c.hostKey(targetURL)
	c.hostMu.Lock()
	defer c.hostMu.Unlock()
	if c.hostSemaphores[host] == nil {
		c.hostSemaphores[host] = make(chan struct{}, c.perHostLimit)
	}
	return c.hostSemaphores[host]
}

// waitForHostDelay waits for the politeness delay between requests to the same host.
func (c *Crawler) waitForHostDelay(targetURL string) {
	if c.crawlDelay <= 0 {
		return
	}
	host := c.hostKey(targetURL)
	c.hostMu.Lock()
	defer c.hostMu.Unlock()
	if last := c.hostLastRequest[host]; !last.IsZero() {
		if wait := c.crawlDelay - time.Since(last); wait > 0 {
			c.hostMu.Unlock()
			time.Sleep(wait)
			c.hostMu.Lock()
		}
	}
	c.hostLastRequest[host] = time.Now()
}

// recordDepth atomically tracks the deepest depth reached.
func (c *Crawler) recordDepth(depth int) {
	for {
		current := c.deepestDepth.Load()
		if depth <= int(current) || c.deepestDepth.CompareAndSwap(current, int64(depth)) {
			return
		}
	}
}
