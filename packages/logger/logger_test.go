package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestNewWithLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		wantQuiet   bool
		wantVerbose bool
		wantDebug   bool
	}{
		{
			name:     "quiet level",
			level:    LevelQuiet,
			wantQuiet: true,
		},
		{
			name:     "info level",
			level:    LevelInfo,
			wantQuiet: false,
		},
		{
			name:     "verbose level",
			level:    LevelVerbose,
			wantVerbose: true,
		},
		{
			name:     "debug level",
			level:    LevelDebug,
			wantVerbose: true,
			wantDebug:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewWithLevel(tt.level)
			
			if logger.IsQuiet() != tt.wantQuiet {
				t.Errorf("IsQuiet() = %v, want %v", logger.IsQuiet(), tt.wantQuiet)
			}
			if logger.IsVerbose() != tt.wantVerbose {
				t.Errorf("IsVerbose() = %v, want %v", logger.IsVerbose(), tt.wantVerbose)
			}
			if logger.IsDebug() != tt.wantDebug {
				t.Errorf("IsDebug() = %v, want %v", logger.IsDebug(), tt.wantDebug)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	logger := NewWithLevel(LevelInfo)
	
	// Test setting to debug
	logger.SetLevel(LevelDebug)
	if !logger.IsDebug() {
		t.Error("Expected IsDebug() to be true after SetLevel(LevelDebug)")
	}
	
	// Test setting to quiet
	logger.SetLevel(LevelQuiet)
	if !logger.IsQuiet() {
		t.Error("Expected IsQuiet() to be true after SetLevel(LevelQuiet)")
	}
	
	// Test setting to verbose
	logger.SetLevel(LevelVerbose)
	if !logger.IsVerbose() {
		t.Error("Expected IsVerbose() to be true after SetLevel(LevelVerbose)")
	}
}

func TestLoggingLevels(t *testing.T) {
	// Capture stderr to verify log output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	
	logger := NewWithLevel(LevelDebug)
	
	logger.Debugf("debug message")
	logger.Infof("info message")
	logger.Warnf("warn message")
	logger.Errorf("error message")
	
	w.Close()
	os.Stderr = oldStderr
	
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify all messages were logged
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
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	
	logger := NewWithLevel(LevelQuiet)
	
	logger.Debugf("debug should not appear")
	logger.Infof("info should not appear")
	logger.Warnf("warn should appear")
	logger.Errorf("error should appear")
	
	w.Close()
	os.Stderr = oldStderr
	
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify only warnings and errors appear
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
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	
	logger := NewWithLevel(LevelVerbose)
	
	logger.Debugf("debug should not appear")
	logger.Infof("info should appear")
	logger.Warnf("warn should appear")
	logger.Errorf("error should appear")
	
	w.Close()
	os.Stderr = oldStderr
	
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify info, warnings, and errors appear, but not debug
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
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	
	logger := NewWithLevel(LevelDebug)
	
	logger.Debugf("debug should appear")
	logger.Infof("info should appear")
	logger.Warnf("warn should appear")
	logger.Errorf("error should appear")
	
	w.Close()
	os.Stderr = oldStderr
	
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify all messages appear
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