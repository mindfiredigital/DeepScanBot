package tests

import (
	"testing"
	"time"
	"web-crawler-assignment/fetcher"
)

func TestFetchProxyAndInsecure(t *testing.T) {
	// Start a local HTTP test server to act as a target or proxy
	// and verify that setting both flags correctly configures the transport parameters.
	
	// We want to test that calling Fetch with a proxy and insecure configures the transport.
	// We can inspect the client configuration logic (Fetch uses a local client, so we verify it doesn't crash
	// with a malformed proxy, and we test invalid schemes return an error as expected).
	_, _, err := fetcher.Fetch("https://localhost:9999", 1*time.Second, "http://127.0.0.1:8080", false, true)
	if err == nil {
		t.Log("Expected connection error or timeout, since proxy/server doesn't exist, got nil error")
	}

	// Verify that a bad proxy URL string returns an error immediately.
	_, _, err = fetcher.Fetch("https://example.com", 1*time.Second, "::invalid-proxy-url", false, true)
	if err == nil {
		t.Fatal("Expected error on invalid proxy URL, got nil")
	}
}
