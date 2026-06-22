package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"sync/atomic"
	"testing"
	"time"
	"web-crawler-assignment/crawler"
	"web-crawler-assignment/storage"
)

func TestCrawlerStartReturnsResultsWithoutWritingFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 0, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
	results, err := c.Start()
	if err != nil {
		t.Fatalf("start crawler: %v", err)
	}

	want := []storage.URLEntry{{URL: server.URL, Source: "href", StatusCode: http.StatusOK}}
	if !reflect.DeepEqual(results, want) {
		t.Errorf("results = %#v, want %#v", results, want)
	}
	if _, err := os.Stat("crawler_results.txt"); !os.IsNotExist(err) {
		t.Errorf("library crawl created crawler_results.txt: %v", err)
	}
	if _, err := os.Stat("crawler_results.json"); !os.IsNotExist(err) {
		t.Errorf("library crawl created crawler_results.json: %v", err)
	}
}

func TestCrawlerRespectsRobots(t *testing.T) {
	var privateRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt":
			_, _ = w.Write([]byte("User-agent: DeepScanBot\nDisallow: /private\n"))
		case "/public":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html></html>"))
		case "/private":
			privateRequests.Add(1)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html></html>"))
		}
	}))
	defer server.Close()

	allowed := crawler.NewCrawler(server.URL+"/public", 0, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
	if results, err := allowed.Start(); err != nil || len(results) != 1 {
		t.Fatalf("allowed crawl results = %#v, error = %v", results, err)
	}

	disallowed := crawler.NewCrawler(server.URL+"/private", 0, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
	if results, err := disallowed.Start(); err != nil || len(results) != 1 || results[0].Error != "disallowed by robots.txt" {
		t.Fatalf("disallowed crawl results = %#v, error = %v", results, err)
	}
	if got := privateRequests.Load(); got != 0 {
		t.Errorf("private page requests = %d, want 0", got)
	}
}

func TestCrawlerProcessesDiscoveredLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt":
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<a href=\"/child\">child</a>"))
		case "/child":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html></html>"))
		}
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 1, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
	results, err := c.Start()
	if err != nil {
		t.Fatalf("start crawler: %v", err)
	}
	urls := make([]string, 0, len(results))
	for _, result := range results {
		urls = append(urls, result.URL)
	}
	sort.Strings(urls)
	want := []string{server.URL, server.URL + "/child"}
	sort.Strings(want)
	if !reflect.DeepEqual(urls, want) {
		t.Errorf("result URLs = %v, want %v", urls, want)
	}
}

func TestCrawlerDoesNotFollowExternalHostsByDefault(t *testing.T) {
	var externalRequests atomic.Int32
	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		externalRequests.Add(1)
		http.NotFound(w, r)
	}))
	defer external.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt":
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<a href=\"" + external.URL + "/outside\">outside</a>"))
		}
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 1, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
	results, err := c.Start()
	if err != nil {
		t.Fatalf("start crawler: %v", err)
	}
	if got := externalRequests.Load(); got != 0 {
		t.Errorf("external requests = %d, want 0", got)
	}
	if len(results) != 1 || results[0].URL != server.URL {
		t.Errorf("results = %#v, want only %q", results, server.URL)
	}
}

func TestCrawlerIgnoreRobotsAllowsDisallowedPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /private\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL+"/private", 0, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, true, false)
	results, err := c.Start()
	if err != nil {
		t.Fatalf("start crawler: %v", err)
	}
	if len(results) != 1 || results[0].URL != server.URL+"/private" {
		t.Errorf("results = %#v, want allowed private URL", results)
	}
}

func TestCrawlerStoresFailedFetchResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 0, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
	results, err := c.Start()
	if err != nil {
		t.Fatalf("start crawler: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1", len(results))
	}
	result := results[0]
	if result.StatusCode != http.StatusServiceUnavailable || result.Error == "" {
		t.Errorf("failed result = %#v, want status %d and an error", result, http.StatusServiceUnavailable)
	}
}
