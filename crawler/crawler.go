package crawler

import (
	"log"
	"strings"
	"sync"
	"time"
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
	storage          *storage.PageStorage
	wg               sync.WaitGroup
	urlChan          chan string
	depthChan        chan int
	sem              chan struct{}
}

func NewCrawler(startURL string, maxDepth int, timeout time.Duration, proxyUrl string, maxSize int, disableRedirects bool, insecure bool, uniqueUrls bool, concurrency int, contentTypes []string) *Crawler {
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
		storage:          storage.NewPageStorage(),
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
