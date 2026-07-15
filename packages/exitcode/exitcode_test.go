package exitcode

import "testing"

func TestSuccessCode(t *testing.T) {
	if Success != 0 {
		t.Errorf("Success = %d, want 0", Success)
	}
}

func TestInvalidInputCode(t *testing.T) {
	if InvalidInput != 1 {
		t.Errorf("InvalidInput = %d, want 1", InvalidInput)
	}
}

func TestValidationErrorCode(t *testing.T) {
	if ValidationError != 2 {
		t.Errorf("ValidationError = %d, want 2", ValidationError)
	}
}

func TestAuthFailureCode(t *testing.T) {
	if AuthFailure != 3 {
		t.Errorf("AuthFailure = %d, want 3", AuthFailure)
	}
}

func TestAuthzFailureCode(t *testing.T) {
	if AuthzFailure != 10 {
		t.Errorf("AuthzFailure = %d, want 10", AuthzFailure)
	}
}

func TestNotFoundCode(t *testing.T) {
	if NotFound != 20 {
		t.Errorf("NotFound = %d, want 20", NotFound)
	}
}

func TestNetworkFailureCode(t *testing.T) {
	if NetworkFailure != 30 {
		t.Errorf("NetworkFailure = %d, want 30", NetworkFailure)
	}
}

func TestTimeoutCode(t *testing.T) {
	if Timeout != 31 {
		t.Errorf("Timeout = %d, want 31", Timeout)
	}
}

func TestInternalErrorCode(t *testing.T) {
	if InternalError != 70 {
		t.Errorf("InternalError = %d, want 70", InternalError)
	}
}

func TestExitCodeError(t *testing.T) {
	ec := &ExitCode{
		Code:    InvalidInput,
		Message: "Something went wrong",
		Hint:    "Try again",
	}

	errStr := ec.Error()
	if errStr != "Something went wrong\nHint: Try again" {
		t.Errorf("Error() = %q, want %q", errStr, "Something went wrong\nHint: Try again")
	}
}

func TestExitCodeErrorNoHint(t *testing.T) {
	ec := &ExitCode{
		Code:    InvalidInput,
		Message: "Something went wrong",
	}

	errStr := ec.Error()
	if errStr != "Something went wrong" {
		t.Errorf("Error() = %q, want %q", errStr, "Something went wrong")
	}
}

func TestExitCodeString(t *testing.T) {
	ec := &ExitCode{
		Code:    InvalidInput,
		Message: "Invalid input",
	}

	str := ec.String()
	if str != "exit code 1: Invalid input" {
		t.Errorf("String() = %q, want %q", str, "exit code 1: Invalid input")
	}
}

func TestExitCodeUnwrap(t *testing.T) {
	ec := &ExitCode{Code: InternalError, Message: "test"}
	if unwrapped := ec.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

func TestErrInvalidURL(t *testing.T) {
	if ErrInvalidURL.Code != InvalidInput {
		t.Errorf("ErrInvalidURL.Code = %d, want %d", ErrInvalidURL.Code, InvalidInput)
	}
	if ErrInvalidURL.Message == "" {
		t.Error("ErrInvalidURL.Message should not be empty")
	}
	if ErrInvalidURL.Hint == "" {
		t.Error("ErrInvalidURL.Hint should not be empty")
	}
}

func TestErrEmptyURL(t *testing.T) {
	if ErrEmptyURL.Code != InvalidInput {
		t.Errorf("ErrEmptyURL.Code = %d, want %d", ErrEmptyURL.Code, InvalidInput)
	}
	if ErrEmptyURL.Message == "" {
		t.Error("ErrEmptyURL.Message should not be empty")
	}
	if ErrEmptyURL.Hint == "" {
		t.Error("ErrEmptyURL.Hint should not be empty")
	}
}

func TestErrEmptyOutputFilename(t *testing.T) {
	if ErrEmptyOutputFilename.Code != ValidationError {
		t.Errorf("ErrEmptyOutputFilename.Code = %d, want %d", ErrEmptyOutputFilename.Code, ValidationError)
	}
	if ErrEmptyOutputFilename.Message == "" {
		t.Error("ErrEmptyOutputFilename.Message should not be empty")
	}
	if ErrEmptyOutputFilename.Hint == "" {
		t.Error("ErrEmptyOutputFilename.Hint should not be empty")
	}
}

func TestErrResumeLoadFailed(t *testing.T) {
	if ErrResumeLoadFailed.Code != InternalError {
		t.Errorf("ErrResumeLoadFailed.Code = %d, want %d", ErrResumeLoadFailed.Code, InternalError)
	}
}

func TestErrScanFailed(t *testing.T) {
	if ErrScanFailed.Code != InternalError {
		t.Errorf("ErrScanFailed.Code = %d, want %d", ErrScanFailed.Code, InternalError)
	}
}

func TestErrWriteOutput(t *testing.T) {
	if ErrWriteOutput.Code != InternalError {
		t.Errorf("ErrWriteOutput.Code = %d, want %d", ErrWriteOutput.Code, InternalError)
	}
}

func TestErrJSONOutput(t *testing.T) {
	if ErrJSONOutput.Code != InternalError {
		t.Errorf("ErrJSONOutput.Code = %d, want %d", ErrJSONOutput.Code, InternalError)
	}
}

func TestErrBuildFailed(t *testing.T) {
	if ErrBuildFailed.Code != InternalError {
		t.Errorf("ErrBuildFailed.Code = %d, want %d", ErrBuildFailed.Code, InternalError)
	}
}