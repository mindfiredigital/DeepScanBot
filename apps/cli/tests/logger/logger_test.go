package tests

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mindfiredigital/DeepScanBot/packages/logger"
)

func TestNewWithLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       logger.LogLevel
		wantQuiet   bool
		wantVerbose bool
		wantDebug   bool
	}{
		{
			name:      "quiet level",
			level:     logger.LevelQuiet,
			wantQuiet: true,
		},
		{
			name:  "info level",
			level: logger.LevelInfo,
		},
		{
			name:        "verbose level",
			level:       logger.LevelVerbose,
			wantVerbose: true,
		},
		{
			name:        "debug level",
			level:       logger.LevelDebug,
			wantVerbose: true,
			wantDebug:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lgr := logger.NewWithLevel(tt.level)

			if lgr.IsQuiet() != tt.wantQuiet {
				t.Errorf("IsQuiet() = %v, want %v", lgr.IsQuiet(), tt.wantQuiet)
			}
			if lgr.IsVerbose() != tt.wantVerbose {
				t.Errorf("IsVerbose() = %v, want %v", lgr.IsVerbose(), tt.wantVerbose)
			}
			if lgr.IsDebug() != tt.wantDebug {
				t.Errorf("IsDebug() = %v, want %v", lgr.IsDebug(), tt.wantDebug)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	lgr := logger.NewWithLevel(logger.LevelInfo)

	// Test setting to debug
	lgr.SetLevel(logger.LevelDebug)
	if !lgr.IsDebug() {
		t.Error("Expected IsDebug() to be true after SetLevel(logger.LevelDebug)")
	}

	// Test setting to quiet
	lgr.SetLevel(logger.LevelQuiet)
	if !lgr.IsQuiet() {
		t.Error("Expected IsQuiet() to be true after SetLevel(logger.LevelQuiet)")
	}

	// Test setting to verbose
	lgr.SetLevel(logger.LevelVerbose)
	if !lgr.IsVerbose() {
		t.Error("Expected IsVerbose() to be true after SetLevel(logger.LevelVerbose)")
	}
}

func TestLoggingLevels(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	lgr := logger.NewWithLevel(logger.LevelDebug)

	lgr.Debugf("debug message")
	lgr.Infof("info message")
	lgr.Warnf("warn message")
	lgr.Errorf("error message")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "debug message") {
		t.Error("Debug message not found in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message not found in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message not found in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message not found in output")
	}
}

func TestQuietLevelFiltersMessages(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	lgr := logger.NewWithLevel(logger.LevelQuiet)

	lgr.Debugf("debug should not appear")
	lgr.Infof("info should not appear")
	lgr.Warnf("warn should appear")
	lgr.Errorf("error should appear")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if strings.Contains(output, "debug should not appear") {
		t.Error("Debug message should not appear in quiet mode")
	}
	if strings.Contains(output, "info should not appear") {
		t.Error("Info message should not appear in quiet mode")
	}
	if !strings.Contains(output, "warn should appear") {
		t.Error("Warn message should appear in quiet mode")
	}
	if !strings.Contains(output, "error should appear") {
		t.Error("Error message should appear in quiet mode")
	}
}

func TestVerboseLevelFiltersMessages(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	lgr := logger.NewWithLevel(logger.LevelVerbose)

	lgr.Debugf("debug should not appear")
	lgr.Infof("info should appear")
	lgr.Warnf("warn should appear")
	lgr.Errorf("error should appear")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if strings.Contains(output, "debug should not appear") {
		t.Error("Debug message should not appear in verbose mode")
	}
	if !strings.Contains(output, "info should appear") {
		t.Error("Info message should appear in verbose mode")
	}
	if !strings.Contains(output, "warn should appear") {
		t.Error("Warn message should appear in verbose mode")
	}
	if !strings.Contains(output, "error should appear") {
		t.Error("Error message should appear in verbose mode")
	}
}

func TestDebugLevelShowsAllMessages(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	lgr := logger.NewWithLevel(logger.LevelDebug)

	lgr.Debugf("debug should appear")
	lgr.Infof("info should appear")
	lgr.Warnf("warn should appear")
	lgr.Errorf("error should appear")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "debug should appear") {
		t.Error("Debug message should appear in debug mode")
	}
	if !strings.Contains(output, "info should appear") {
		t.Error("Info message should appear in debug mode")
	}
	if !strings.Contains(output, "warn should appear") {
		t.Error("Warn message should appear in debug mode")
	}
	if !strings.Contains(output, "error should appear") {
		t.Error("Error message should appear in debug mode")
	}
}
