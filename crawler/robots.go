package crawler

import (
	"net/http"
	"net/url"

	"github.com/temoto/robotstxt"
)

// allowedByRobots checks robots.txt rules for the given URL.
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

// fetchRobots fetches and parses robots.txt from the given origin.
func (c *Crawler) fetchRobots(origin string) *robotstxt.RobotsData {
	client := &http.Client{Timeout: c.timeout}
	transport := c.buildTransport()

	if transport != nil {
		client.Transport = transport
	}

	req, err := http.NewRequest(http.MethodGet, origin+"/robots.txt", nil)
	if err != nil {
		c.log.Errorf("Error creating robots.txt request: %v", err)
		return nil
	}

	req.Header.Set("User-Agent", "DeepScanBot/1.0")

	response, err := client.Do(req)
	if err != nil {
		c.log.Errorf("Error fetching robots.txt from %s: %v", origin, err)
		return nil
	}

	defer response.Body.Close()

	robotsData, err := robotstxt.FromResponse(response)
	if err != nil {
		c.log.Errorf("Error parsing robots.txt from %s: %v", origin, err)
		return nil
	}

	return robotsData
}
