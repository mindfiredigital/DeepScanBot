package logger

import (
	"os"
)

// IsTTY returns true if stdout is connected to a terminal.
func IsTTY() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// IsTerminal returns true if stderr is connected to a terminal.
func IsTerminal() bool {
	stat, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// IsInputTTY returns true if stdin is connected to a terminal.
func IsInputTTY() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
