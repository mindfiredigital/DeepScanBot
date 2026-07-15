package cli_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindfiredigital/DeepScanBot/apps/cli/tests/testutil"
)

// Helper to run the CLI and return combined stdout and stderr
func runCLI(t *testing.T, binary, workdir string, args ...string) ([]byte, error) {
	t.Helper()
	stdout, stderr, code := testutil.CombinedOutputFor(t, binary, workdir, args...)
	var err error
	if code != 0 {
		err = fmt.Errorf("cli exited with code %d", code)
	}
	return []byte(stdout + stderr), err
}

// Helper to run the CLI and return separated stdout and stderr streams
func runCLIWithSeparateOutput(t *testing.T, binary, workdir string, args ...string) ([]byte, []byte, error) {
	t.Helper()
	stdout, stderr, code := testutil.CombinedOutputFor(t, binary, workdir, args...)
	var err error
	if code != 0 {
		err = fmt.Errorf("cli exited with code %d", code)
	}
	return []byte(stdout), []byte(stderr), err
}

func TestCLIVersionJSONOutput(t *testing.T) {
	binary := testutil.BuildCLI(t)

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
			wantText: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output []byte

			if tt.wantJSON {
				// For JSON output, capture stdout separately to avoid mixing with stderr logs
				stdout, _, _ := runCLIWithSeparateOutput(t, binary, t.TempDir(), tt.args...)
				output = stdout
			} else {
				output, _ = runCLI(t, binary, t.TempDir(), tt.args...)
			}

			outputStr := string(output)

			if tt.wantJSON {
				// Verify it's valid JSON
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, outputStr)
				}

				// Verify schemaVersion exists and is correct
				schemaVersion, ok := result["schemaVersion"].(string)
				if !ok {
					t.Error("Expected schemaVersion field to be present and a string")
				} else if schemaVersion != "1.0.0" {
					t.Errorf("Expected schemaVersion '1.0.0', got '%s'", schemaVersion)
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
	binary := testutil.BuildCLI(t)

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
				stdout, _, _ := runCLIWithSeparateOutput(t, binary, t.TempDir(), tt.args...)
				output = stdout
			} else {
				output, _ = runCLI(t, binary, t.TempDir(), tt.args...)
			}

			outputStr := string(output)

			if tt.wantJSON {
				// Verify it's valid JSON
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, outputStr)
				}

				// Verify schemaVersion exists and is correct
				schemaVersion, ok := result["schemaVersion"].(string)
				if !ok {
					t.Error("Expected schemaVersion field to be present and a string")
				} else if schemaVersion != "1.0.0" {
					t.Errorf("Expected schemaVersion '1.0.0', got '%s'", schemaVersion)
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
			} else if !strings.Contains(outputStr, "Running diagnostics") {
				// Human-readable output should contain diagnostic messages
				t.Error("Expected human-readable output to contain diagnostic messages")
			}
		})
	}
}

func TestCLIScanJSONOutput(t *testing.T) {
	binary := testutil.BuildCLI(t)

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
		name     string
		args     []string
		wantJSON bool
		wantFile bool
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
				stdout, stderrOut, _ := runCLIWithSeparateOutput(t, binary, workdir, tt.args...)
				output = stdout
				stderr = stderrOut
			} else {
				output, _ = runCLI(t, binary, workdir, tt.args...)
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

				// Verify schemaVersion exists and is correct
				schemaVersion, ok := result["schemaVersion"].(string)
				if !ok {
					t.Error("Expected schemaVersion field to be present and a string")
				} else if schemaVersion != "1.0.0" {
					t.Errorf("Expected schemaVersion '1.0.0', got '%s'", schemaVersion)
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
	binary := testutil.BuildCLI(t)

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
	stdout, stderr, _ := runCLIWithSeparateOutput(t, binary, workdir, "scan", server.URL, "depth=0", "--json", "output=json-test")

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
		if strings.Contains(stderrStr, `"status"`) && strings.Contains(stderrStr, `"success"`) {
			t.Error("stderr should not contain JSON output")
		}
	}
}

