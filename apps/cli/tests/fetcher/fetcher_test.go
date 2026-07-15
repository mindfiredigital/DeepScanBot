package fetcher_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mindfiredigital/DeepScanBot/packages/fetcher"
)

func TestFetchWithDetailsReturnsRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	result := fetcher.FetchWithDetails(server.URL, 2*time.Second, "", false, false, -1, []string{"text/html"})
	if result.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusTooManyRequests)
	}

	if result.RetryAfter != 5*time.Second {
		t.Fatalf("RetryAfter = %v, want 5s", result.RetryAfter)
	}

	if result.Err == nil {
		t.Fatal("expected error for 429, got nil")
	}
}

func TestFetchWithDetailsRetryAfterHTTPDate(t *testing.T) {
	retryTime := time.Now().Add(3 * time.Second).Format(time.RFC1123)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Retry-After", retryTime)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	result := fetcher.FetchWithDetails(server.URL, 2*time.Second, "", false, false, -1, []string{"text/html"})
	if result.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusTooManyRequests)
	}
	// Should be approximately 3 seconds (give or take processing time)
	if result.RetryAfter < 2*time.Second || result.RetryAfter > 5*time.Second {
		t.Fatalf("RetryAfter = %v, want approximately 3s", result.RetryAfter)
	}
}

func TestFetchWithDetailsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	result := fetcher.FetchWithDetails(server.URL, 2*time.Second, "", false, false, -1, []string{"text/html"})
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
}

func TestFetchRetryAfterNoHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	result := fetcher.FetchWithDetails(server.URL, 2*time.Second, "", false, false, -1, []string{"text/html"})
	if result.RetryAfter != 0 {
		t.Fatalf("RetryAfter = %v, want 0 for no header", result.RetryAfter)
	}
}

func TestFetchWithDetailsBadContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-1.4..."))
	}))
	defer server.Close()

	result := fetcher.FetchWithDetails(server.URL, 2*time.Second, "", false, false, -1, []string{"text/html"})
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}

	if result.Body != nil {
		t.Fatal("expected nil body for disallowed content type")
	}

	if result.Size <= 0 {
		t.Fatal("expected positive size even for disallowed content type")
	}
}

func TestFetchProxyAndInsecure(t *testing.T) {
	//nolint:dogsled // Fetch returns multiple values, only error is relevant for this test
	_, _, _, _, fetchErr := fetcher.Fetch("https://localhost:9999", 1*time.Second, "http://127.0.0.1:8080", false, true, -1, nil)
	if fetchErr == nil {
		t.Log("Expected connection error or timeout, since proxy/server doesn't exist, got nil error")
	}

	//nolint:dogsled // Fetch returns multiple values, only error is relevant for this test
	_, _, _, _, fetchErr = fetcher.Fetch("https://example.com", 1*time.Second, "::invalid-proxy-url", false, true, -1, nil)
	if fetchErr == nil {
		t.Fatal("Expected error on invalid proxy URL, got nil")
	}
}

func TestFetchHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if r.URL.Path == "/404" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.URL.Path == "/500" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// Test 200 OK
	body, _, _, statusCode, err := fetcher.Fetch(server.URL+"/200", 2*time.Second, "", false, false, -1, nil)
	if err != nil {
		t.Fatalf("Expected success for 200 OK, got error: %v", err)
	}

	if string(body) != "success" {
		t.Errorf("Expected body 'success', got '%s'", string(body))
	}

	if statusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", statusCode, http.StatusOK)
	}

	// Test 404
	//nolint:dogsled // Fetch returns multiple values, statusCode and error are relevant
	_, _, _, statusCode, fetchErr := fetcher.Fetch(server.URL+"/404", 2*time.Second, "", false, false, -1, nil)
	if fetchErr == nil {
		t.Error("Expected error for 404 status code, got nil")
	}

	if statusCode != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", statusCode, http.StatusNotFound)
	}

	// Test 500
	//nolint:dogsled // Fetch returns multiple values, only error is relevant for 500
	_, _, _, _, fetchErr = fetcher.Fetch(server.URL+"/500", 2*time.Second, "", false, false, -1, nil)
	if fetchErr == nil {
		t.Error("Expected error for 500 status code, got nil")
	}
}

func TestFetchUserAgent(t *testing.T) {
	var capturedUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	//nolint:dogsled // Fetch returns multiple values, only error is relevant
	_, _, _, _, fetchErr := fetcher.Fetch(server.URL, 2*time.Second, "", false, false, -1, nil)
	if fetchErr != nil {
		t.Fatalf("Fetch failed: %v", fetchErr)
	}

	expectedUA := "DeepScanBot/1.0"
	if capturedUserAgent != expectedUA {
		t.Errorf("Expected User-Agent '%s', got '%s'", expectedUA, capturedUserAgent)
	}
}

func TestFetchConfiguredContentTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case "/pdf":
			w.Header().Set("Content-Type", "application/pdf")
		case "/jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		default:
			w.Header().Set("Content-Type", "text/plain")
		}

		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer server.Close()

	allowed := []string{"text/html", "application/pdf", "image/*"}

	for _, path := range []string{"/html", "/pdf", "/jpeg"} {
		t.Run(path, func(t *testing.T) {
			body, _, _, _, err := fetcher.Fetch(server.URL+path, 2*time.Second, "", false, false, -1, allowed)
			if err != nil {
				t.Fatalf("fetch configured content type: %v", err)
			}

			if got, want := string(body), path; got != want {
				t.Errorf("body = %q, want %q", got, want)
			}
		})
	}

	//nolint:dogsled // Fetch returns multiple values, body and error are relevant
	body, _, _, _, fetchErr := fetcher.Fetch(server.URL+"/text", 2*time.Second, "", false, false, -1, allowed)
	if fetchErr != nil {
		t.Fatalf("fetch unconfigured content type: %v", fetchErr)
	}

	if body != nil {
		t.Errorf("unconfigured content type body = %q, want nil", body)
	}
}
