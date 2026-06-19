package tests

import (
	"reflect"
	"testing"
	"web-crawler-assignment/parser"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		baseURL    string
		wantLinks  map[string]string
	}{
		{
			name:    "resolve relative href and src",
			baseURL: "https://example.com/blog/post1",
			body: []byte(`
				<html>
					<body>
						<a href="/about">About</a>
						<a href="contact">Contact</a>
						<a href="../home">Home</a>
						<a href="https://other.com/page">External</a>
						<script src="/static/js/app.js"></script>
						<script src="relative.js"></script>
					</body>
				</html>
			`),
			wantLinks: map[string]string{
				"https://example.com/about":         "href",
				"https://example.com/blog/contact":   "href",
				"https://example.com/home":          "href",
				"https://other.com/page":            "href",
				"https://example.com/static/js/app.js": "script",
				"https://example.com/blog/relative.js": "script",
			},
		},
		{
			name:    "invalid base URL",
			baseURL: ":invalid-url",
			body: []byte(`
				<a href="/about">About</a>
			`),
			wantLinks: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Parse(tt.body, tt.baseURL)
			if !reflect.DeepEqual(got, tt.wantLinks) {
				t.Errorf("Parse() = %v, want %v", got, tt.wantLinks)
			}
		})
	}
}