func TestCLIErrorJSONOutput(t *testing.T) {
	binary := testutil.BuildCLI(t)

	// Test with invalid URL
	output, err := runCLI(t, binary, t.TempDir(), "scan", "not-a-url", "--json")

	// Should fail
	if err == nil {
		t.Fatal("Expected command to fail with invalid URL")
	}

	// Try to parse as JSON
	var result map[string]interface{}
	if parseErr := json.Unmarshal(output, &result); parseErr != nil {
		t.Logf("Error output is not JSON (this is acceptable): %v", parseErr)
		return
	}

	// If it is JSON, verify structure
	if result["status"] != "error" {
		t.Errorf("Expected status 'error' for invalid URL, got '%v'", result["status"])
	}
}

func TestCLIHelpJSONOutput(t *testing.T) {
	binary := testutil.BuildCLI(t)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help with --json flag",
			args:    []string{"--help", "--json"},
			wantErr: false,
		},
		{
			name:    "help -h with --json flag",
			args:    []string{"-h", "--json"},
			wantErr: false,
		},
		{
			name:    "subcommand help with --json",
			args:    []string{"scan", "--help", "--json"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, binary, t.TempDir(), tt.args...)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, output)
			}

			// Verify it's valid JSON
			var result map[string]interface{}
			if parseErr := json.Unmarshal(output, &result); parseErr != nil {
				t.Fatalf("Output is not valid JSON: %v\nOutput: %s", parseErr, string(output))
			}

			// Verify response structure
			if result["status"] != "success" {
				t.Errorf("Expected status 'success', got '%v'", result["status"])
			}

			if result["meta"] == nil {
				t.Error("Expected meta field to be present")
			}

			// Verify command tree data structure
			data, ok := result["data"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected data to be an object")
			}

			// Verify top-level tree fields
			if data["cli_name"] == nil {
				t.Error("Expected cli_name in command tree")
			}

			rootCmd, ok := data["root_command"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected root_command in command tree")
			}

			if rootCmd["use"] == nil {
				t.Error("Expected use field in root command")
			}

			if rootCmd["short"] == nil {
				t.Error("Expected short field in root command")
			}

			// Verify subcommands exist
			subs, ok := rootCmd["subcommands"].([]interface{})
			if !ok || len(subs) == 0 {
				t.Error("Expected subcommands in root command")
			}

			// Verify flags exist
			flags, ok := rootCmd["flags"].([]interface{})
			if !ok || len(flags) == 0 {
				t.Error("Expected flags in root command")
			}

			// Check for known commands and flags
			cmdNames := make(map[string]bool)
			for _, sub := range subs {
				if subMap, ok := sub.(map[string]interface{}); ok {
					if use, ok := subMap["use"].(string); ok {
						cmdNames[use] = true
					}
				}
			}

			expectedCmds := []string{"scan", "version", "doctor", "config", "completion"}
			for _, expected := range expectedCmds {
				found := false
				for name := range cmdNames {
					if strings.HasPrefix(name, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected subcommand %q in command tree, found: %v", expected, cmdNames)
				}
			}

			// Verify flag structure
			flagNames := make(map[string]bool)
			for _, flag := range flags {
				if flagMap, ok := flag.(map[string]interface{}); ok {
					if name, ok := flagMap["name"].(string); ok {
						flagNames[name] = true
					}
					if flagMap["description"] == nil {
						t.Error("Expected description in flag")
					}
					if flagMap["type"] == nil {
						t.Error("Expected type in flag")
					}
				}
			}

			if !flagNames["json"] {
				t.Errorf("Expected --json flag in command tree, found: %v", flagNames)
			}
		})
	}
}

