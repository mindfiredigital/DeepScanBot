package testutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

var (
	buildOnce sync.Once
	cachedBin string
	buildErr  error
	buildDir  string
)

func init() {
	// Create a stable temporary directory for the cached binary.
	// This directory lives for the duration of the test process.
	buildDir, buildErr = os.MkdirTemp("", "deepscanbot-build-*")
	if buildErr != nil {
		return
	}
}

// RunCLI is a wrapper around CombinedOutputFor to facilitate simple CLI assertions.
func RunCLI(t *testing.T, binary string, dir string, args ...string) (string, error) {
	t.Helper()
	stdout, stderr, code := CombinedOutputFor(t, binary, dir, args...)

	// Combine stdout and stderr into one string so tests can assert against the full output
	fullOutput := stdout + stderr

	if code != 0 {
		return fullOutput, fmt.Errorf("CLI failed with exit code %d: %s", code, stderr)
	}
	return fullOutput, nil
}

// NewTestServer sets up a local mock HTTP server for crawler testing
func NewTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><h1>DeepScanBot Test Page</h1></body></html>`))
	}))
}

// BuildCLI compiles the CLI binary exactly once using sync.Once, storing the
// binary in a stable temp directory shared across all callers.  Subsequent
// calls return the cached path.  If the one-time build fails, all callers see
// the same fatal error immediately.
func BuildCLI(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			buildErr = errors.New("could not determine current file path")
			return
		}

		cliDir := filepath.Join(filepath.Dir(currentFile), "..", "..")
		cachedBin = filepath.Join(buildDir, "deepscanbot")

		cmd := exec.CommandContext(ctx, "go", "build", "-o", cachedBin, ".")
		cmd.Dir = cliDir

		if err := cmd.Run(); err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				buildErr = errors.New("go build timed out after 2 minutes")
				return
			}
			buildErr = err
			return
		}
	})

	if buildErr != nil {
		t.Fatalf("BuildCLI: %v", buildErr)
	}

	return cachedBin
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
