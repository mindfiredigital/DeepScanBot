package storage

import (
	"bytes"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to a file atomically using a temporary file and rename.
// This ensures that the file is either completely written or not modified at all,
// preventing corruption if the operation is interrupted.
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := EnsureDirectory(dir); err != nil {
		return err
	}

	// Create a temporary file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		// Clean up temp file if it still exists
		os.Remove(tmpPath)
	}()

	// Write data to temporary file
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomically rename temp file to target file
	if err := os.Rename(tmpPath, filename); err != nil {
		return err
	}

	return nil
}

// WriteFileAtomicString writes a string to a file atomically.
func WriteFileAtomicString(filename string, content string, perm os.FileMode) error {
	return WriteFileAtomic(filename, []byte(content), perm)
}

// FileExists checks if a file exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// FileContentEquals checks if a file's content matches the given bytes.
// Returns false if the file doesn't exist.
func FileContentEquals(filename string, expected []byte) bool {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	return bytes.Equal(content, expected)
}

// EnsureDirectory creates a directory if it doesn't exist.
func EnsureDirectory(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}