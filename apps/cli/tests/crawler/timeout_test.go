package crawler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mindfiredigital/DeepScanBot/packages/crawler"
	"github.com/mindfiredigital/DeepScanBot/packages/exitcode"
)

func TestCrawlerTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 1, 500*time.Millisecond, "", 0, false, false, false, 1, nil, false, false)

	_, err := c.StartReport()
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !errors.Is(err, exitcode.ErrTimeout) {
		t.Errorf("Expected ErrTimeout, got: %v", err)
	}
}

func TestCrawlerNoTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 1, 0, "", 0, false, false, false, 1, nil, false, false)

	report, err := c.StartReport()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.Summary.Total == 0 {
		t.Error("Expected at least one URL to be crawled")
	}
}

func TestCrawlerTimeoutLongerThanCrawl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 1, 10*time.Second, "", 0, false, false, false, 1, nil, false, false)

	report, err := c.StartReport()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.Summary.Total == 0 {
		t.Error("Expected at least one URL to be crawled")
	}
}

func TestCrawlerProgressLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	c := crawler.NewCrawler(server.URL, 1, 5*time.Second, "", 0, false, false, false, 1, nil, false, false)

	report, err := c.StartReport()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.Summary.Total == 0 {
		t.Error("Expected at least one URL to be crawled")
	}
}
