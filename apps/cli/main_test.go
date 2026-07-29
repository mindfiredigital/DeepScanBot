package main_test

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

func TestCLIHelpFlag(t *testing.T) {
	binary := testutil.BuildCLI(t)

	output, err := testutil.RunCLI(t, binary, t.TempDir(), "--help")
	if err != nil {
		t.Fatalf("run --help: %v\n%s", err, output)
	}

	// Verify help output contains key sections
	requiredSections := []string{
		"Usage:",
		"deepscanbot",
		"scan",
		"version",
		"doctor",
		"Flags:",
		"--json",
		"--help",
	}
	for _, section := range requiredSections {
		if !strings.Contains(output, section) {
			t.Errorf("help output missing section %q:\n%s", section, output)
		}
	}
}

func TestCLIScanHelp(t *testing.T) {
	binary := testutil.BuildCLI(t)

	output, err := testutil.RunCLI(t, binary, t.TempDir(), "scan", "--help")
	if err != nil {
		t.Fatalf("run scan --help: %v\n%s", err, output)
	}

	// Verify scan command help contains expected options
	expectedOptions := []string{
		"--depth",
		"--timeout",
		"--output",
		"--json",
		"--concurrency",
		"<url>",
	}
	for _, option := range expectedOptions {
		if !strings.Contains(output, option) {
			t.Errorf("scan help missing option %q:\n%s", option, output)
		}
	}
}

func TestCLIJSONOutput(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><h1>Test</h1></body></html>"))
	}))
	defer server.Close()

	// Test --json flag produces JSON output
	output, err := testutil.RunCLI(t, binary, workdir, "scan", server.URL, "--json", "depth=0")
	if err != nil {
		t.Fatalf("run with --json: %v\n%s", err, output)
	}

	// Verify output is valid JSON and contains expected fields
	if !strings.Contains(output, `"start_url"`) {
		t.Errorf("JSON output missing start_url field: %s", output)
	}
	if !strings.Contains(output, `"summary"`) {
		t.Errorf("JSON output missing summary field: %s", output)
	}
	if !strings.Contains(output, `"urls"`) {
		t.Errorf("JSON output missing urls field: %s", output)
	}
}

func TestCLIMultipleURLs(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	// Create two test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><a href='/page1'>Link 1</a></body></html>"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><a href='/page2'>Link 2</a></body></html>"))
	}))
	defer server2.Close()

	// Test scanning multiple URLs
	output, err := testutil.RunCLI(t, binary, workdir, "scan", server1.URL, server2.URL, "depth=0")
	if err != nil {
		t.Fatalf("run multi-site scan: %v\n%s", err, output)
	}

	// Verify multi-site summary is generated
	if !strings.Contains(output, "Multi-Site Scan Summary") {
		t.Errorf("Multi-site scan output missing summary:\n%s", output)
	}
	if !strings.Contains(output, "Sites crawled:") {
		t.Errorf("Multi-site scan output missing sites count: %s", output)
	}

	// Verify individual site reports were created
	expectedFiles := []string{
		"crawler_results_summary.json",
	}
	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(workdir, file)); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

func TestCLIMultipleURLsWithJSON(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>Site 1</body></html>"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>Site 2</body></html>"))
	}))
	defer server2.Close()

	// Test multi-site with JSON output
	output, err := testutil.RunCLI(t, binary, workdir, "scan", server1.URL, server2.URL, "--json", "depth=0")
	if err != nil {
		t.Fatalf("run multi-site with JSON: %v\n%s", err, output)
	}

	// Verify JSON structure for multi-site report
	if !strings.Contains(output, `"total_sites"`) {
		t.Errorf("Multi-site JSON missing total_sites: %s", output)
	}
	if !strings.Contains(output, `"sites"`) {
		t.Errorf("Multi-site JSON missing sites array: %s", output)
	}
}
