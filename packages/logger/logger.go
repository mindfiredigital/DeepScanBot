package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/mindfiredigital/DeepScanBot/packages/exitcode"
)

// Logger wraps slog.Logger to provide a consistent logging interface.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger that writes to stderr with the given level.
func New(level string) *Logger {
	var l slog.Level

	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
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

	return &Logger{slog.New(handler)}
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
