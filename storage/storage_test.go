package storage

import (
	"os"
	"testing"
)

func TestTextOutputIsTruncatedForEachStorageInstance(t *testing.T) {
	const filename = "crawler_results.txt"
	t.Cleanup(func() {
		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove test output: %v", err)
		}
	})

	if err := os.WriteFile(filename, []byte("result from a previous crawl\n"), 0644); err != nil {
		t.Fatalf("seed previous output: %v", err)
	}

	storage := NewPageStorage(false, -1)
	storage.StoreContent("https://example.com/current", nil, false)
	storage.Close()

	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got, want := string(contents), "https://example.com/current\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
