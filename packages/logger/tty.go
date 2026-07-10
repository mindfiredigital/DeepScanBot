package logger

import (
	"os"
)

// IsTTY returns true if the output is connected to a terminal (TTY)
func IsTTY() bool {
	// os.Stdout is already *os.File, no need for type assertion
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	
	// On Unix systems, check if it's a character device (terminal)
	// On Windows, this will be handled differently
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// IsTerminal returns true if stderr is connected to a terminal
// This is more reliable for progress output since logs go to stderr
func IsTerminal() bool {
	// os.Stderr is already *os.File, no need for type assertion
	stat, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	
	return (stat.Mode() & os.ModeCharDevice) != 0
}
