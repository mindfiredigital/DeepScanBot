package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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

	c := crawler.NewCrawler(server.URL, 0, time.Second, "", -1, false, false, true, 1, []string{"text/html"}, false)
	results, err := c.Start()
	if err != nil {
		t.Fatalf("start crawler: %v", err)
	}

	want := []storage.URLEntry{{URL: server.URL, Source: "href"}}
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
