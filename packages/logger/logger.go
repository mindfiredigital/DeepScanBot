package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/mindfiredigital/DeepScanBot/packages/exitcode"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelQuiet   LogLevel = "quiet"
	LevelInfo    LogLevel = "info"
	LevelVerbose LogLevel = "verbose"
	LevelDebug   LogLevel = "debug"
)

// Logger wraps slog.Logger to provide a consistent logging interface.
type Logger struct {
	*slog.Logger
	level LogLevel
}

// New creates a new Logger that writes to stderr with the given level.
func New(level string) *Logger {
	return NewWithLevel(LogLevel(level))
}

// NewWithLevel creates a new Logger with a specific LogLevel
func NewWithLevel(level LogLevel) *Logger {
	var l slog.Level

	switch level {
	case LevelDebug:
		l = slog.LevelDebug
	case LevelVerbose:
		l = slog.LevelInfo
	case LevelInfo:
		l = slog.LevelInfo
	case LevelQuiet:
		// Quiet mode shows warnings and errors, so use Warn level
		l = slog.LevelWarn
	default:
		l = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: l,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})

	return &Logger{slog.New(handler), level}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
	// Note: slog.HandlerOptions doesn't allow changing level after creation
	// For simplicity, we just update the level field
}

// Infof logs a formatted info message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}

// Fatalf logs a formatted error message and calls os.Exit(1).
//
// Deprecated: Use FatalfExit with an *exitcode.ExitCode to provide a
// meaningful exit code and actionable error message.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// FatalfExit logs a formatted error message and exits with the code
// carried by the given *exitcode.ExitCode.  The exit code is printed
// together with a hint when present.
//
// Usage:
//
//	log.FatalfExit("write results", exitcode.ErrWriteOutput)
func (l *Logger) FatalfExit(msg string, ec *exitcode.ExitCode) {
	fullMsg := fmt.Sprintf("%s: %s", msg, ec.Message)
	l.Log(context.Background(), slog.LevelError, fullMsg)
	exitcode.HandleError(ec)
}
