package crawler

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// sitemapDocument represents an XML sitemap index or URL set.
type sitemapDocument struct {
	URLs     []sitemapLocation `xml:"url"`
	Sitemaps []sitemapLocation `xml:"sitemap"`
}

type sitemapLocation struct {
	Loc string `xml:"loc"`
}

// enqueueSitemapURLs discovers and queues URLs from the site's sitemap.xml.
func (c *Crawler) enqueueSitemapURLs() {
	if c.seedOrigin == "" {
		return
	}

	sitemapURL := c.seedOrigin + "/sitemap.xml"
	urls, err := c.fetchSitemapURLs(sitemapURL, 0)

	if err != nil {
		c.log.Infof("Sitemap unavailable at %s: %v", sitemapURL, err)
		return
	}

	for _, entry := range urls {
		c.handleDiscoveredLink(entry, "sitemap", 1)
	}

	c.log.Infof("Sitemap discovery queued %d URLs from %s", len(urls), sitemapURL)
}

// fetchSitemapURLs recursively fetches and parses a sitemap, supporting sitemap indexes.
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
			c.log.Infof("Sitemap child unavailable at %s: %v", loc, err)
			continue
		}

		urls = append(urls, childURLs...)
	}

	return urls, nil
}
