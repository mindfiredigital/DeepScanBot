package logger_test

import (
	"os"
	"testing"

	"github.com/mindfiredigital/DeepScanBot/packages/logger"
)

func TestIsTTY(t *testing.T) {
	_ = logger.IsTTY()
}

func TestIsTerminal(t *testing.T) {
	_ = logger.IsTerminal()
}

func TestIsInputTTY(t *testing.T) {
	_ = logger.IsInputTTY()
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

	if logger.IsTTY() {
		t.Error("logger.IsTTY() should return false when stdout is piped/redirected")
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

	if logger.IsTerminal() {
		t.Error("logger.IsTerminal() should return false when stderr is piped/redirected")
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

	if logger.IsInputTTY() {
		t.Error("logger.IsInputTTY() should return false when stdin is piped/redirected")
	}
}
