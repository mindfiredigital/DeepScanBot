package version

import (
	"testing"
)

func TestInfoString(t *testing.T) {
	tests := []struct {
		name     string
		info     *Info
		expected string
	}{
		{
			name: "version only",
			info: &Info{
				Version: "1.0.0",
			},
			expected: "1.0.0",
		},
		{
			name: "version with git commit",
			info: &Info{
				Version:   "1.0.0",
				GitCommit: "abc123",
			},
			expected: "1.0.0 (abc123)",
		},
		{
			name: "version with build date",
			info: &Info{
				Version:   "1.0.0",
				BuildDate: "2024-01-15T10:30:00Z",
			},
			expected: "1.0.0 built on 2024-01-15T10:30:00Z",
		},
		{
			name: "version with all metadata",
			info: &Info{
				Version:   "1.0.0",
				GitCommit: "abc123",
				BuildDate: "2024-01-15T10:30:00Z",
			},
			expected: "1.0.0 (abc123) built on 2024-01-15T10:30:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInfoJSON(t *testing.T) {
	info := &Info{
		Version:   "1.0.0",
		GitCommit: "abc123",
		BuildDate: "2024-01-15T10:30:00Z",
		GoVersion: "go1.21.0",
		Platform:  "linux/amd64",
	}

	result := info.JSON()
	
	if result["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", result["version"])
	}
	if result["name"] != "DeepScanBot CLI" {
		t.Errorf("Expected name 'DeepScanBot CLI', got '%s'", result["name"])
	}
	if result["git_commit"] != "abc123" {
		t.Errorf("Expected git_commit 'abc123', got '%s'", result["git_commit"])
	}
	if result["build_date"] != "2024-01-15T10:30:00Z" {
		t.Errorf("Expected build_date '2024-01-15T10:30:00Z', got '%s'", result["build_date"])
	}
	if result["go_version"] != "go1.21.0" {
		t.Errorf("Expected go_version 'go1.21.0', got '%s'", result["go_version"])
	}
	if result["platform"] != "linux/amd64" {
		t.Errorf("Expected platform 'linux/amd64', got '%s'", result["platform"])
	}
}

func TestInfoJSONMinimal(t *testing.T) {
	info := &Info{
		Version: "dev",
	}

	result := info.JSON()
	
	if result["version"] != "dev" {
		t.Errorf("Expected version 'dev', got '%s'", result["version"])
	}
	if result["name"] != "DeepScanBot CLI" {
		t.Errorf("Expected name 'DeepScanBot CLI', got '%s'", result["name"])
	}
	// These should not be present when empty
	if _, ok := result["git_commit"]; ok {
		t.Error("Expected git_commit to be absent when empty")
	}
	if _, ok := result["build_date"]; ok {
		t.Error("Expected build_date to be absent when empty")
	}
}

func TestDefault(t *testing.T) {
	info := Default()
	
	if info == nil {
		t.Fatal("Default() returned nil")
	}
	
	if info.Version != "dev" {
		t.Errorf("Expected default version 'dev', got '%s'", info.Version)
	}
	
	if info.GoVersion == "" {
		t.Error("Expected GoVersion to be set")
	}
	
	if info.Platform == "" {
		t.Error("Expected Platform to be set")
	}
	
	if info.BuildDate == "" {
		t.Error("Expected BuildDate to be set")
	}
}