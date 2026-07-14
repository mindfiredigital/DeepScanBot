package tests

import (
	"testing"

	"github.com/mindfiredigital/DeepScanBot/packages/exitcode"
)

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