package crawler

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
	storage          *storage.PageStorage
	robotsMu         sync.Mutex
	robotsCache      map[string]*robotstxt.RobotsData
	robotsLoaded     map[string]bool
	wg               sync.WaitGroup
	urlChan          chan string
	depthChan        chan int
	sem              chan struct{}
}

func NewCrawler(startURL string, maxDepth int, timeout time.Duration, proxyUrl string, maxSize int, disableRedirects bool, insecure bool, uniqueUrls bool, concurrency int, contentTypes []string, ignoreRobots bool) *Crawler {
	if concurrency <= 0 {
		concurrency = 10
	}
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
		storage:          storage.NewPageStorage(),
		robotsCache:      make(map[string]*robotstxt.RobotsData),
		robotsLoaded:     make(map[string]bool),
		urlChan:          make(chan string),
		depthChan:        make(chan int),
		sem:              make(chan struct{}, concurrency),
	}
}

func (c *Crawler) Start() ([]storage.URLEntry, error) {
	log.Println("Start crawler", c)

	c.storage.StoreSource(c.startURL, "href")
	c.wg.Add(1)
	go c.crawl(c.startURL, 0)

	go func() {
		for url := range c.urlChan {
			depth := <-c.depthChan
			if depth <= c.maxDepth && (!c.uniqueUrls || !c.storage.HasVisited(url)) {
				c.wg.Add(1)
				go c.crawl(url, depth)
			}
		}
	}()

	c.wg.Wait()
	log.Println("Finished crawler", c)
	return c.storage.Results(), nil
}

func (c *Crawler) crawl(url string, depth int) {
	defer c.wg.Done()

	if depth > c.maxDepth {
		return
	}
	if !c.allowedByRobots(url) {
		log.Printf("Skipping %s because robots.txt disallows it", url)
		return
	}

	if c.uniqueUrls {
		c.storage.MarkVisited(url)
	}

	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	data, size, contentType, err := fetcher.Fetch(url, c.timeout, c.proxyUrl, c.disableRedirects, c.insecure, c.maxSize, c.contentTypes)
	if err != nil {
		log.Printf("Error fetching URL %s: %v\n", url, err)
		return
	}

	if c.maxSize > 0 && size > c.maxSize*1024 {
		log.Printf("Skipping URL %s due to size limit (%d bytes > %d bytes)\n", url, size, c.maxSize*1024)
		return
	}

	c.storage.StoreContent(url)

	if strings.Contains(strings.ToLower(contentType), "text/html") {
		links := parser.Parse(data, url)
		for link, source := range links {
			if !c.uniqueUrls || !c.storage.HasVisited(link) {
				if source == "href" || source == "iframe" {
					c.urlChan <- link
					c.depthChan <- depth + 1
					c.storage.StoreSource(link, source)
				} else {
					if c.uniqueUrls {
						c.storage.MarkVisited(link)
					}
					c.storage.StoreSource(link, source)
					c.storage.StoreContent(link)
				}
			}
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
