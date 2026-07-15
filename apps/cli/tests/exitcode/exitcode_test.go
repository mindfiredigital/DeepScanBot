package exitcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindfiredigital/DeepScanBot/apps/cli/tests/testutil"
	"github.com/mindfiredigital/DeepScanBot/packages/exitcode"
)

// Helper function leveraging testutil to fetch just the exit code
func exitCodeFor(t *testing.T, binary, workdir string, args ...string) int {
	t.Helper()
	_, _, code := testutil.CombinedOutputFor(t, binary, workdir, args...)
	return code
}

func TestCLIExitCodeSuccess(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "version")
	if code != 0 {
		t.Errorf("version command exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeInvalidURL(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	tests := []struct {
		name      string
		args      []string
		wantCode  int
		wantErrIn string
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
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	tests := []struct {
		name       string
		args       []string
		wantHintIn string
		wantErrIn  string
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
			_, stderr, _ := testutil.CombinedOutputFor(t, binary, workdir, tt.args...)
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
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer server.Close()

	code := exitCodeFor(t, binary, workdir, "scan", server.URL, "depth=0")
	if code != 0 {
		t.Errorf("scan of valid server exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeNetworkFailure(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	_, stderr, code := testutil.CombinedOutputFor(t, binary, workdir, "scan", "http://127.0.0.1:0", "--timeout=1s")
	t.Logf("network failure: exit code=%d, stderr=%s", code, stderr)

	// Unconditionally require failure. An exit code of 0 means the CLI failed to catch a broken connection.
	if code == 0 {
		t.Fatalf("Expected non-zero exit code for network failure, got 0")
	}

	// Validate that stderr contains a clear, actionable error message
	lowerStderr := strings.ToLower(stderr)
	if !strings.Contains(lowerStderr, "error") && !strings.Contains(lowerStderr, "failed") && !strings.Contains(lowerStderr, "timeout") && !strings.Contains(lowerStderr, "refused") {
		t.Errorf("stderr should contain an actionable error message, got: %s", stderr)
	}
}

func TestCLIExitCodeVersion(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "version")
	if code != 0 {
		t.Errorf("version exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeVersionJSON(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "version", "--json")
	if code != 0 {
		t.Errorf("version --json exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeDoctor(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "doctor")
	if code != 0 {
		t.Errorf("doctor exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeHelp(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "--help")
	if code != 0 {
		t.Errorf("help exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeHelpJSON(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "--help", "--json")
	if code != 0 {
		t.Errorf("help --json exit code = %d, want 0", code)
	}
}

func TestCLIExitCodeScanWithJSONOutput(t *testing.T) {
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

	stdout, stderr, code := testutil.CombinedOutputFor(t, binary, workdir, "scan", server.URL, "depth=0", "--json", "output=exit-json-test")

	if code != 0 {
		t.Errorf("scan --json exit code = %d, want 0; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}

	if result["status"] != "success" {
		t.Errorf("status = %v, want 'success'", result["status"])
	}

	outputFile := filepath.Join(workdir, "exit-json-test.json")
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestCLIExitCodeScanTextOutput(t *testing.T) {
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

	code := exitCodeFor(t, binary, workdir, "scan", server.URL, "depth=0", "output=exit-text-test")
	if code != 0 {
		t.Errorf("scan text output exit code = %d, want 0", code)
	}

	outputFile := filepath.Join(workdir, "exit-text-test.txt")
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("text output file not created: %v", err)
	}
}

func TestCLIExitCodeErrorForEmptyOutputFilename(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	_, stderr, code := testutil.CombinedOutputFor(t, binary, workdir, "scan", "http://example.com", "output=")

	t.Logf("stderr: %s", stderr)
	if code == 0 {
		t.Error("scan with empty output= should fail with non-zero exit code")
	}
}

func TestCLIExitCodeInvalidFlag(t *testing.T) {
	binary := testutil.BuildCLI(t)
	workdir := t.TempDir()

	code := exitCodeFor(t, binary, workdir, "scan", "http://example.com", "--nonexistent-flag")
	if code == 0 {
		t.Error("scan with unknown flag should fail with non-zero exit code")
	}
}

func TestSuccessCode(t *testing.T) {
	if exitcode.Success != 0 {
		t.Errorf("Success = %d, want 0", exitcode.Success)
	}
}

func TestInvalidInputCode(t *testing.T) {
	if exitcode.InvalidInput != 1 {
		t.Errorf("InvalidInput = %d, want 1", exitcode.InvalidInput)
	}
}

func TestValidationErrorCode(t *testing.T) {
	if exitcode.ValidationError != 2 {
		t.Errorf("ValidationError = %d, want 2", exitcode.ValidationError)
	}
}

func TestAuthFailureCode(t *testing.T) {
	if exitcode.AuthFailure != 3 {
		t.Errorf("AuthFailure = %d, want 3", exitcode.AuthFailure)
	}
}

func TestAuthzFailureCode(t *testing.T) {
	if exitcode.AuthzFailure != 10 {
		t.Errorf("AuthzFailure = %d, want 10", exitcode.AuthzFailure)
	}
}

func TestNotFoundCode(t *testing.T) {
	if exitcode.NotFound != 20 {
		t.Errorf("NotFound = %d, want 20", exitcode.NotFound)
	}
}

func TestNetworkFailureCode(t *testing.T) {
	if exitcode.NetworkFailure != 30 {
		t.Errorf("NetworkFailure = %d, want 30", exitcode.NetworkFailure)
	}
}

func TestTimeoutCode(t *testing.T) {
	if exitcode.Timeout != 31 {
		t.Errorf("Timeout = %d, want 31", exitcode.Timeout)
	}
}

func TestInternalErrorCode(t *testing.T) {
	if exitcode.InternalError != 70 {
		t.Errorf("InternalError = %d, want 70", exitcode.InternalError)
	}
}

func TestExitCodeError(t *testing.T) {
	ec := &exitcode.ExitCode{
		Code:    exitcode.InvalidInput,
		Message: "Something went wrong",
		Hint:    "Try again",
	}

	errStr := ec.Error()
	if errStr != "Something went wrong\nHint: Try again" {
		t.Errorf("Error() = %q, want %q", errStr, "Something went wrong\nHint: Try again")
	}
}

func TestExitCodeErrorNoHint(t *testing.T) {
	ec := &exitcode.ExitCode{
		Code:    exitcode.InvalidInput,
		Message: "Something went wrong",
	}

	errStr := ec.Error()
	if errStr != "Something went wrong" {
		t.Errorf("Error() = %q, want %q", errStr, "Something went wrong")
	}
}

func TestExitCodeString(t *testing.T) {
	ec := &exitcode.ExitCode{
		Code:    exitcode.InvalidInput,
		Message: "Invalid input",
	}

	str := ec.String()
	if str != "exit code 1: Invalid input" {
		t.Errorf("String() = %q, want %q", str, "exit code 1: Invalid input")
	}
}

func TestExitCodeUnwrap(t *testing.T) {
	ec := &exitcode.ExitCode{Code: exitcode.InternalError, Message: "test"}
	if unwrapped := ec.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

func TestErrInvalidURL(t *testing.T) {
	if exitcode.ErrInvalidURL.Code != exitcode.InvalidInput {
		t.Errorf("ErrInvalidURL.Code = %d, want %d", exitcode.ErrInvalidURL.Code, exitcode.InvalidInput)
	}
	if exitcode.ErrInvalidURL.Message == "" {
		t.Error("ErrInvalidURL.Message should not be empty")
	}
	if exitcode.ErrInvalidURL.Hint == "" {
		t.Error("ErrInvalidURL.Hint should not be empty")
	}
}

func TestErrEmptyURL(t *testing.T) {
	if exitcode.ErrEmptyURL.Code != exitcode.InvalidInput {
		t.Errorf("ErrEmptyURL.Code = %d, want %d", exitcode.ErrEmptyURL.Code, exitcode.InvalidInput)
	}
	if exitcode.ErrEmptyURL.Message == "" {
		t.Error("ErrEmptyURL.Message should not be empty")
	}
	if exitcode.ErrEmptyURL.Hint == "" {
		t.Error("ErrEmptyURL.Hint should not be empty")
	}
}

func TestErrEmptyOutputFilename(t *testing.T) {
	if exitcode.ErrEmptyOutputFilename.Code != exitcode.ValidationError {
		t.Errorf("ErrEmptyOutputFilename.Code = %d, want %d", exitcode.ErrEmptyOutputFilename.Code, exitcode.ValidationError)
	}
	if exitcode.ErrEmptyOutputFilename.Message == "" {
		t.Error("ErrEmptyOutputFilename.Message should not be empty")
	}
	if exitcode.ErrEmptyOutputFilename.Hint == "" {
		t.Error("ErrEmptyOutputFilename.Hint should not be empty")
	}
}

func TestErrResumeLoadFailed(t *testing.T) {
	if exitcode.ErrResumeLoadFailed.Code != exitcode.InternalError {
		t.Errorf("ErrResumeLoadFailed.Code = %d, want %d", exitcode.ErrResumeLoadFailed.Code, exitcode.InternalError)
	}
}

func TestErrScanFailed(t *testing.T) {
	if exitcode.ErrScanFailed.Code != exitcode.InternalError {
		t.Errorf("ErrScanFailed.Code = %d, want %d", exitcode.ErrScanFailed.Code, exitcode.InternalError)
	}
}

func TestErrWriteOutput(t *testing.T) {
	if exitcode.ErrWriteOutput.Code != exitcode.InternalError {
		t.Errorf("ErrWriteOutput.Code = %d, want %d", exitcode.ErrWriteOutput.Code, exitcode.InternalError)
	}
}

func TestErrJSONOutput(t *testing.T) {
	if exitcode.ErrJSONOutput.Code != exitcode.InternalError {
		t.Errorf("ErrJSONOutput.Code = %d, want %d", exitcode.ErrJSONOutput.Code, exitcode.InternalError)
	}
}

func TestErrBuildFailed(t *testing.T) {
	if exitcode.ErrBuildFailed.Code != exitcode.InternalError {
		t.Errorf("ErrBuildFailed.Code = %d, want %d", exitcode.ErrBuildFailed.Code, exitcode.InternalError)
	}
}
