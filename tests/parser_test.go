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
			name:    "filter non-HTTP schemes",
			baseURL: "https://example.com/blog/post1",
			body: []byte(`
				<html>
					<body>
						<a href="mailto:info@example.com">Email</a>
						<a href="javascript:void(0)">JS</a>
						<a href="tel:+123456789">Tel</a>
						<a href="#section-1">Anchor</a>
						<a href="">Empty</a>
						<a href="   ">Spaces</a>
					</body>
				</html>
			`),
			wantLinks: map[string]string{},
		},
		{
			name:    "resolve expanded tags img link iframe form",
			baseURL: "https://example.com/blog/post1",
			body: []byte(`
				<html>
					<head>
						<link rel="stylesheet" href="/styles.css">
					</head>
					<body>
						<img src="logo.png">
						<iframe src="/embed/video"></iframe>
						<form action="/submit-form" method="post"></form>
					</body>
				</html>
			`),
			wantLinks: map[string]string{
				"https://example.com/styles.css":   "link",
				"https://example.com/blog/logo.png": "img",
				"https://example.com/embed/video":  "iframe",
				"https://example.com/submit-form":  "form",
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
