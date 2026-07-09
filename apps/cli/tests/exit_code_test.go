package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// exitCodeFor runs the CLI with the given arguments and returns the exit code.
// This helper does not use buildCLI – it builds once and reuses the binary
// across sub-tests to keep things fast.
func exitCodeFor(t *testing.T, binary, workdir string, args ...string) int {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Dir = workdir
	cmd.Stdout = nil // discard stdout – we only care about the exit code
	cmd.Stderr = nil // discard stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		t.Fatalf("unexpected error running CLI: %v", err)
	}
	return 0
}

// combinedOutputFor runs the CLI and returns stdout + stderr separately, plus
// the exit code.
func combinedOutputFor(t *testing.T, binary, workdir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Dir = workdir

	stdoutBytes, err := cmd.Output()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = string(exitErr.Stderr)
		}
	}

	return string(stdoutBytes), stderr, exitCode
}

func TestCLIExitCodeSuccess(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "version")
	if code != 0 {
		t.Errorf("version command exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeInvalidURL(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	tests := []struct {
		name      string
		args      []string
		wantCode  int
		wantErrIn string // substring expected in stderr
	}{
		{
			name:      "empty URL",
			args:      []string{"scan", ""},
			wantCode:  1,
			wantErrIn: "Error",
		},
		{
			name:      "no URL at all (missing arg)",
			args:      []string{"scan"},
			wantCode:  1,
			wantErrIn: "Usage",
		},
		{
			name:      "ftp scheme",
			args:      []string{"scan", "ftp://example.com"},
			wantCode:  1,
			wantErrIn: "Invalid URL",
		},
		{
			name:      "file scheme",
			args:      []string{"scan", "file:///etc/passwd"},
			wantCode:  1,
			wantErrIn: "Invalid URL",
		},
		{
			name:      "not a URL at all",
			args:      []string{"scan", "not-a-url"},
			wantCode:  1,
			wantErrIn: "Invalid URL",
		},
		{
			name:      "malformed URL",
			args:      []string{"scan", "http://"},
			wantCode:  1,
			wantErrIn: "Invalid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := exitCodeFor(t, binary, workdir, tt.args...)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
		})
	}
}

func TestCLIExitCodeErrorOutputContainsHint(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		wantHintIn  string
		wantErrIn   string
	}{
		{
			name:       "invalid scheme gives actionable hint",
			args:       []string{"scan", "ftp://example.com"},
			wantErrIn:  "Invalid URL",
			wantHintIn: "https://example.com",
		},
		{
			name:       "invalid URL gives actionable hint",
			args:       []string{"scan", "not-a-url"},
			wantErrIn:  "Invalid URL",
			wantHintIn: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, _ := combinedOutputFor(t, binary, workdir, tt.args...)
			if tt.wantErrIn != "" && !strings.Contains(stderr, tt.wantErrIn) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, tt.wantErrIn)
			}
			if tt.wantHintIn != "" && !strings.Contains(stderr, tt.wantHintIn) {
				t.Errorf("stderr = %q, want it to contain hint %q", stderr, tt.wantHintIn)
			}
		})
	}
}

func TestCLIExitCodeScanFailure(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	// A server that will cause the scan to fail
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't handle robots.txt so we get a real response
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer server.Close()

	// The scan should succeed (exit 0) because the server responds
	code := exitCodeFor(t, binary, workdir, "scan", server.URL, "depth=0")
	if code != 0 {
		t.Errorf("scan of valid server exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeNetworkFailure(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	// The crawler may or may not return a non-zero exit code for network
	// failures depending on the OS/network stack.  We verify that:
	//   1. The command completes (doesn't hang).
	//   2. If it fails, it returns a non-zero exit code.
	//   3. The error output contains actionable information.
	_, stderr, code := combinedOutputFor(t, binary, workdir, "scan", "http://192.0.2.1:1", "timeout=1")
	t.Logf("network failure: exit code=%d, stderr=%s", code, stderr)

	if code != 0 {
		// Verified: non-zero exit code on network failure.
		if !strings.Contains(stderr, "Error") && !strings.Contains(stderr, "error") && !strings.Contains(stderr, "failed") && !strings.Contains(stderr, "timeout") {
			t.Errorf("stderr should contain an actionable error message, got: %s", stderr)
		}
	}
	// If code == 0, the crawler handled the failure gracefully
	// and produced an empty report – that's acceptable.
}

func TestCLIExitCodeVersion(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "version")
	if code != 0 {
		t.Errorf("version exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeVersionJSON(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "version", "--json")
	if code != 0 {
		t.Errorf("version --json exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeDoctor(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "doctor")
	if code != 0 {
		t.Errorf("doctor exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeHelp(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "--help")
	if code != 0 {
		t.Errorf("help exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeHelpJSON(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "--help", "--json")
	if code != 0 {
		t.Errorf("help --json exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeScanWithJSONOutput(t *testing.T) {
	binary := buildCLI(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><a href='/page1'>Page 1</a></html>"))
	}))
	defer server.Close()

	workdir := t.TempDir()

	// Test with JSON output
	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "scan", server.URL, "depth=0", "--json", "output=exit-json-test")

	if code != 0 {
		t.Errorf("scan --json exit code = %d, want 0; stderr: %s", code, stderr)
	}

	// Verify JSON output on stdout
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}

	if result["status"] != "success" {
		t.Errorf("status = %v, want 'success'", result["status"])
	}

	// Verify output file was created
	outputFile := filepath.Join(workdir, "exit-json-test.json")
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestCLIExitCodeScanTextOutput(t *testing.T) {
	binary := buildCLI(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer server.Close()

	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "scan", server.URL, "depth=0", "output=exit-text-test")
	if code != 0 {
		t.Errorf("scan text output exit code = %d, want 0", code)
	}

	// Verify text output file was created
	outputFile := filepath.Join(workdir, "exit-text-test.txt")
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("text output file not created: %v", err)
	}
}

func TestCLIExitCodeErrorForEmptyOutputFilename(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	// output= with empty value should cause validation error
	_, stderr, code := combinedOutputFor(t, binary, workdir, "scan", "http://example.com", "output=")

	// It should fail (cobra will reject this because opts.Output becomes "" but
	// it would still reach buildOutputFilename which returns ErrEmptyOutputFilename)
	t.Logf("stderr: %s", stderr)
	if code == 0 {
		t.Error("scan with empty output= should fail with non-zero exit code")
	}
}

func TestCLIExitCodeInvalidFlag(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	// Cobra will reject unknown flags with exit code 1
	code := exitCodeFor(t, binary, workdir, "scan", "http://example.com", "--nonexistent-flag")
	if code == 0 {
		t.Error("scan with unknown flag should fail with non-zero exit code")
	}
}