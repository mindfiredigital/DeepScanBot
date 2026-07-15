package testutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// RunCLI is a wrapper around CombinedOutputFor to facilitate simple CLI assertions.
// This matches the signature expected in your main_test.go.
func RunCLI(t *testing.T, binary string, dir string, args ...string) (string, error) {
	t.Helper()
	stdout, stderr, exitCode := CombinedOutputFor(t, binary, dir, args...)
	
	if exitCode != 0 {
		// Combine stdout and stderr so tests can assert against errors printed to either stream
		return stdout + stderr, fmt.Errorf("exit code %d", exitCode)
	}
	
	return stdout, nil
}

// NewTestServer sets up a local mock HTTP server for crawler testing
func NewTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><h1>DeepScanBot Test Page</h1></body></html>`))
	}))
}

// BuildCLI compiles the CLI binary safely with a 2-minute context deadline
func BuildCLI(t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine current file path")
	}

	cliDir := filepath.Join(filepath.Dir(currentFile), "..", "..")
	binary := filepath.Join(t.TempDir(), "deepscanbot")

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	cmd.Dir = cliDir

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("go build timed out after 2 minutes: %v", err)
		}
		t.Fatalf("go build failed: %v", err)
	}

	return binary
}

// CombinedOutputFor executes the binary with a 30-second bounded deadline context
func CombinedOutputFor(t *testing.T, binary string, dir string, args ...string) (string, string, int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("CLI command execution timed out after 30s: %v", err)
			return "", "", -1
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return stdout.String(), stderr.String(), exitErr.ExitCode()
		}
		t.Fatalf("command failed to run: %v", err)
	}

	return stdout.String(), stderr.String(), 0
}
