package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"web-crawler-assignment/fetcher"
)

func TestFetchProxyAndInsecure(t *testing.T) {
	_, _, err := fetcher.Fetch("https://localhost:9999", 1*time.Second, "http://127.0.0.1:8080", false, true)
	if err == nil {
		t.Log("Expected connection error or timeout, since proxy/server doesn't exist, got nil error")
	}

	_, _, err = fetcher.Fetch("https://example.com", 1*time.Second, "::invalid-proxy-url", false, true)
	if err == nil {
		t.Fatal("Expected error on invalid proxy URL, got nil")
	}
}

func TestFetchHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	body, _, err := fetcher.Fetch(server.URL+"/200", 2*time.Second, "", false, false)
	if err != nil {
		t.Fatalf("Expected success for 200 OK, got error: %v", err)
	}
	if string(body) != "success" {
		t.Errorf("Expected body 'success', got '%s'", string(body))
	}

	// Test 404
	_, _, err = fetcher.Fetch(server.URL+"/404", 2*time.Second, "", false, false)
	if err == nil {
		t.Error("Expected error for 404 status code, got nil")
	}

	// Test 500
	_, _, err = fetcher.Fetch(server.URL+"/500", 2*time.Second, "", false, false)
	if err == nil {
		t.Error("Expected error for 500 status code, got nil")
	}
}

func TestFetchUserAgent(t *testing.T) {
	var capturedUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	_, _, err := fetcher.Fetch(server.URL, 2*time.Second, "", false, false)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	expectedUA := "DeepScanBot/1.0"
	if capturedUserAgent != expectedUA {
		t.Errorf("Expected User-Agent '%s', got '%s'", expectedUA, capturedUserAgent)
	}
}


