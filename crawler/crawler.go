package crawler

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/temoto/robotstxt"
	"web-crawler-assignment/fetcher"
	"web-crawler-assignment/parser"
	"web-crawler-assignment/storage"
)

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
	storage          *storage.PageStorage
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
}

type Options struct {
	Retries            int
	RetryBackoff       time.Duration
	CrawlDelay         time.Duration
	PerHostConcurrency int
	IncludeSitemap     bool
	ResumeEntries      []storage.URLEntry
}

func NewCrawler(startURL string, maxDepth int, timeout time.Duration, proxyUrl string, maxSize int, disableRedirects bool, insecure bool, uniqueUrls bool, concurrency int, contentTypes []string, ignoreRobots bool, crossDomain bool) *Crawler {
	return NewCrawlerWithOptions(startURL, maxDepth, timeout, proxyUrl, maxSize, disableRedirects, insecure, uniqueUrls, concurrency, contentTypes, ignoreRobots, crossDomain, Options{})
}

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
	pageStorage := storage.NewPageStorage()
	pageStorage.SeedEntries(options.ResumeEntries)
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
		storage:          pageStorage,
		robotsCache:      make(map[string]*robotstxt.RobotsData),
		robotsLoaded:     make(map[string]bool),
		hostSemaphores:   make(map[string]chan struct{}),
		hostLastRequest:  make(map[string]time.Time),
		sem:              make(chan struct{}, concurrency),
	}
}

func (c *Crawler) Start() ([]storage.URLEntry, error) {
	report, err := c.StartReport()
	if err != nil {
		return nil, err
	}
	return append(append([]storage.URLEntry{}, report.URLs...), report.Skipped...), nil
}

func (c *Crawler) StartReport() (storage.CrawlReport, error) {
	startedAt := time.Now()
	log.Printf("Starting crawl: url=%s max-depth=%d concurrency=%d per-host-concurrency=%d retries=%d delay=%s sitemap=%t resumed=%d cpu-cores=%d gomaxprocs=%d",
		c.startURL, c.maxDepth, cap(c.sem), c.perHostLimit, c.retries, c.crawlDelay, c.includeSitemap, len(c.resumeEntries), runtime.NumCPU(), runtime.GOMAXPROCS(0))

	c.enqueueCrawl(c.startURL, "href", 0)
	if c.includeSitemap && c.maxDepth > 0 {
		c.enqueueSitemapURLs()
	}

	c.wg.Wait()
	finishedAt := time.Now()
	report := storage.NewCrawlReport(c.startURL, "", startedAt, finishedAt, c.storage.Results())
	log.Printf("Crawl finished: url=%s total=%d passed=%d failed=%d skipped=%d discovered=%d retried=%d max-depth=%d duration=%s",
		c.startURL, report.Summary.Total, report.Summary.Passed, report.Summary.Failed, report.Summary.Skipped, report.Summary.Discovered, report.Summary.RetriedRequests, report.Summary.MaxDepth, finishedAt.Sub(startedAt).Round(time.Millisecond))
	return report, nil
}

func (c *Crawler) crawl(url string, depth int) {
	defer c.wg.Done()

	c.recordDepth(depth)
	if !c.allowedByRobots(url) {
		c.skipped.Add(1)
		c.storage.StoreEntry(storage.URLEntry{
			URL:           url,
			Depth:         depth,
			Result:        "skipped",
			SkippedReason: "disallowed by robots.txt",
		})
		log.Printf("Skipping %s because robots.txt disallows it", url)
		return
	}

	release := c.acquireRequestSlots(url)
	defer release()

	data, size, contentType, statusCode, attempts, err := c.fetchWithRetry(url)
	if err != nil {
		c.failed.Add(1)
		c.storage.StoreEntry(storage.URLEntry{
			URL:         url,
			Depth:       depth,
			StatusCode:  statusCode,
			ContentType: contentType,
			Result:      "failed",
			Error:       err.Error(),
			Attempts:    attempts,
		})
		log.Printf("Error fetching URL %s: %v\n", url, err)
		return
	}

	if c.maxSize > 0 && size > c.maxSize*1024 {
		c.failed.Add(1)
		c.storage.StoreEntry(storage.URLEntry{
			URL:         url,
			Depth:       depth,
			StatusCode:  statusCode,
			ContentType: contentType,
			Result:      "failed",
			Error:       "page exceeds configured size limit",
			Attempts:    attempts,
		})
		log.Printf("Skipping URL %s due to size limit (%d bytes > %d bytes)\n", url, size, c.maxSize*1024)
		return
	}

	if data == nil && contentType != "" {
		c.skipped.Add(1)
		c.storage.StoreEntry(storage.URLEntry{
			URL:           url,
			Depth:         depth,
			StatusCode:    statusCode,
			ContentType:   contentType,
			Result:        "skipped",
			SkippedReason: "content type not allowed",
			Attempts:      attempts,
		})
		log.Printf("Skipping %s because content type %q is not allowed", url, contentType)
		return
	}

	c.storage.StoreEntry(storage.URLEntry{
		URL:         url,
		Depth:       depth,
		StatusCode:  statusCode,
		ContentType: contentType,
		Result:      "passed",
		Attempts:    attempts,
	})
	c.fetched.Add(1)

	if depth < c.maxDepth && strings.Contains(strings.ToLower(contentType), "text/html") {
		links := parser.Parse(data, url)
		for link, source := range links {
			c.handleDiscoveredLink(link, source, depth+1)
		}
	}
}

