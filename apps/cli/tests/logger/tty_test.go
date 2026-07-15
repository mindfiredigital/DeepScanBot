package tests

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
	r, w, _ := os.Pipe()
	os.Stdout = w

	if logger.IsTTY() {
		t.Error("logger.IsTTY() should return false when stdout is piped/redirected")
	}

	w.Close()
	os.Stdout = oldStdout
	r.Close()
}

func TestIsTerminalWithRedirectedOutput(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	if logger.IsTerminal() {
		t.Error("logger.IsTerminal() should return false when stderr is piped/redirected")
	}

	w.Close()
	os.Stderr = oldStderr
	r.Close()
}

func TestIsInputTTYWithRedirectedInput(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	if logger.IsInputTTY() {
		t.Error("logger.IsInputTTY() should return false when stdin is piped/redirected")
	}

	w.Close()
	os.Stdin = oldStdin
	r.Close()
}
