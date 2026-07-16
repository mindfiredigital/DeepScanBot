package exitcode

import "fmt"

const (
	// Success - command completed without errors
	Success int = 0

	// InvalidInput - invalid argument or option value
	InvalidInput int = 1

	// ValidationError - data fails semantic validation
	ValidationError int = 2

	// AuthFailure - authentication failed
	AuthFailure int = 3

	// AuthzFailure - authenticated but lacks permission
	AuthzFailure int = 10

	// NotFound - requested resource cannot be located
	NotFound int = 20

	// NetworkFailure - network request failed (non-timeout)
	NetworkFailure int = 30

	// Timeout - operation exceeded configured deadline
	Timeout int = 31

	// InternalError - unexpected error (likely a bug)
	InternalError int = 70
)

// ExitCode holds a standard exit code together with a user-facing error
// message and, optionally, a resolution hint that tells the user how to fix
// the problem.
type ExitCode struct {
	Code    int
	Message string
	Hint    string // optional; may be empty
}

// Error implements the error interface so that ExitCode can be used as a
// standard Go error and easily wrapped.
func (e *ExitCode) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\nHint: %s", e.Message, e.Hint)
	}
	return e.Message
}

// Unwrap is a no-op – ExitCode does not wrap another error.  Keeping this
// method makes it safe to use with errors.Is / errors.As in the future.
func (e *ExitCode) Unwrap() error { return nil }

// String returns a human-readable summary of the exit code.
func (e *ExitCode) String() string {
	return fmt.Sprintf("exit code %d: %s", e.Code, e.Message)
}

// --- Pre-built sentinel errors for the most common scenarios ----------------

var (
	// ErrInvalidURL is returned when a URL cannot be parsed or uses an
	// unsupported scheme.
	ErrInvalidURL = &ExitCode{
		Code:    InvalidInput,
		Message: "Invalid URL: the URL could not be parsed or uses an unsupported scheme.",
		Hint:    "Use an absolute http:// or https:// URL. Example: https://example.com",
	}

	// ErrEmptyURL is returned when no URL is supplied.
	ErrEmptyURL = &ExitCode{
		Code:    InvalidInput,
		Message: "No URL provided.",
		Hint:    "Specify a URL to scan. Example: deepscanbot scan https://example.com",
	}

	// ErrEmptyOutputFilename is returned when the output= option is empty.
	ErrEmptyOutputFilename = &ExitCode{
		Code:    ValidationError,
		Message: "Output filename must not be empty.",
		Hint:    "Use output=<filename> with a non-empty value.",
	}

	// ErrFileRead is returned when a file cannot be read.
	ErrFileRead = &ExitCode{
		Code:    InternalError,
		Message: "Failed to read file.",
		Hint:    "Check that the file exists and is readable.",
	}

	// ErrResumeLoadFailed is returned when the --resume file cannot be read.
	ErrResumeLoadFailed = &ExitCode{
		Code:    InternalError,
		Message: "Failed to load existing results for resume mode.",
		Hint:    "Check that the output file exists and is readable.",
	}

	// ErrScanFailed is returned when the crawl itself fails.
	ErrScanFailed = &ExitCode{
		Code:    InternalError,
		Message: "Scan failed unexpectedly.",
		Hint:    "Retry with reduced concurrency or check the target URL.",
	}

	// ErrTimeout is returned when the crawl exceeds the configured timeout.
	ErrTimeout = &ExitCode{
		Code:    Timeout,
		Message: "Crawl timed out.",
		Hint:    "Increase the timeout with --timeout=<seconds> or reduce the crawl scope.",
	}

	// ErrWriteOutput is returned when writing the output file fails.
	ErrWriteOutput = &ExitCode{
		Code:    InternalError,
		Message: "Failed to write output file.",
		Hint:    "Ensure the output path is writable and there is free disk space.",
	}

	// ErrJSONOutput is returned when serializing the JSON response fails.
	ErrJSONOutput = &ExitCode{
		Code:    InternalError,
		Message: "Failed to write JSON output.",
		Hint:    "This is an internal error. Please file a bug report.",
	}

	// ErrBuildFailed is an internal error placeholder (used for build-time
	// failures in tests, should never appear at runtime).
	ErrBuildFailed = &ExitCode{
		Code:    InternalError,
		Message: "Failed to build CLI binary.",
	}
)
