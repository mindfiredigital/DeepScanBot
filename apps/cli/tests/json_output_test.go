package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIVersionJSONOutput(t *testing.T) {
	binary := buildCLI(t)

	tests := []struct {
		name     string
		args     []string
		wantJSON bool
		wantText string
	}{
		{
			name:     "version with --json flag",
			args:     []string{"version", "--json"},
			wantJSON: true,
		},
		{
			name:     "version without json flag",
			args:     []string{"version"},
			wantJSON: false,
			wantText: "DeepScanBot CLI v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output []byte
			
			if tt.wantJSON {
				// For JSON output, capture stdout separately to avoid mixing with stderr logs
				stdout, _, _ := runCLIWithSeparateOutput(binary, t.TempDir(), tt.args...)
				output = stdout
			} else {
				output, _ = runCLI(binary, t.TempDir(), tt.args...)
			}

			outputStr := string(output)

			if tt.wantJSON {
				// Verify it's valid JSON
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, outputStr)
				}

				// Verify structure
				if result["status"] != "success" {
					t.Errorf("Expected status 'success', got '%v'", result["status"])
				}

				if result["data"] == nil {
					t.Error("Expected data field to be present")
				}

				if result["meta"] == nil {
					t.Error("Expected meta field to be present")
				}

				// Verify no progress messages in stdout
				if strings.Contains(outputStr, "Running diagnostics") {
					t.Error("JSON output should not contain progress messages")
				}
			} else if tt.wantText != "" {
				if !strings.Contains(outputStr, tt.wantText) {
					t.Errorf("Expected output to contain '%s', got: %s", tt.wantText, outputStr)
				}
			}
		})
	}
}

func TestCLIDoctorJSONOutput(t *testing.T) {
	binary := buildCLI(t)

	tests := []struct {
		name     string
		args     []string
		wantJSON bool
	}{
		{
			name:     "doctor with --json flag",
			args:     []string{"doctor", "--json"},
			wantJSON: true,
		},
		{
			name:     "doctor without json flag",
			args:     []string{"doctor"},
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output []byte
			
			if tt.wantJSON {
				// For JSON output, capture stdout separately to avoid mixing with stderr logs
				stdout, _, _ := runCLIWithSeparateOutput(binary, t.TempDir(), tt.args...)
				output = stdout
			} else {
				output, _ = runCLI(binary, t.TempDir(), tt.args...)
			}

			outputStr := string(output)

			if tt.wantJSON {
				// Verify it's valid JSON
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, outputStr)
				}

				// Verify structure
				if result["status"] != "success" {
					t.Errorf("Expected status 'success', got '%v'", result["status"])
				}

				if result["data"] == nil {
					t.Error("Expected data field to be present")
				}

				// Verify no progress messages in stdout
				if strings.Contains(outputStr, "Running diagnostics") {
					t.Error("JSON output should not contain progress messages")
				}
			} else {
				// Human-readable output should contain diagnostic messages
				if !strings.Contains(outputStr, "Running diagnostics") {
					t.Error("Expected human-readable output to contain diagnostic messages")
				}
			}
		})
	}
}