func (c *Crawler) handleDiscoveredLink(targetURL, source string, depth int) {
	if source == "href" || source == "iframe" || source == "sitemap" {
		c.enqueueCrawl(targetURL, source, depth)
		return
	}

	if !c.shouldFollow(targetURL) {
		c.storeSkipped(targetURL, source, depth, "outside domain scope")
		return
	}
	if c.uniqueUrls && !c.storage.MarkVisitedIfNew(targetURL) {
		c.storeSkipped(targetURL, source, depth, "duplicate")
		return
	}
	c.storage.StoreSource(targetURL, source)
	c.storage.StoreEntry(storage.URLEntry{
		URL:    targetURL,
		Depth:  depth,
		Result: "discovered",
	})
}

func (c *Crawler) enqueueCrawl(targetURL, source string, depth int) {
	if depth > c.maxDepth {
		c.storeSkipped(targetURL, source, depth, "max depth exceeded")
		return
	}
	if !c.shouldFollow(targetURL) {
		c.storeSkipped(targetURL, source, depth, "outside domain scope")
		return
	}
	if c.uniqueUrls && !c.storage.MarkVisitedIfNew(targetURL) {
		c.storeSkipped(targetURL, source, depth, "duplicate")
		return
	}
	c.storage.StoreSource(targetURL, source)
	c.wg.Add(1)
	go c.crawl(targetURL, depth)
}

func (c *Crawler) storeSkipped(targetURL, source string, depth int, reason string) {
	c.skipped.Add(1)
	c.storage.StoreSource(targetURL, source)
	c.storage.StoreEntry(storage.URLEntry{
		URL:           targetURL,
		Depth:         depth,
		Result:        "skipped",
		SkippedReason: reason,
	})
}

func (c *Crawler) fetchWithRetry(targetURL string) ([]byte, int, string, int, int, error) {
	var data []byte
	var size int
	var contentType string
	var statusCode int
	var err error
	maxAttempts := c.retries + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		c.waitForHostDelay(targetURL)

		// Use FetchWithDetails to get retry-after header info
		result := fetcher.FetchWithDetails(targetURL, c.timeout, c.proxyUrl, c.disableRedirects, c.insecure, c.maxSize, c.contentTypes)
		data, size, contentType, statusCode, err = result.Body, result.Size, result.ContentType, result.StatusCode, result.Err

		if err == nil {
			return data, size, contentType, statusCode, attempt, nil
		}

		if attempt == maxAttempts || !isRetryable(statusCode) {
			return data, size, contentType, statusCode, attempt, err
		}

		// Use Retry-After header if available, otherwise calculate backoff
		delay := c.retryDelay(attempt, statusCode, result.RetryAfter)
		log.Printf("Retrying URL %s after %s because status=%d error=%v", targetURL, delay, statusCode, err)
		time.Sleep(delay)
	}
	return data, size, contentType, statusCode, maxAttempts, err
}

func isRetryable(statusCode int) bool {
	return statusCode == 0 || statusCode == http.StatusRequestTimeout || statusCode == http.StatusTooManyRequests || statusCode >= 500
}

// retryDelay calculates the delay before the next retry attempt.
// Uses the Retry-After header if provided, otherwise uses exponential backoff.
// For 429 Too Many Requests specifically, adds extra multiplier to be more respectful.
func (c *Crawler) retryDelay(attempt int, statusCode int, retryAfter time.Duration) time.Duration {
	// If the server specified a Retry-After duration, respect it
	if retryAfter > 0 {
		return retryAfter
	}

	// Exponential backoff: base * attempt
	delay := time.Duration(attempt) * c.retryBackoff

	// Extra backoff for rate-limiting (429) scenarios like Goodreads
	if statusCode == http.StatusTooManyRequests {
		delay *= 3
	}

	// Cap the maximum delay to 30 seconds
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}

	return delay
}

func (c *Crawler) acquireRequestSlots(targetURL string) func() {
	hostSem := c.hostSemaphore(targetURL)
	hostSem <- struct{}{}
	c.sem <- struct{}{}
	return func() {
		<-c.sem
		<-hostSem
	}
}

func (c *Crawler) hostSemaphore(targetURL string) chan struct{} {
	host := c.hostKey(targetURL)
	c.hostMu.Lock()
	defer c.hostMu.Unlock()
	if c.hostSemaphores[host] == nil {
		c.hostSemaphores[host] = make(chan struct{}, c.perHostLimit)
	}
	return c.hostSemaphores[host]
}

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

