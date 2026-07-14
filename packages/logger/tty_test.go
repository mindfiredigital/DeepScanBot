package logger

import (
	"os"
	"testing"
)

func TestIsTTY(t *testing.T) {
	// When running tests via `go test`, stdout is typically not a TTY
	// (unless run with a TTY attached). This test verifies the function
	// works without panicking and returns a boolean.
	result := IsTTY()
	// We just verify it returns something (true/false) without error
	_ = result
}

func TestIsTerminal(t *testing.T) {
	// When running tests via `go test`, stderr is typically not a TTY
	// (unless run with a TTY attached). This test verifies the function
	// works without panicking and returns a boolean.
	result := IsTerminal()
	// We just verify it returns something (true/false) without error
	_ = result
}

func TestIsInputTTY(t *testing.T) {
	// When running tests via `go test`, stdin is typically not a TTY
	// (unless run with a TTY attached). This test verifies the function
	// works without panicking and returns a boolean.
	result := IsInputTTY()
	// We just verify it returns something (true/false) without error
	_ = result
}

func TestIsTTYWithRedirectedOutput(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe to simulate redirected output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// IsTTY should return false when stdout is a pipe (redirected)
	if IsTTY() {
		t.Error("IsTTY() should return false when stdout is piped/redirected")
	}

	// Clean up
	w.Close()
	os.Stdout = oldStdout
	r.Close()
}

func TestIsTerminalWithRedirectedOutput(t *testing.T) {
	// Save original stderr
	oldStderr := os.Stderr

	// Create a pipe to simulate redirected output
	r, w, _ := os.Pipe()
	os.Stderr = w

	// IsTerminal should return false when stderr is a pipe (redirected)
	if IsTerminal() {
		t.Error("IsTerminal() should return false when stderr is piped/redirected")
	}

	// Clean up
	w.Close()
	os.Stderr = oldStderr
	r.Close()
}

func TestIsInputTTYWithRedirectedInput(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin

	// Create a pipe to simulate redirected input
	r, w, _ := os.Pipe()
	os.Stdin = r

	// IsInputTTY should return false when stdin is a pipe (redirected)
	if IsInputTTY() {
		t.Error("IsInputTTY() should return false when stdin is piped/redirected")
	}

	// Clean up
	w.Close()
	os.Stdin = oldStdin
	r.Close()
}