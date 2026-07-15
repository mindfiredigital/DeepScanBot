package input_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mindfiredigital/DeepScanBot/packages/input"
)

func TestReadFromFile(t *testing.T) {
	// Create a temporary file with URLs
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "urls.txt")

	content := "https://example.com\nhttps://example.org\nhttps://example.net\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Read URLs from file
	urls, err := input.ReadInput(tmpFile, false)
	if err != nil {
		t.Fatalf("ReadInput() failed: %v", err)
	}

	// Verify results
	if len(urls) != 3 {
		t.Errorf("Expected 3 URLs, got %d", len(urls))
	}

	if urls[0] != "https://example.com" {
		t.Errorf("Expected first URL to be 'https://example.com', got '%s'", urls[0])
	}
	if urls[1] != "https://example.org" {
		t.Errorf("Expected second URL to be 'https://example.org', got '%s'", urls[1])
	}
	if urls[2] != "https://example.net" {
		t.Errorf("Expected third URL to be 'https://example.net', got '%s'", urls[2])
	}
}

func TestReadFromFileWithEmptyLines(t *testing.T) {
	// Create a temporary file with URLs and empty lines
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "urls.txt")

	content := "https://example.com\n\nhttps://example.org\n\n\nhttps://example.net\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Read URLs from file
	urls, err := input.ReadInput(tmpFile, false)
	if err != nil {
		t.Fatalf("ReadInput() failed: %v", err)
	}

	// Verify results (empty lines should be ignored)
	if len(urls) != 3 {
		t.Errorf("Expected 3 URLs, got %d", len(urls))
	}
}

func TestReadFromFileNotFound(t *testing.T) {
	// Try to read from a non-existent file
	_, err := input.ReadInput("/nonexistent/file.txt", false)
	if err == nil {
		t.Error("Expected error when reading non-existent file, got nil")
	}
}

func TestReadInputPrecedence(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "urls.txt")

	content := "https://example.com\nhttps://example.org\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// When both file and stdin are provided, file should take precedence
	urls, err := input.ReadInput(tmpFile, true) // true = use stdin, but file should take precedence
	if err != nil {
		t.Fatalf("ReadInput() failed: %v", err)
	}

	// Should read from file, not stdin
	if len(urls) != 2 {
		t.Errorf("Expected 2 URLs from file, got %d", len(urls))
	}
}

func TestReadInputNoInput(t *testing.T) {
	// When neither file nor stdin is requested, should return nil
	urls, err := input.ReadInput("", false)
	if err != nil {
		t.Fatalf("ReadInput() failed: %v", err)
	}

	if urls != nil {
		t.Errorf("Expected nil when no input specified, got %v", urls)
	}
}

func TestHasStdinData(t *testing.T) {
	// Verify the function doesn't panic
	hasData := input.HasStdinData()
	_ = hasData
}