func (c *Crawler) hostKey(targetURL string) string {
	parsedURL, err := url.Parse(targetURL)
	if err != nil || parsedURL.Host == "" {
		return targetURL
	}
	return parsedURL.Host
}

func (c *Crawler) shouldFollow(targetURL string) bool {
	if c.crossDomain {
		return true
	}
	parsedURL, err := url.Parse(targetURL)
	return err == nil && parsedURL.Host == c.seedHost
}

func (c *Crawler) enqueueSitemapURLs() {
	if c.seedOrigin == "" {
		return
	}
	sitemapURL := c.seedOrigin + "/sitemap.xml"
	urls, err := c.fetchSitemapURLs(sitemapURL, 0)
	if err != nil {
		log.Printf("Sitemap unavailable at %s: %v", sitemapURL, err)
		return
	}
	for _, sitemapEntry := range urls {
		c.handleDiscoveredLink(sitemapEntry, "sitemap", 1)
	}
	log.Printf("Sitemap discovery queued %d URLs from %s", len(urls), sitemapURL)
}

type sitemapDocument struct {
	URLs     []sitemapLocation `xml:"url"`
	Sitemaps []sitemapLocation `xml:"sitemap"`
}

type sitemapLocation struct {
	Loc string `xml:"loc"`
}

func (c *Crawler) fetchSitemapURLs(sitemapURL string, depth int) ([]string, error) {
	if depth > 1 {
		return nil, nil
	}
	req, err := http.NewRequest(http.MethodGet, sitemapURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "DeepScanBot/1.0")
	response, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}
	var document sitemapDocument
	if err := xml.Unmarshal(body, &document); err != nil {
		return nil, err
	}

	var urls []string
	for _, entry := range document.URLs {
		if loc := strings.TrimSpace(entry.Loc); loc != "" {
			urls = append(urls, loc)
		}
	}
	for _, sitemap := range document.Sitemaps {
		loc := strings.TrimSpace(sitemap.Loc)
		if loc == "" {
			continue
		}
		childURLs, err := c.fetchSitemapURLs(loc, depth+1)
		if err != nil {
			log.Printf("Sitemap child unavailable at %s: %v", loc, err)
			continue
		}
		urls = append(urls, childURLs...)
	}
	return urls, nil
}

func (c *Crawler) httpClient() *http.Client {
	client := &http.Client{Timeout: c.timeout}
	if c.disableRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	transport := &http.Transport{}
	hasCustomTransport := false
	if c.proxyUrl != "" {
		proxy, err := url.Parse(c.proxyUrl)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxy)
			hasCustomTransport = true
		}
	}
	if c.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		hasCustomTransport = true
	}
	if hasCustomTransport {
		client.Transport = transport
	}
	return client
}

func (c *Crawler) recordDepth(depth int) {
	for {
		current := c.deepestDepth.Load()
		if depth <= int(current) || c.deepestDepth.CompareAndSwap(current, int64(depth)) {
			return
		}
	}
}

func (c *Crawler) allowedByRobots(targetURL string) bool {
	if c.ignoreRobots {
		return true
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	origin := parsedURL.Scheme + "://" + parsedURL.Host
	c.robotsMu.Lock()
	defer c.robotsMu.Unlock()
	if !c.robotsLoaded[origin] {
		c.robotsCache[origin] = c.fetchRobots(origin)
		c.robotsLoaded[origin] = true
	}

	robotsData := c.robotsCache[origin]
	if robotsData == nil {
		return true
	}
	path := parsedURL.EscapedPath()
	if path == "" {
		path = "/"
	}
	return robotsData.TestAgent(path, "DeepScanBot")
}

func (c *Crawler) fetchRobots(origin string) *robotstxt.RobotsData {
	client := &http.Client{Timeout: c.timeout}
	transport := &http.Transport{}
	hasCustomTransport := false
	if c.proxyUrl != "" {
		proxy, err := url.Parse(c.proxyUrl)
		if err != nil {
			log.Printf("Error parsing proxy for robots.txt: %v", err)
			return nil
		}
		transport.Proxy = http.ProxyURL(proxy)
		hasCustomTransport = true
	}
	if c.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		hasCustomTransport = true
	}
	if hasCustomTransport {
		client.Transport = transport
	}

	req, err := http.NewRequest(http.MethodGet, origin+"/robots.txt", nil)
	if err != nil {
		log.Printf("Error creating robots.txt request: %v", err)
		return nil
	}
	req.Header.Set("User-Agent", "DeepScanBot/1.0")
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching robots.txt from %s: %v", origin, err)
		return nil
	}
	defer response.Body.Close()

	robotsData, err := robotstxt.FromResponse(response)
	if err != nil {
		log.Printf("Error parsing robots.txt from %s: %v", origin, err)
		return nil
	}
	return robotsData
}
