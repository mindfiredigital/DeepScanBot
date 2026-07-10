package version

import (
	"fmt"
	"runtime"
	"time"
)

// Info holds version and build metadata
type Info struct {
	Version      string
	GitCommit    string
	BuildDate    string
	GoVersion    string
	Platform     string
}

// String returns a human-readable version string
func (v *Info) String() string {
	version := v.Version
	
	if v.GitCommit != "" {
		version += fmt.Sprintf(" (%s)", v.GitCommit)
	}
	
	if v.BuildDate != "" {
		version += fmt.Sprintf(" built on %s", v.BuildDate)
	}
	
	return version
}

// JSON returns a map suitable for JSON output
func (v *Info) JSON() map[string]string {
	data := map[string]string{
		"version":   v.Version,
		"name":      "DeepScanBot CLI",
	}
	
	if v.GitCommit != "" {
		data["git_commit"] = v.GitCommit
	}
	if v.BuildDate != "" {
		data["build_date"] = v.BuildDate
	}
	if v.GoVersion != "" {
		data["go_version"] = v.GoVersion
	}
	if v.Platform != "" {
		data["platform"] = v.Platform
	}
	
	return data
}

// Default returns version info with sensible defaults
func Default() *Info {
	return &Info{
		Version:   "dev",
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		BuildDate: time.Now().Format(time.RFC3339),
	}
}