func TestCLIHelpJSONCommandTreeStructure(t *testing.T) {
	binary := testutil.BuildCLI(t)

	// Get help JSON output
	output, err := runCLI(t, binary, t.TempDir(), "--help", "--json")
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Unmarshal into structured types for deep validation
	var result struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
		Meta   map[string]interface{} `json:"meta"`
	}
	if parseErr := json.Unmarshal(output, &result); parseErr != nil {
		t.Fatalf("Failed to parse JSON: %v", parseErr)
	}

	// Validate command tree structure
	data := result.Data
	if data["cli_name"] != "deepscanbot" {
		t.Errorf("Expected cli_name 'deepscanbot', got '%v'", data["cli_name"])
	}

	root := data["root_command"].(map[string]interface{})

	// Validate root command fields
	expectedRootFields := []string{"use", "short", "subcommands", "flags"}
	for _, field := range expectedRootFields {
		if root[field] == nil {
			t.Errorf("Root command missing field: %s", field)
		}
	}

	// Verify each subcommand has required fields
	subs := root["subcommands"].([]interface{})
	for _, sub := range subs {
		subMap := sub.(map[string]interface{})
		requiredFields := []string{"use", "short"}
		for _, field := range requiredFields {
			if subMap[field] == nil {
				t.Errorf("Subcommand %v missing required field: %s", subMap["use"], field)
			}
		}

		if short, ok := subMap["short"].(string); !ok || short == "" {
			t.Errorf("Subcommand %v has empty short description", subMap["use"])
		}

		if flags, ok := subMap["flags"].([]interface{}); ok {
			for _, flag := range flags {
				flagMap := flag.(map[string]interface{})
				if flagMap["name"] == nil || flagMap["description"] == nil || flagMap["type"] == nil {
					t.Errorf("Flag in subcommand %v has missing required fields: %v", subMap["use"], flagMap)
				}
			}
		}
	}

	// Verify scan subcommand has depth and timeout in flags
	var scanCmd map[string]interface{}
	for _, sub := range subs {
		subMap := sub.(map[string]interface{})
		if use, ok := subMap["use"].(string); ok && strings.HasPrefix(use, "scan") {
			scanCmd = subMap
			break
		}
	}
	if scanCmd != nil {
		if flags, ok := scanCmd["flags"].([]interface{}); ok {
			hasDepth := false
			hasTimeout := false
			for _, flag := range flags {
				flagMap := flag.(map[string]interface{})
				if flagMap["name"] == "depth" {
					hasDepth = true
				}
				if flagMap["name"] == "timeout" {
					hasTimeout = true
				}
			}
			if !hasDepth {
				t.Error("Scan command missing 'depth' flag")
			}
			if !hasTimeout {
				t.Error("Scan command missing 'timeout' flag")
			}
		}
	}
}

func TestCLIHelpJSONHasConsistentOrder(t *testing.T) {
	binary := testutil.BuildCLI(t)

	output1, err := runCLI(t, binary, t.TempDir(), "--help", "--json")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	output2, err := runCLI(t, binary, t.TempDir(), "--help", "--json")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	var resp1, resp2 struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(output1, &resp1); err != nil {
		t.Fatalf("Failed to parse first output: %v", err)
	}
	if err := json.Unmarshal(output2, &resp2); err != nil {
		t.Fatalf("Failed to parse second output: %v", err)
	}

	data1, err := json.Marshal(resp1.Data)
	if err != nil {
		t.Fatalf("Failed to marshal first data: %v", err)
	}
	data2, err := json.Marshal(resp2.Data)
	if err != nil {
		t.Fatalf("Failed to marshal second data: %v", err)
	}

	if string(data1) != string(data2) {
		t.Error("Command tree JSON output is not consistent between calls")
	}
}

func TestCLIJSONOutputBackwardCompatibility(t *testing.T) {
	binary := testutil.BuildCLI(t)

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

	output, err := runCLI(t, binary, workdir, "scan", server.URL, "depth=0", "output=compat-test")
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputFile := filepath.Join(workdir, "compat-test.txt")
	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("Text output file was not created: %v", err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, `"status"`) && strings.Contains(outputStr, `"success"`) {
		t.Error("Output should be human-readable when --json is not specified")
	}
}
