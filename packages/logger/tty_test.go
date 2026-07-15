package logger

import (
	"os"
	"testing"
)

func TestIsTTY(t *testing.T) {
	_ = IsTTY()
}

func TestIsTerminal(t *testing.T) {
	_ = IsTerminal()
}

func TestIsInputTTY(t *testing.T) {
	_ = IsInputTTY()
}

func TestIsTTYWithRedirectedOutput(t *testing.T) {
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }() // Safely restore stdout if the test fails

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stdout = w

	if IsTTY() {
		t.Error("IsTTY() should return false when stdout is piped/redirected")
	}
}

func TestIsTerminalWithRedirectedOutput(t *testing.T) {
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }() // Safely restore stderr if the test fails

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stderr = w

	if IsTerminal() {
		t.Error("IsTerminal() should return false when stderr is piped/redirected")
	}
}

func TestIsInputTTYWithRedirectedInput(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Safely restore stdin if the test fails

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stdin = r

	if IsInputTTY() {
		t.Error("IsInputTTY() should return false when stdin is piped/redirected")
	}
}
