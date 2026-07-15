package testutil

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// NewTestServer spins up a local HTTP server with a default mock response for CLI testing
func NewTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><body><h1>DeepScanBot Target</h1><a href="/child">Child Link</a></body></html>`))
	}))
}

// BuildCLI builds the CLI binary for testing by locating the main application package.
func BuildCLI(t *testing.T) string {
	t.Helper()

	// Get the absolute path of this helper file dynamically
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to identify source file path path using runtime.Caller")
	}

	// Move up two directories: testutil/ -> tests/ -> cli/ (where main.go lives)
	cliDir := filepath.Join(filepath.Dir(currentFile), "..", "..")

	binary := filepath.Join(t.TempDir(), "deepscanbot")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = cliDir

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	return binary
}

// CombinedOutputFor executes the binary within a specific working directory, capturing
// stdout, stderr, and extraction of the numeric exit code.
func CombinedOutputFor(t *testing.T, binary, workdir string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = workdir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run command: %v", err)
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// RunCLI runs the CLI and returns combined raw bytes output (kept for backward compatibility)
func RunCLI(binary, workdir string, args ...string) ([]byte, error) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = workdir
	return cmd.CombinedOutput()
}

// RunCLIWithSeparateOutput runs the CLI and returns stdout, stderr separately as bytes.
// It uses in-memory buffers to safely avoid pipe deadlocks.
func RunCLIWithSeparateOutput(binary, workdir string, args ...string) ([]byte, []byte, error) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = workdir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	return stdoutBuf.Bytes(), stderrBuf.Bytes(), err
}
