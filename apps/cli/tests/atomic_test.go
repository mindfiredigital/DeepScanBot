package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mindfiredigital/DeepScanBot/packages/storage"
)

// Helper functions to wrap the storage package functions
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	return storage.WriteFileAtomic(filename, data, perm)
}

func FileExists(filename string) bool {
	return storage.FileExists(filename)
}

func FileContentEquals(filename string, content []byte) bool {
	return storage.FileContentEquals(filename, content)
}

func EnsureDirectory(path string) error {
	return storage.EnsureDirectory(path)
}

func TestWriteFileAtomic(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	data := []byte("Hello, World!")
	if err := WriteFileAtomic(filename, data, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic() failed: %v", err)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("Expected %q, got %q", string(data), string(content))
	}
}

func TestWriteFileAtomicCreatesDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "subdir", "test.txt")

	// Write should create the directory
	data := []byte("test data")
	if err := WriteFileAtomic(filename, data, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic() failed: %v", err)
	}

	// Verify file exists
	if !FileExists(filename) {
		t.Error("File was not created")
	}
}

func TestWriteFileAtomicOverwrites(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	initialData := []byte("initial content")
	if err := WriteFileAtomic(filename, initialData, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic() failed: %v", err)
	}

	// Overwrite with new content
	newData := []byte("new content")
	if err := WriteFileAtomic(filename, newData, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic() failed on overwrite: %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(newData) {
		t.Errorf("Expected %q, got %q", string(newData), string(content))
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "exists.txt")

	if err := os.WriteFile(filename, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// File should exist
	if !FileExists(filename) {
		t.Error("FileExists() returned false for existing file")
	}

	// Non-existent file should not exist
	nonExistent := filepath.Join(tmpDir, "nonexistent.txt")
	if FileExists(nonExistent) {
		t.Error("FileExists() returned true for non-existent file")
	}
}

func TestFileContentEquals(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(filename, content, 0o644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Should match
	if !FileContentEquals(filename, content) {
		t.Error("FileContentEquals() returned false for matching content")
	}

	// Should not match different content
	if FileContentEquals(filename, []byte("different content")) {
		t.Error("FileContentEquals() returned true for different content")
	}

	// Should not match non-existent file
	if FileContentEquals(filepath.Join(tmpDir, "nonexistent.txt"), content) {
		t.Error("FileContentEquals() returned true for non-existent file")
	}
}

func TestEnsureDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Test creating a new directory
	newDir := filepath.Join(tmpDir, "new", "nested", "dir")
	if err := EnsureDirectory(newDir); err != nil {
		t.Fatalf("EnsureDirectory() failed: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Path is not a directory")
	}

	// Test that calling again doesn't error
	if err := EnsureDirectory(newDir); err != nil {
		t.Fatalf("EnsureDirectory() failed on existing directory: %v", err)
	}
}

func TestWriteFileAtomicIdempotency(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "idempotent.txt")

	data := []byte("idempotent content")

	// Write the same content multiple times
	for i := 0; i < 5; i++ {
		if err := WriteFileAtomic(filename, data, 0o644); err != nil {
			t.Fatalf("WriteFileAtomic() failed on iteration %d: %v", i, err)
		}

		// Verify content is correct
		if !FileContentEquals(filename, data) {
			t.Errorf("Content mismatch on iteration %d", i)
		}
	}
}
