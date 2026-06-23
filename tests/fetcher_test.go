package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"web-crawler-assignment/fetcher"
)

func TestFetchProxyAndInsecure(t *testing.T) {
	_, _, _, _, err := fetcher.Fetch("https://localhost:9999", 1*time.Second, "http://127.0.0.1:8080", false, true, -1, nil)
	if err == nil {
		t.Log("Expected connection error or timeout, since proxy/server doesn't exist, got nil error")
	}

	_, _, _, _, err = fetcher.Fetch("https://example.com", 1*time.Second, "::invalid-proxy-url", false, true, -1, nil)
	if err == nil {
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
	_, _, _, statusCode, err = fetcher.Fetch(server.URL+"/404", 2*time.Second, "", false, false, -1, nil)
	if err == nil {
		t.Error("Expected error for 404 status code, got nil")
	}

	if statusCode != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", statusCode, http.StatusNotFound)
	}

	// Test 500
	_, _, _, statusCode, err = fetcher.Fetch(server.URL+"/500", 2*time.Second, "", false, false, -1, nil)
	if err == nil {
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

	_, _, _, _, err := fetcher.Fetch(server.URL, 2*time.Second, "", false, false, -1, nil)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
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

	body, _, _, _, err := fetcher.Fetch(server.URL+"/text", 2*time.Second, "", false, false, -1, allowed)
	if err != nil {
		t.Fatalf("fetch unconfigured content type: %v", err)
	}

	if body != nil {
		t.Errorf("unconfigured content type body = %q, want nil", body)
	}
}
