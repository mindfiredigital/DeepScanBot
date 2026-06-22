package crawler

import (
	"net/http"
	"net/http/httptest"
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

	c := NewCrawler(server.URL, 0, time.Second, "", -1, false, false, false, 1, []string{"text/html"}, false)
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

func TestIgnoreRobotsAllowsDisallowedPaths(t *testing.T) {
	c := NewCrawler("https://example.com", 0, time.Second, "", -1, false, false, false, 1, []string{"text/html"}, true)
	if !c.allowedByRobots("https://example.com/private") {
		t.Error("ignore-robots should allow every valid URL")
	}
}
