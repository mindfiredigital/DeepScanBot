package main

import "testing"

func TestValidateStartURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
		err  bool
	}{
		{name: "HTTP", url: "http://example.com", want: "http://example.com"},
		{name: "HTTPS with path", url: "https://example.com/docs?q=1", want: "https://example.com/docs?q=1"},
		{name: "trims whitespace", url: "  https://example.com  ", want: "https://example.com"},
		{name: "missing", url: "", err: true},
		{name: "whitespace only", url: "   ", err: true},
		{name: "FTP", url: "ftp://example.com", err: true},
		{name: "file", url: "file:///etc/passwd", err: true},
		{name: "relative", url: "not-a-url", err: true},
		{name: "missing host", url: "http://", err: true},
		{name: "malformed HTTP", url: "http:/missing-slash.com", err: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateStartURL(tt.url)
			if (err != nil) != tt.err {
				t.Fatalf("validateStartURL(%q) error = %v, want error = %v", tt.url, err, tt.err)
			}
			if got != tt.want {
				t.Errorf("validateStartURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}
