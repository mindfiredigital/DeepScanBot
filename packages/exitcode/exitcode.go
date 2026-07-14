// Package exitcode defines standardized exit codes and provides a standard
// error type that carries an exit code, a user-facing message, and an optional
// resolution hint. Every command in the CLI should use these codes so that
// scripts, CI/CD pipelines, and AI agents can rely on consistent exit behaviour.
package exitcode

import "fmt"

// Standard exit codes for the DeepScanBot CLI.
//
// Codes use the POSIX convention where 0 means success and values in the
// range 1–78 are reserved for general errors. Codes higher than 78 avoid
// conflicts with /usr/include/sysexits.h on Unix systems.
const (
	// Success indicates the command completed without errors.
	Success int = 0

	// InvalidInput is returned when the user supplies an invalid argument
	// or option value (e.g. a malformed URL, unknown flag).
	InvalidInput int = 1

	// ValidationError is returned when the provided data fails semantic
	// validation (e.g. depth < 0, empty output filename).
	ValidationError int = 2

	// AuthFailure is returned when authentication with a remote service
	// fails (e.g. invalid API token, missing credentials).
	AuthFailure int = 3

	// AuthzFailure is returned when the user is authenticated but lacks
	// permission to access the requested resource.
	AuthzFailure int = 10

	// NotFound is returned when a requested resource (URL, file, etc.)
	// cannot be located.
	NotFound int = 20

	// NetworkFailure is returned when a network request fails for reasons
	// other than timeouts (e.g. DNS resolution failure, connection refused).
	NetworkFailure int = 30

	// Timeout is returned when an operation exceeds its configured deadline.
	Timeout int = 31

	// InternalError is returned for unexpected errors that do not fit into
	// any other category. These usually indicate a bug.
	InternalError int = 70
)

// ExitCode holds a standard exit code together with a user-facing error
// message and, optionally, a resolution hint that tells the user how to fix
// the problem.
type ExitCode struct {
	Code     int
	Message  string
	Hint     string // optional; may be empty
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

	// ErrWriteOutput is returned when writing the output file fails.
	ErrWriteOutput = &ExitCode{
		Code:    InternalError,
		Message: "Failed to write output file.",
		Hint:    "Ensure the output path is writable and there is free disk space.",
	}

	// ErrJSONOutput is returned when serialising the JSON response fails.
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