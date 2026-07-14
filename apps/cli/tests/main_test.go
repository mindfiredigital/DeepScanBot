package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIRejectsInvalidStartURL(t *testing.T) {
	binary := buildCLI(t)

	for _, targetURL := range []string{"", "ftp://example.com", "file:///etc/passwd", "not-a-url", "http://", "http:/missing-slash.com"} {
		t.Run(targetURL, func(t *testing.T) {
			output, err := runCLI(binary, t.TempDir(), "scan", targetURL)
			if err == nil {
				t.Fatalf("CLI accepted invalid URL %q", targetURL)
			}

			if !strings.Contains(string(output), "URL") {
				t.Errorf("CLI output = %q, want an actionable URL error", output)
			}
		})
	}
}

func TestCLIConfiguresOutputFilename(t *testing.T) {
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
	if output, err := runCLI(binary, workdir, "scan", server.URL, "depth=0", "output=scan-results"); err != nil {
		t.Fatalf("run text output: %v\n%s", err, output)
	}

	if _, err := os.Stat(filepath.Join(workdir, "scan-results.txt")); err != nil {
		t.Fatalf("custom text output was not created: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workdir, "crawler_results.txt")); !os.IsNotExist(err) {
		t.Errorf("default text output should not be created: %v", err)
	}

	if output, err := runCLI(binary, workdir, "scan", server.URL, "depth=0", "json=true", "output=scan-json"); err != nil {
		t.Fatalf("run JSON output: %v\n%s", err, output)
	}

	if _, err := os.Stat(filepath.Join(workdir, "scan-json.json")); err != nil {
		t.Fatalf("custom JSON output was not created: %v", err)
	}
}

func TestCLIHelpDocumentsHelpFlag(t *testing.T) {
	binary := buildCLI(t)

	output, err := runCLI(binary, t.TempDir(), "--help")
	if err != nil {
		t.Fatalf("run help: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "--help") && !strings.Contains(string(output), "-h") {
		t.Errorf("help output does not document help flag: %s", output)
	}
}

func TestCLIVersionFlag(t *testing.T) {
	binary := buildCLI(t)

	// Test --version flag
	output, err := runCLI(binary, t.TempDir(), "--version")
	if err != nil {
		t.Fatalf("run --version: %v\n%s", err, output)
	}

	outputStr := string(output)
	// The --version flag uses cobra's built-in version which outputs to stdout
	if !strings.Contains(outputStr, "dev") {
		t.Errorf("--version output = %q, want to contain version 'dev' (or set via ldflags)", outputStr)
	}
}

func TestCLIVersionCommand(t *testing.T) {
	binary := buildCLI(t)

	// Test version subcommand
	output, err := runCLI(binary, t.TempDir(), "version")
	if err != nil {
		t.Fatalf("run version: %v\n%s", err, output)
	}

	outputStr := string(output)
	// Version subcommand uses logger, so it outputs to stderr with slog format
	if !strings.Contains(outputStr, "dev") {
		t.Errorf("version output = %q, want to contain version 'dev' (or set via ldflags)", outputStr)
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}

	binary := filepath.Join(t.TempDir(), "deepscanbot")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = repoRoot

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	return binary
}

func runCLI(binary, workdir string, args ...string) ([]byte, error) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = workdir

	return cmd.CombinedOutput()
}

func runCLIWithSeparateOutput(binary, workdir string, args ...string) (stdout, stderr []byte, err error) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = workdir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	stdout, err = readAll(stdoutPipe)
	if err != nil {
		return nil, nil, err
	}

	stderr, err = readAll(stderrPipe)
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Wait(); err != nil {
		// Check if it's an exit error
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, stderr, exitErr
		}
		return stdout, stderr, err
	}

	return stdout, stderr, nil
}

func readAll(pipe io.ReadCloser) ([]byte, error) {
	defer pipe.Close()
	return io.ReadAll(pipe)
}
