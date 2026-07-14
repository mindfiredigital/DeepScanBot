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
	r, w, _ := os.Pipe()
	os.Stdout = w

	if IsTTY() {
		t.Error("IsTTY() should return false when stdout is piped/redirected")
	}

	w.Close()
	os.Stdout = oldStdout
	r.Close()
}

func TestIsTerminalWithRedirectedOutput(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	if IsTerminal() {
		t.Error("IsTerminal() should return false when stderr is piped/redirected")
	}

	w.Close()
	os.Stderr = oldStderr
	r.Close()
}

func TestIsInputTTYWithRedirectedInput(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	if IsInputTTY() {
		t.Error("IsInputTTY() should return false when stdin is piped/redirected")
	}

	w.Close()
	os.Stdin = oldStdin
	r.Close()
}
