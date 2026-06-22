package crawler

import (
	"crypto/tls"
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
	seedHost         string
	storage          *storage.PageStorage
	robotsMu         sync.Mutex
	robotsCache      map[string]*robotstxt.RobotsData
	robotsLoaded     map[string]bool
	wg               sync.WaitGroup
	sem              chan struct{}
	fetched          atomic.Int64
	failed           atomic.Int64
	skipped          atomic.Int64
	deepestDepth     atomic.Int64
}

func NewCrawler(startURL string, maxDepth int, timeout time.Duration, proxyUrl string, maxSize int, disableRedirects bool, insecure bool, uniqueUrls bool, concurrency int, contentTypes []string, ignoreRobots bool, crossDomain bool) *Crawler {
	if concurrency <= 0 {
		concurrency = runtime.GOMAXPROCS(0)
	}
	if maxDepth < 0 {
		maxDepth = 0
	}
	parsedStartURL, _ := url.Parse(startURL)
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
		seedHost:         parsedStartURL.Host,
		storage:          storage.NewPageStorage(),
		robotsCache:      make(map[string]*robotstxt.RobotsData),
		robotsLoaded:     make(map[string]bool),
		sem:              make(chan struct{}, concurrency),
	}
}

func (c *Crawler) Start() ([]storage.URLEntry, error) {
	startedAt := time.Now()
	log.Printf("Starting crawl: url=%s max-depth=%d concurrency=%d cpu-cores=%d gomaxprocs=%d", c.startURL, c.maxDepth, cap(c.sem), runtime.NumCPU(), runtime.GOMAXPROCS(0))

	c.storage.StoreSource(c.startURL, "href")
	c.wg.Add(1)
	go c.crawl(c.startURL, 0)

	c.wg.Wait()
	log.Printf("Crawl finished: url=%s fetched=%d failed=%d skipped=%d max-depth=%d duration=%s", c.startURL, c.fetched.Load(), c.failed.Load(), c.skipped.Load(), c.deepestDepth.Load(), time.Since(startedAt).Round(time.Millisecond))
	return c.storage.Results(), nil
}

func (c *Crawler) crawl(url string, depth int) {
	defer c.wg.Done()

	c.recordDepth(depth)
	if !c.allowedByRobots(url) {
		c.skipped.Add(1)
		c.storage.StoreResult(url, depth, 0, "disallowed by robots.txt")
		log.Printf("Skipping %s because robots.txt disallows it", url)
		return
	}

	if c.uniqueUrls {
		c.storage.MarkVisited(url)
	}

	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	data, size, contentType, statusCode, err := fetcher.Fetch(url, c.timeout, c.proxyUrl, c.disableRedirects, c.insecure, c.maxSize, c.contentTypes)
	if err != nil {
		c.failed.Add(1)
		c.storage.StoreResult(url, depth, statusCode, err.Error())
		log.Printf("Error fetching URL %s: %v\n", url, err)
		return
	}

	if c.maxSize > 0 && size > c.maxSize*1024 {
		c.failed.Add(1)
		c.storage.StoreResult(url, depth, statusCode, "page exceeds configured size limit")
		log.Printf("Skipping URL %s due to size limit (%d bytes > %d bytes)\n", url, size, c.maxSize*1024)
		return
	}

	c.storage.StoreResult(url, depth, statusCode, "")
	c.fetched.Add(1)

	if depth < c.maxDepth && strings.Contains(strings.ToLower(contentType), "text/html") {
		links := parser.Parse(data, url)
		for link, source := range links {
			if !c.shouldFollow(link) || (c.uniqueUrls && c.storage.HasVisited(link)) {
				continue
			}
			c.storage.StoreSource(link, source)
			if source == "href" || source == "iframe" {
				// Add before launching the child so Wait cannot observe a zero count
				// while discovered work is still pending.
				c.wg.Add(1)
				go c.crawl(link, depth+1)
			} else {
				if c.uniqueUrls {
					c.storage.MarkVisited(link)
				}
				c.storage.StoreResult(link, depth+1, 0, "")
			}
		}
	}
}

func (c *Crawler) shouldFollow(targetURL string) bool {
	if c.crossDomain {
		return true
	}
	parsedURL, err := url.Parse(targetURL)
	return err == nil && parsedURL.Host == c.seedHost
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
