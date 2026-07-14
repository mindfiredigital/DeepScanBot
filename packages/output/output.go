package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ResponseStatus represents the status of a command response
type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "success"
	StatusError   ResponseStatus = "error"
)

// Response is a standardized JSON response format for all CLI commands
type Response struct {
	Status  ResponseStatus    `json:"status"`
	Data    interface{}       `json:"data,omitempty"`
	Error   *ErrorDetail      `json:"error,omitempty"`
	Meta    *ResponseMetadata `json:"meta,omitempty"`
}

// ErrorDetail contains error information for error responses
type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ResponseMetadata contains additional metadata about the response
type ResponseMetadata struct {
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command,omitempty"`
	Duration  int64     `json:"duration_ms,omitempty"`
}

// Formatter handles output formatting for different formats
type Formatter struct {
	jsonMode bool
}

// NewFormatter creates a new output formatter
func NewFormatter(jsonMode bool) *Formatter {
	return &Formatter{
		jsonMode: jsonMode,
	}
}

// WriteSuccess writes a successful response
func (f *Formatter) WriteSuccess(w io.Writer, data interface{}, meta *ResponseMetadata) error {
	if f.jsonMode {
		resp := Response{
			Status: StatusSuccess,
			Data:   data,
			Meta:   meta,
		}
		return writeJSON(w, resp)
	}

	// Human-readable format
	return writeHumanReadable(w, data, meta)
}

// WriteError writes an error response
func (f *Formatter) WriteError(w io.Writer, message string, code string, meta *ResponseMetadata) error {
	if f.jsonMode {
		resp := Response{
			Status: StatusError,
			Error: &ErrorDetail{
				Message: message,
				Code:    code,
			},
			Meta: meta,
		}
		return writeJSON(w, resp)
	}

	// Human-readable format
	_, err := fmt.Fprintf(w, "Error: %s\n", message)
	return err
}

// IsJSONMode returns true if JSON mode is enabled
func (f *Formatter) IsJSONMode() bool {
	return f.jsonMode
}

// writeJSON writes a JSON response to stdout
func writeJSON(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// writeHumanReadable writes human-readable output
func writeHumanReadable(w io.Writer, data interface{}, meta *ResponseMetadata) error {
	// Default human-readable formatting
	// Specific commands can override this with custom formatting
	switch d := data.(type) {
	case string:
		_, err := fmt.Fprint(w, d)
		return err
	default:
		_, err := fmt.Fprintf(w, "%v\n", data)
		return err
	}
}

// WriteDiagnostic writes diagnostic messages to stderr
func WriteDiagnostic(message string) {
	fmt.Fprintln(os.Stderr, message)
}

// WriteDiagnosticf writes a formatted diagnostic message to stderr
func WriteDiagnosticf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// NewResponseMetadata creates a new response metadata
func NewResponseMetadata(command string, duration time.Duration) *ResponseMetadata {
	return &ResponseMetadata{
		Timestamp: time.Now(),
		Command:   command,
		Duration:  duration.Milliseconds(),
	}
}