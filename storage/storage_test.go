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
	if err := storage.Close(); err != nil {
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

func TestTextOutputUsesBufferedWriterUntilClose(t *testing.T) {
	const filename = "crawler_results.txt"
	t.Cleanup(func() {
		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove test output: %v", err)
		}
	})

	storage := NewPageStorage(false, -1)
	if storage.file == nil || storage.writer == nil {
		t.Fatal("text output should open one file and create one buffered writer")
	}
	file, writer := storage.file, storage.writer

	storage.StoreContent("https://example.com/one", nil, false)
	storage.StoreContent("https://example.com/two", nil, false)
	if storage.file != file || storage.writer != writer {
		t.Error("writes should reuse the original file and buffered writer")
	}
	if writer.Buffered() == 0 {
		t.Error("expected output to remain buffered until close")
	}
	if err := storage.Close(); err != nil {
		t.Fatalf("close output: %v", err)
	}
	if storage.file != nil || storage.writer != nil {
		t.Error("close should release the file and buffered writer")
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got, want := string(contents), "https://example.com/one\nhttps://example.com/two\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
