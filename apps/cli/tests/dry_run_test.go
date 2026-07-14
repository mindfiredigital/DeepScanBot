package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIDryRunPreview(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantIn     string
	}{
		{
			name:     "dry-run with scan previews without executing",
			args:     []string{"scan", server.URL, "depth=0", "--dry-run"},
			wantCode: 0,
			wantIn:   "Dry Run",
		},
		{
			name:     "dry-run with scan and json outputs plan",
			args:     []string{"scan", server.URL, "depth=0", "--dry-run", "--json"},
			wantCode: 0,
			wantIn:   `"action": "scan"`,
		},
		{
			name:     "dry-run with help succeeds",
			args:     []string{"--help", "--dry-run"},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := combinedOutputFor(t, binary, workdir, tt.args...)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d; stderr: %s", code, tt.wantCode, stderr)
			}

			if tt.wantIn != "" {
				combined := stdout + stderr
				if !strings.Contains(combined, tt.wantIn) {
					t.Errorf("output should contain %q, got stdout: %s, stderr: %s", tt.wantIn, stdout, stderr)
				}
			}
		})
	}
}

func TestCLIDryRunDoesNotCreateFiles(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	outputFile := filepath.Join(workdir, "crawler_results.txt")

	// Run with --dry-run
	_, stderr, code := combinedOutputFor(t, binary, workdir, "scan", server.URL, "depth=0", "output="+filepath.Base(outputFile)[:len(filepath.Base(outputFile))-4], "--dry-run")

	if code != 0 {
		t.Errorf("dry-run should succeed, got exit code %d; stderr: %s", code, stderr)
	}

	// Verify no output file was created
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Error("dry-run should not create output files")
	}
}

func TestCLIDryRunJSONOutput(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "scan", server.URL, "depth=0", "--dry-run", "--json")

	if code != 0 {
		t.Errorf("dry-run --json should succeed, got exit code %d; stderr: %s", code, stderr)
	}

	if !strings.Contains(stdout, `"status": "success"`) {
		t.Errorf("dry-run JSON should contain success status, got: %s", stdout)
	}

	if !strings.Contains(stdout, `"action": "scan"`) {
		t.Errorf("dry-run JSON should contain action field, got: %s", stdout)
	}

	if !strings.Contains(stdout, `"target_url"`) {
		t.Errorf("dry-run JSON should contain target_url field, got: %s", stdout)
	}
}

func TestCLIDryRunShowsExistingFileWarning(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	// Create an existing output file
	existingFile := filepath.Join(workdir, "crawler_results.txt")
	if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("create existing file: %v", err)
	}

	// Run with --dry-run
	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "scan", server.URL, "depth=0", "--dry-run")

	if code != 0 {
		t.Errorf("dry-run should succeed even with existing file, got exit code %d; stderr: %s", code, stderr)
	}

	combined := stdout + stderr
	if !strings.Contains(combined, "already exists") && !strings.Contains(combined, "overwritten") {
		t.Errorf("dry-run should warn about existing file, got: %s", combined)
	}

	// Verify the existing file was NOT modified
	data, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("read existing file: %v", err)
	}
	if string(data) != "existing" {
		t.Error("dry-run should not modify existing files")
	}
}

func TestCLIYesFlagAutoConfirms(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	// Create an existing output file
	existingFile := filepath.Join(workdir, "crawler_results.txt")
	if err := os.WriteFile(existingFile, []byte("old data"), 0644); err != nil {
		t.Fatalf("create existing file: %v", err)
	}

		// Run with --yes flag (should auto-confirm overwrite)
	// Note: --yes is a local flag on scan, so it must come before positional args
	_, stderr, code := combinedOutputFor(t, binary, workdir, "scan", "--yes", server.URL, "depth=0")

	if code != 0 {
		t.Errorf("--yes should auto-confirm and succeed, got exit code %d; stderr: %s", code, stderr)
	}

	// Verify the file was overwritten
	data, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("read overwritten file: %v", err)
	}
	if len(data) == 0 {
		t.Error("output file should have been overwritten")
	}
}

func TestCLIDryRunWithYesFlag(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	// --dry-run takes precedence, so --yes should be irrelevant
	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "scan", server.URL, "depth=0", "--dry-run", "--yes")

	if code != 0 {
		t.Errorf("dry-run with --yes should succeed, got exit code %d; stderr: %s", code, stderr)
	}

	// Combined output should mention dry run
	combined := stdout + stderr
	if !strings.Contains(combined, "Dry Run") {
		t.Errorf("output should mention dry run, got stdout: %s, stderr: %s", stdout, stderr)
	}
}

func TestCLIDryRunWithForceFlag(t *testing.T) {
	binary := buildCLI(t)
	workdir := t.TempDir()

	server := newTestServer()
	defer server.Close()

	// --dry-run takes precedence, so --force should be irrelevant
	stdout, stderr, code := combinedOutputFor(t, binary, workdir, "scan", server.URL, "depth=0", "--dry-run", "--force")

	if code != 0 {
		t.Errorf("dry-run with --force should succeed, got exit code %d; stderr: %s", code, stderr)
	}

	combined := stdout + stderr
	if !strings.Contains(combined, "Dry Run") {
		t.Errorf("output should mention dry run, got stdout: %s, stderr: %s", stdout, stderr)
	}
}
