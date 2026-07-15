package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLINoInputFlag(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	// Create an existing output file to test the overwrite guard
	existingFile := filepath.Join(workdir, "crawler_results.txt")
	if err := os.WriteFile(existingFile, []byte("existing"), 0o644); err != nil {
		t.Fatalf("create existing file: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		wantCode  int
		wantErrIn string
	}{
		{
			name:      "no-input flag with missing URL fails with invalid input",
			args:      []string{"--no-input", "scan", "http://192.0.2.1:1"},
			wantCode:  1,
			wantErrIn: "Error",
		},
		{
			name:     "no-input flag with version succeeds",
			args:     []string{"--no-input", "version"},
			wantCode: 0,
		},
		{
			name:     "no-input flag with doctor succeeds",
			args:     []string{"--no-input", "doctor"},
			wantCode: 0,
		},
		{
			name:     "no-input flag with --help succeeds",
			args:     []string{"--no-input", "--help"},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			cmd.Dir = workdir

			err := cmd.Run()
			code := 0
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					code = exitErr.ExitCode()
				}
			}

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
		})
	}
}

func TestCLINoInputPreventsOverwrite(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	// Create a server that returns valid HTML for the scan
	server := newTestServer()
	defer server.Close()

	// Create an existing output file
	existingFile := filepath.Join(workdir, "crawler_results.txt")
	if err := os.WriteFile(existingFile, []byte("existing data"), 0o644); err != nil {
		t.Fatalf("create existing file: %v", err)
	}

	// Run scan with --no-input but without --force: should fail because
	// the output file already exists.
	_, stderr, code := combinedOutputFor(t, binary, workdir, "--no-input", "scan", server.URL, "depth=0")

	if code == 0 {
		t.Error("Expected non-zero exit code when output file exists in --no-input mode without --force")
	}

	if !strings.Contains(stderr, "already exists") {
		t.Errorf("stderr should mention existing file, got: %s", stderr)
	}

	// Verify the existing file was NOT overwritten
	data, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("read existing file: %v", err)
	}
	if string(data) != "existing data" {
		t.Errorf("existing file was overwritten: got %q, want %q", string(data), "existing data")
	}
}

func TestCLINoInputWithForceOverwrites(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	// Create a server that returns valid HTML for the scan
	server := newTestServer()
	defer server.Close()

	// Create an existing output file
	existingFile := filepath.Join(workdir, "crawler_results.txt")
	if err := os.WriteFile(existingFile, []byte("old data"), 0o644); err != nil {
		t.Fatalf("create existing file: %v", err)
	}

	// Run scan with --no-input AND --force: should succeed and overwrite
	_, stderr, code := combinedOutputFor(t, binary, workdir, "--no-input", "scan", server.URL, "depth=0", "--force")

	if code != 0 {
		t.Errorf("expected exit code 0 with --force, got %d; stderr: %s", code, stderr)
	}

	// Verify the file was overwritten with scan results
	data, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("read overwritten file: %v", err)
	}
	if len(data) == 0 {
		t.Error("output file was overwritten but is empty")
	}
	if string(data) == "old data" {
		t.Error("output file was not overwritten")
	}
}

func TestCLINoInputScanWithOutputFlag(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	// Run scan with --no-input, using a custom output name that doesn't exist yet
	outputName := filepath.Join(workdir, "custom-output.txt")
	_, stderr, code := combinedOutputFor(t, binary, workdir, "--no-input", "scan", server.URL, "depth=0", "output="+outputName[:len(outputName)-4])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// Verify the custom output file was created
	if _, err := os.Stat(outputName); err != nil {
		t.Errorf("custom output file was not created: %v", err)
	}
}

func TestCLINoInputVersionJSON(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "--no-input", "version", "--json")

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	if !strings.Contains(stdout, `"status": "success"`) {
		t.Errorf("stdout should contain success JSON, got: %s", stdout)
	}
}

func TestCLINoInputDoctorJSON(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "--no-input", "doctor", "--json")

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	if !strings.Contains(stdout, `"status": "success"`) {
		t.Errorf("stdout should contain success JSON, got: %s", stdout)
	}
}

func TestCLINoInputHelpJSON(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "--no-input", "--help", "--json")

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	if !strings.Contains(stdout, `"status": "success"`) {
		t.Errorf("stdout should contain success JSON, got: %s", stdout)
	}
}

// newTestServer creates a simple HTTP test server that responds with valid
// HTML for scanning.
func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
}