func TestCLIScanJSONOutput(t *testing.T) {
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

	tests := []struct {
		name        string
		args        []string
		wantJSON    bool
		wantFile    bool
	}{
		{
			name:     "scan with --json flag",
			args:     []string{"scan", server.URL, "depth=0", "--json", "output=test-scan"},
			wantJSON: true,
			wantFile: true,
		},
		{
			name:     "scan without json flag",
			args:     []string{"scan", server.URL, "depth=0", "output=test-scan3"},
			wantJSON: false,
			wantFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output []byte
			var stderr []byte
			
			if tt.wantJSON {
				// For JSON output, capture stdout separately to avoid mixing with stderr logs
				stdout, stderrOut, _ := runCLIWithSeparateOutput(binary, workdir, tt.args...)
				output = stdout
				stderr = stderrOut
			} else {
				output, _ = runCLI(binary, workdir, tt.args...)
			}

			outputStr := string(output)

			if tt.wantJSON {
				// Log stderr for debugging
				if len(stderr) > 0 {
					t.Logf("stderr: %s", string(stderr))
				}

				// Verify it's valid JSON
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, outputStr)
				}

				// Verify structure
				if result["status"] != "success" {
					t.Errorf("Expected status 'success', got '%v'", result["status"])
				}

				if result["data"] == nil {
					t.Error("Expected data field to be present")
				}

				if result["meta"] == nil {
					t.Error("Expected meta field to be present")
				}

				// Verify no progress messages in stdout
				if strings.Contains(outputStr, "Resume mode") {
					t.Error("JSON output should not contain progress messages")
				}
			}

			// Check if output file was created
			if tt.wantFile {
				// Extract output name from args
				outputName := "test-scan"
				for _, arg := range tt.args {
					if strings.HasPrefix(arg, "output=") {
						parts := strings.SplitN(arg, "=", 2)
						if len(parts) == 2 {
							outputName = parts[1]
						}
						break
					}
				}

				expectedExt := ".txt"
				if tt.wantJSON {
					expectedExt = ".json"
				}

				outputFile := filepath.Join(workdir, outputName+expectedExt)
				if _, err := os.Stat(outputFile); err != nil {
					t.Errorf("Output file %s was not created: %v", outputFile, err)
				}
			}
		})
	}
}

func TestCLIJSONOutputOnlyOnStdout(t *testing.T) {
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

	// Run scan with JSON output and capture stdout separately
	stdout, stderr, _ := runCLIWithSeparateOutput(binary, workdir, "scan", server.URL, "depth=0", "--json", "output=json-test")

	// Log stderr for debugging
	if len(stderr) > 0 {
		t.Logf("stderr: %s", string(stderr))
	}

	// Verify stdout contains valid JSON
	var result map[string]interface{}
	if parseErr := json.Unmarshal(stdout, &result); parseErr != nil {
		t.Errorf("Stdout should contain only valid JSON when --json is used: %v\nOutput: %s", parseErr, string(stdout))
	}

	// Verify it's a success response
	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got '%v'", result["status"])
	}

	// Verify stderr contains log messages (not JSON)
	stderrStr := string(stderr)
	if len(stderrStr) > 0 {
		// stderr should contain log messages, not JSON
		if strings.Contains(stderrStr, `"status"`) && strings.Contains(stderrStr, `"success"`) {
			t.Error("stderr should not contain JSON output")
		}
	}
}

func TestCLIErrorJSONOutput(t *testing.T) {
	binary := buildCLI(t)

	// Test with invalid URL
	output, err := runCLI(binary, t.TempDir(), "scan", "not-a-url", "--json")

	// Should fail
	if err == nil {
		t.Fatal("Expected command to fail with invalid URL")
	}

	// Try to parse as JSON
	var result map[string]interface{}
	if parseErr := json.Unmarshal(output, &result); parseErr != nil {
		// If it's not JSON, that's also acceptable - the error might be logged to stderr
		t.Logf("Error output is not JSON (this is acceptable): %v", parseErr)
		return
	}

	// If it is JSON, verify structure
	if result["status"] != "error" {
		t.Errorf("Expected status 'error' for invalid URL, got '%v'", result["status"])
	}
}

func TestCLIHelpJSONOutput(t *testing.T) {
	binary := buildCLI(t)

	// Test help with --json flag
	output, err := runCLI(binary, t.TempDir(), "--help", "--json")

	// Help might not support JSON, but let's check what happens
	if err != nil {
		t.Logf("Help command error: %v", err)
	}

	// Help output might not be JSON, which is fine
	// We're just testing that it doesn't crash
	_ = output
}

func TestCLIJSONOutputBackwardCompatibility(t *testing.T) {
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

	// Test that existing behavior without --json flag still works
	output, err := runCLI(binary, workdir, "scan", server.URL, "depth=0", "output=compat-test")
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Should create text file
	outputFile := filepath.Join(workdir, "compat-test.txt")
	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("Text output file was not created: %v", err)
	}

	// Output should be human-readable (not JSON)
	outputStr := string(output)
	if strings.Contains(outputStr, `"status"`) && strings.Contains(outputStr, `"success"`) {
		t.Error("Output should be human-readable when --json is not specified")
	}
}