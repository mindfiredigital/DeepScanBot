package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchWithDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/retry-after-seconds":
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)

		case "/retry-after-httpdate":
			retryTime := time.Now().Add(3 * time.Second).Format(time.RFC1123)
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("Retry-After", retryTime)
			w.WriteHeader(http.StatusTooManyRequests)

		case "/success":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>ok</body></html>"))

		case "/no-retry-after":
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)

		case "/bad-content-type":
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%PDF-1.4..."))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		path    string
		wantErr bool
		check   func(t *testing.T, result FetchResult)
	}{
		{
			name:    "ReturnsRetryAfter",
			path:    "/retry-after-seconds",
			wantErr: true,
			check: func(t *testing.T, result FetchResult) {
				if result.StatusCode != http.StatusTooManyRequests {
					t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusTooManyRequests)
				}
				if result.RetryAfter != 5*time.Second {
					t.Fatalf("RetryAfter = %v, want 5s", result.RetryAfter)
				}
				if result.Err == nil {
					t.Fatal("expected error for 429, got nil")
				}
			},
		},
		{
			name:    "RetryAfterHTTPDate",
			path:    "/retry-after-httpdate",
			wantErr: true,
			check: func(t *testing.T, result FetchResult) {
				if result.StatusCode != http.StatusTooManyRequests {
					t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusTooManyRequests)
				}
				// Should be approximately 3 seconds (give or take processing time)
				if result.RetryAfter < 2*time.Second || result.RetryAfter > 5*time.Second {
					t.Fatalf("RetryAfter = %v, want approximately 3s", result.RetryAfter)
				}
			},
		},
		{
			name:    "Success",
			path:    "/success",
			wantErr: false,
			check: func(t *testing.T, result FetchResult) {
				if result.Err != nil {
					t.Fatalf("unexpected error: %v", result.Err)
				}
				if result.StatusCode != http.StatusOK {
					t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusOK)
				}
				if len(result.Body) == 0 {
					t.Fatal("body is empty, expected content")
				}
				if result.ContentType != "text/html; charset=utf-8" {
					t.Fatalf("content type = %q, want text/html; charset=utf-8", result.ContentType)
				}
			},
		},
		{
			name:    "RetryAfterNoHeader",
			path:    "/no-retry-after",
			wantErr: true,
			check: func(t *testing.T, result FetchResult) {
				if result.RetryAfter != 0 {
					t.Fatalf("RetryAfter = %v, want 0 for no header", result.RetryAfter)
				}
			},
		},
		{
			name:    "BadContentType",
			path:    "/bad-content-type",
			wantErr: false,
			check: func(t *testing.T, result FetchResult) {
				if result.Err != nil {
					t.Fatalf("unexpected error: %v", result.Err)
				}
				if result.Body != nil {
					t.Fatal("expected nil body for disallowed content type")
				}
				if result.Size <= 0 {
					t.Fatal("expected positive size even for disallowed content type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FetchWithDetails(server.URL+tt.path, 2*time.Second, "", false, false, -1, []string{"text/html"})
			if tt.wantErr && result.Err == nil {
				t.Fatal("expected error, got nil")
			}
			tt.check(t, result)
		})
	}
}