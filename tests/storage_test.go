package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"web-crawler-assignment/storage"
)

func TestTextOutputIsTruncatedForEachStorageInstance(t *testing.T) {
	const filename = "crawler_results.txt"
	t.Chdir(t.TempDir())

	if err := os.WriteFile(filename, []byte("result from a previous crawl\n"), 0644); err != nil {
		t.Fatalf("seed previous output: %v", err)
	}

	pageStorage := storage.NewPageStorage(false, -1)
	pageStorage.StoreContent("https://example.com/current", false)
	if err := pageStorage.Close(); err != nil {
		t.Fatalf("close output: %v", err)
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got, want := string(contents), "https://example.com/current\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestTextOutputIsFlushedOnClose(t *testing.T) {
	const filename = "crawler_results.txt"
	t.Chdir(t.TempDir())

	pageStorage := storage.NewPageStorage(false, -1)
	pageStorage.StoreContent("https://example.com/one", false)
	pageStorage.StoreContent("https://example.com/two", false)
	if err := pageStorage.Close(); err != nil {
		t.Fatalf("close output: %v", err)
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got, want := string(contents), "https://example.com/one\nhttps://example.com/two\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestJSONOutputUsesStructuredURLEntries(t *testing.T) {
	pageStorage := storage.NewPageStorage(true, -1)
	pageStorage.StoreSource("https://example.com/about", "href")
	pageStorage.StoreContent("https://example.com/about", false)
	pageStorage.StoreContent("https://example.com/standalone", false)

	filename := filepath.Join(t.TempDir(), "results.json")
	if err := pageStorage.WriteJSONToFile(filename); err != nil {
		t.Fatalf("write JSON output: %v", err)
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read JSON output: %v", err)
	}
	var output struct {
		URLs []storage.URLEntry `json:"urls"`
	}
	if err := json.Unmarshal(contents, &output); err != nil {
		t.Fatalf("unmarshal JSON output: %v", err)
	}
	want := []storage.URLEntry{
		{URL: "https://example.com/about", Source: "href"},
		{URL: "https://example.com/standalone"},
	}
	if !reflect.DeepEqual(output.URLs, want) {
		t.Errorf("URLs = %#v, want %#v", output.URLs, want)
	}

	var rawOutput struct {
		URLs []map[string]string `json:"urls"`
	}
	if err := json.Unmarshal(contents, &rawOutput); err != nil {
		t.Fatalf("unmarshal raw JSON output: %v", err)
	}
	if _, found := rawOutput.URLs[1]["source"]; found {
		t.Error("unknown source should be omitted from JSON output")
	}
}
