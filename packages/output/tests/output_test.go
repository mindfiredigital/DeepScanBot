package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	output "github.com/mindfiredigital/DeepScanBot/packages/output"
)

func TestNewFormatter(t *testing.T) {
	formatter := output.NewFormatter(true)
	if formatter == nil {
		t.Fatal("NewFormatter returned nil")
	}
	if !formatter.IsJSONMode() {
		t.Error("Expected JSON mode to be enabled")
	}

	formatter2 := output.NewFormatter(false)
	if formatter2.IsJSONMode() {
		t.Error("Expected JSON mode to be disabled")
	}
}

func TestWriteSuccessJSON(t *testing.T) {
	var buf bytes.Buffer
	formatter := output.NewFormatter(true)

	data := map[string]string{
		"version": "1.0.0",
		"name":    "DeepScanBot CLI",
	}

	meta := &output.ResponseMetadata{
		Timestamp: time.Now(),
		Command:   "version",
		Duration:  0,
	}

	err := formatter.WriteSuccess(&buf, data, meta)
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	// Parse the JSON output
	var resp output.Response
	decoder := json.NewDecoder(&buf)
	if err := decoder.Decode(&resp); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify structure
	if resp.Status != output.StatusSuccess {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}

	if resp.Data == nil {
		t.Error("Expected data to be present")
	}

	if resp.Meta == nil {
		t.Error("Expected metadata to be present")
	}

	if resp.Meta.Command != "version" {
		t.Errorf("Expected command 'version', got '%s'", resp.Meta.Command)
	}

	if resp.Error != nil {
		t.Error("Expected error to be nil for success response")
	}
}

func TestWriteSuccessHumanReadable(t *testing.T) {
	var buf bytes.Buffer
	formatter := output.NewFormatter(false)

	data := "Hello, World!"
	meta := &output.ResponseMetadata{
		Timestamp: time.Now(),
		Command:   "test",
		Duration:  100,
	}

	err := formatter.WriteSuccess(&buf, data, meta)
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	outputStr := buf.String()
	if !strings.Contains(outputStr, "Hello, World!") {
		t.Errorf("Expected human-readable output to contain 'Hello, World!', got: %s", outputStr)
	}
}

func TestWriteErrorJSON(t *testing.T) {
	var buf bytes.Buffer
	formatter := output.NewFormatter(true)

	meta := &output.ResponseMetadata{
		Timestamp: time.Now(),
		Command:   "scan",
		Duration:  0,
	}

	err := formatter.WriteError(&buf, "Invalid URL", "invalid_url", meta)
	if err != nil {
		t.Fatalf("WriteError failed: %v", err)
	}

	// Parse the JSON output
	var resp output.Response
	decoder := json.NewDecoder(&buf)
	if err := decoder.Decode(&resp); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify structure
	if resp.Status != output.StatusError {
		t.Errorf("Expected status 'error', got '%s'", resp.Status)
	}

	if resp.Error == nil {
		t.Fatal("Expected error detail to be present")
	}

	if resp.Error.Message != "Invalid URL" {
		t.Errorf("Expected error message 'Invalid URL', got '%s'", resp.Error.Message)
	}

	if resp.Error.Code != "invalid_url" {
		t.Errorf("Expected error code 'invalid_url', got '%s'", resp.Error.Code)
	}

	if resp.Data != nil {
		t.Error("Expected data to be nil for error response")
	}
}

func TestWriteErrorHumanReadable(t *testing.T) {
	var buf bytes.Buffer
	formatter := output.NewFormatter(false)

	meta := &output.ResponseMetadata{
		Timestamp: time.Now(),
		Command:   "scan",
		Duration:  0,
	}

	err := formatter.WriteError(&buf, "Invalid URL", "invalid_url", meta)
	if err != nil {
		t.Fatalf("WriteError failed: %v", err)
	}

	outputStr := buf.String()
	if !strings.Contains(outputStr, "Error: Invalid URL") {
		t.Errorf("Expected human-readable error output, got: %s", outputStr)
	}
}

func TestWriteDiagnostic(t *testing.T) {
	// This test verifies that WriteDiagnostic doesn't panic
	// We can't easily capture stderr in a test, but we can verify it doesn't crash
	output.WriteDiagnostic("Test diagnostic message")
	output.WriteDiagnosticf("Test diagnostic message with %s", "format")
}

func TestNewResponseMetadata(t *testing.T) {
	start := time.Now()
	duration := 1500 * time.Millisecond

	meta := output.NewResponseMetadata("scan", duration)

	if meta == nil {
		t.Fatal("NewResponseMetadata returned nil")
	}

	if meta.Command != "scan" {
		t.Errorf("Expected command 'scan', got '%s'", meta.Command)
	}

	if meta.Duration != 1500 {
		t.Errorf("Expected duration 1500ms, got %d", meta.Duration)
	}

	if meta.Timestamp.Before(start) {
		t.Error("Expected timestamp to be after start time")
	}
}

func TestResponseJSONSerialization(t *testing.T) {
	// Test that Response can be properly serialized to JSON
	resp := output.Response{
		Status: output.StatusSuccess,
		Data: map[string]interface{}{
			"urls":  []string{"http://example.com"},
			"count": 1,
		},
		Error: nil,
		Meta: &output.ResponseMetadata{
			Timestamp: time.Now(),
			Command:   "scan",
			Duration:  5000,
		},
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["status"] != "success" {
		t.Errorf("Expected status 'success', got '%v'", parsed["status"])
	}
}

func TestErrorDetailJSONSerialization(t *testing.T) {
	errorDetail := output.ErrorDetail{
		Message: "Something went wrong",
		Code:    "internal_error",
	}

	jsonData, err := json.Marshal(errorDetail)
	if err != nil {
		t.Fatalf("Failed to marshal error detail: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["message"] != "Something went wrong" {
		t.Errorf("Expected message 'Something went wrong', got '%s'", parsed["message"])
	}

	if parsed["code"] != "internal_error" {
		t.Errorf("Expected code 'internal_error', got '%s'", parsed["code"])
	}
}

func TestResponseMetadataJSONSerialization(t *testing.T) {
	meta := output.ResponseMetadata{
		Timestamp: time.Now(),
		Command:   "test",
		Duration:  1000,
	}

	jsonData, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["command"] != "test" {
		t.Errorf("Expected command 'test', got '%v'", parsed["command"])
	}

	if parsed["duration_ms"] != float64(1000) {
		t.Errorf("Expected duration 1000, got '%v'", parsed["duration_ms"])
	}
}