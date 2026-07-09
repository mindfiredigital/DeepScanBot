package exitcode

import (
	"fmt"
	"os"
)

// HandleError prints the error to stderr and exits the process with the
// appropriate exit code.  If err is not an *ExitCode it falls back to
// InternalError.
//
// Usage:
//
//	exitcode.HandleError(err)
//
// This function never returns (it calls os.Exit).
func HandleError(err error) {
	if err == nil {
		os.Exit(Success)
	}

	var ec *ExitCode
	if asExitCode(err, &ec) {
		printError(ec)
		os.Exit(ec.Code)
	}

	// Unknown error type – treat as internal error.
	unknown := &ExitCode{
		Code:    InternalError,
		Message: "An unexpected error occurred.",
		Hint:    fmt.Sprintf("Error details: %s", err.Error()),
	}
	printError(unknown)
	os.Exit(InternalError)
}

// HandleErrorWithMessage prints a custom message together with the error's
// exit code and then exits.  If err is not an *ExitCode it falls back to
// InternalError.
//
// Usage:
//
//	exitcode.HandleErrorWithMessage("scan failed", err)
//
// This function never returns (it calls os.Exit).
func HandleErrorWithMessage(msg string, err error) {
	if err == nil {
		os.Exit(Success)
	}

	var ec *ExitCode
	if asExitCode(err, &ec) {
		ec.Message = fmt.Sprintf("%s: %s", msg, ec.Message)
		printError(ec)
		os.Exit(ec.Code)
	}

	wrapped := &ExitCode{
		Code:    InternalError,
		Message: fmt.Sprintf("%s: %s", msg, err.Error()),
		Hint:    "This is an internal error. Please file a bug report.",
	}
	printError(wrapped)
	os.Exit(InternalError)
}

// ExitSuccess exits the process with code 0.
func ExitSuccess() {
	os.Exit(Success)
}

// asExitCode is a type-assertion helper that avoids importing reflect.
func asExitCode(err error, target **ExitCode) bool {
	ec, ok := err.(*ExitCode)
	if !ok {
		return false
	}
	*target = ec
	return true
}

// printError writes the error message and hint to stderr.
func printError(ec *ExitCode) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", ec.Message)
	if ec.Hint != "" {
		_, _ = fmt.Fprintf(os.Stderr, "Hint: %s\n", ec.Hint)
	}
}