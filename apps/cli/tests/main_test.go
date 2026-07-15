package cli_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindfiredigital/DeepScanBot/apps/cli/tests/testutil"
)

func TestCLIRejectsInvalidStartURL(t *testing.T) {
	binary := testutil.BuildCLI(t)

	for _, targetURL := range []string{"", "ftp://example.com", "file:///etc/passwd", "not-a-url", "http://", "http:/missing-slash.com"} {
		t.Run(targetURL, func(t *testing.T) {
			output, err := testutil.RunCLI(t, binary, t.TempDir(), "scan", targetURL)
			if err == nil {
				t.Fatalf("CLI accepted invalid URL %q", targetURL)
			}

			if !strings.Contains(output, "URL") {
				t.Errorf("CLI output = %q, want an actionable URL error", output)
			}
		})
	}
}

func TestCLIConfiguresOutputFilename(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer server.Close()

	// Test custom text output
	if output, err := testutil.RunCLI(t, binary, workdir, "scan", server.URL, "depth=0", "output=scan-results"); err != nil {
		t.Fatalf("run text output: %v\n%s", err, output)
	}

	if _, err := os.Stat(filepath.Join(workdir, "scan-results.txt")); err != nil {
		t.Fatalf("custom text output was not created: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workdir, "crawler_results.txt")); !os.IsNotExist(err) {
		t.Errorf("default text output should not be created: %v", err)
	}

	// Test custom JSON output
	if output, err := testutil.RunCLI(t, binary, workdir, "scan", server.URL, "depth=0", "json=true", "output=scan-json"); err != nil {
		t.Fatalf("run JSON output: %v\n%s", err, output)
	}

	if _, err := os.Stat(filepath.Join(workdir, "scan-json.json")); err != nil {
		t.Fatalf("custom JSON output was not created: %v", err)
	}
}

func TestCLIHelpDocumentsHelpFlag(t *testing.T) {
	binary := testutil.BuildCLI(t)

	output, err := testutil.RunCLI(t, binary, t.TempDir(), "--help")
	if err != nil {
		t.Fatalf("run help: %v\n%s", err, output)
	}

	if !strings.Contains(output, "--help") && !strings.Contains(output, "-h") {
		t.Errorf("help output does not document help flag: %s", output)
	}
}

func TestCLIVersionFlag(t *testing.T) {
	binary := testutil.BuildCLI(t)

	output, err := testutil.RunCLI(t, binary, t.TempDir(), "--version")
	if err != nil {
		t.Fatalf("run --version: %v\n%s", err, output)
	}

	if !strings.Contains(output, "dev") {
		t.Errorf("--version output = %q, want version 'dev'", output)
	}
}

func TestCLIVersionCommand(t *testing.T) {
	binary := testutil.BuildCLI(t)

	output, err := testutil.RunCLI(t, binary, t.TempDir(), "version")
	if err != nil {
		t.Fatalf("run version: %v\n%s", err, output)
	}

	if !strings.Contains(output, "dev") {
		t.Errorf("version output = %q, want version 'dev'", output)
	}
}
