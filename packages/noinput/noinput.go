// Package noinput provides utilities for detecting and enforcing
// non-interactive CLI execution.  When --no-input is set or stdin is
// not a terminal, the CLI must never wait for user input and must
// fail immediately with a clear error message if required input is
// missing.
package noinput

import (
	"os"

	"github.com/mindfiredigital/DeepScanBot/packages/exitcode"
)

// IsInteractive returns true when both:
//   - the --no-input flag has NOT been set, AND
//   - stdin is connected to a terminal (TTY).
//
// When this returns false, the CLI must not prompt the user.
var IsInteractive = func() bool {
	return !noInputFlag && isTerminal(os.Stdin)
}

// noInputFlag is set by the --no-input global flag.
var noInputFlag bool

// SetNoInputFlag enables non-interactive mode regardless of TTY state.
func SetNoInputFlag() {
	noInputFlag = true
}

// RequireInteractive checks whether the CLI is running in interactive
// mode.  If not, it prints an actionable error and exits.
//
// Usage:
//
//	noinput.RequireInteractive("output filename", "use --force to overwrite")
func RequireInteractive(context, alternative string) {
	if !IsInteractive() {
		exitcode.HandleError(&exitcode.ExitCode{
			Code:    exitcode.InvalidInput,
			Message: "Cannot prompt for " + context + " in non-interactive mode.",
			Hint:    alternative,
		})
	}
}

// isTerminal returns true if the given file is a terminal.
// On platforms where term.IsTerminal is unavailable, it falls back
// to checking that the file is a character device (Unix).
var isTerminal = func(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	// Character device (Unix terminals are /dev/ttysxxx)
	return (stat.Mode() & os.ModeCharDevice) != 0
}