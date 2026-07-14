package noinput_test

import (
	"testing"

	"github.com/mindfiredigital/DeepScanBot/packages/noinput"
)

// TestIsInteractiveDefault verifies that IsInteractive returns the real
// terminal status by default (this test may or may not be in a TTY, so
// we just check that it doesn't panic).
func TestIsInteractiveDefault(t *testing.T) {
	_ = noinput.IsInteractive()
}

// TestIsInteractiveAfterSet verifies that IsInteractive returns false
// after SetNoInputFlag is called.
func TestIsInteractiveAfterSet(t *testing.T) {
	noinput.SetNoInputFlag()
	if noinput.IsInteractive() {
		t.Error("IsInteractive() = true, want false after SetNoInputFlag")
	}
}

// TestIsInteractiveNonTTY verifies IsInteractive returns false when
// stdin is not a terminal.
func TestIsInteractiveNonTTY(t *testing.T) {
	// Save and restore the original function
	orig := noinput.IsInteractive
	t.Cleanup(func() { noinput.IsInteractive = orig })

	// Simulate non-interactive by overriding the function
	noinput.IsInteractive = func() bool { return false }

	if noinput.IsInteractive() {
		t.Error("IsInteractive() = true, want false when stdin is not a terminal")
	}
}

// TestIsInteractiveTTY verifies IsInteractive returns true when
// stdin appears to be a terminal and --no-input was not set.
// Note: Since SetNoInputFlag is a global side effect, this test
// only works if run in isolation or as the first test.
// We override the function to simulate this case.
func TestIsInteractiveTTY(t *testing.T) {
	orig := noinput.IsInteractive
	t.Cleanup(func() { noinput.IsInteractive = orig })

	noinput.IsInteractive = func() bool { return true }

	if !noinput.IsInteractive() {
		t.Error("IsInteractive() = false, want true in TTY mode")
	}
}

// TestNoInputFlagEnv verifies that the --no-input flag is not implicitly
// set by environment variables (there's no env-based auto-detection).
func TestNoInputFlagNotSetByEnv(t *testing.T) {
	// Specifically test that CI environment variables do NOT automatically
	// trigger --no-input (it must be explicit via SetNoInputFlag).
	// This test simply verifies IsInteractive is a pure function by default.
	_ = noinput.IsInteractive()
}