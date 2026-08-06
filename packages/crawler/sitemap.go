package crawler

import (
	"context"
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
		// Check if it's a 404 (not found) - this is expected and not an error
		if strings.Contains(err.Error(), "bad status code: 404") {
			c.log.Warnf("Sitemap not found at %s (404), continuing without sitemap", sitemapURL)
		} else {
			c.log.Errorf("Sitemap fetch failed at %s: %v", sitemapURL, err)
		}
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, sitemapURL, nil)
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

	// Check for 404 specifically to allow graceful handling
	if response.StatusCode == 404 {
		return nil, fmt.Errorf("bad status code: 404")
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
