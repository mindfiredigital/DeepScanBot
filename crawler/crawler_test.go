package crawler

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllowedByRobots(t *testing.T) {
	var robotsRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/robots.txt" {
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
		robotsRequests.Add(1)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("User-agent: DeepScanBot\nDisallow: /private\n"))
	}))
	defer server.Close()

	c := NewCrawler(server.URL, 0, time.Second, "", -1, false, false, false, 1, []string{"text/html"}, false, false)
	if !c.allowedByRobots(server.URL + "/public") {
		t.Error("public path should be allowed")
	}
	if c.allowedByRobots(server.URL + "/private/report") {
		t.Error("private path should be disallowed")
	}
	if got := robotsRequests.Load(); got != 1 {
		t.Errorf("robots.txt requests = %d, want 1", got)
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
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := NewCrawler(server.URL, 1, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
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
	if len(urls) != len(want) {
		t.Fatalf("result count = %d, want %d (%v)", len(urls), len(want), urls)
	}
	for i := range want {
		if urls[i] != want[i] {
			t.Errorf("result URL %d = %q, want %q", i, urls[i], want[i])
		}
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
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := NewCrawler(server.URL, 1, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false, false)
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

func TestIgnoreRobotsAllowsDisallowedPaths(t *testing.T) {
	c := NewCrawler("https://example.com", 0, time.Second, "", -1, false, false, false, 1, []string{"text/html"}, true, false)
	if !c.allowedByRobots("https://example.com/private") {
		t.Error("ignore-robots should allow every valid URL")
	}
}
