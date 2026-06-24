package tests

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"web-crawler-assignment/storage"
)

func TestEnhancedCrawlReportJSONSchema(t *testing.T) {
	entries := []storage.URLEntry{
		{URL: "https://example.com", Source: "href", Depth: 0, StatusCode: 200, ContentType: "text/html", Result: "passed", Attempts: 1},
		{URL: "https://example.com/about", Source: "href", Depth: 1, StatusCode: 200, ContentType: "text/html", Result: "passed", Attempts: 1},
		{URL: "https://example.com/404", Source: "href", Depth: 1, StatusCode: 404, ContentType: "text/html", Result: "failed", Error: "bad status code: 404", Attempts: 1},
		{URL: "https://example.com/discovered-link", Source: "href", Depth: 2, Result: "discovered"},
		{URL: "https://example.com/private", Source: "href", Depth: 1, Result: "skipped", SkippedReason: "disallowed by robots.txt"},
		{URL: "https://example.com/dupe", Source: "href", Depth: 1, Result: "skipped", SkippedReason: "duplicate"},
		{URL: "https://external.com/page", Source: "href", Depth: 1, Result: "skipped", SkippedReason: "outside domain scope"},
		{URL: "https://example.com/deep", Source: "href", Depth: 5, Result: "skipped", SkippedReason: "max depth exceeded"},
	}

	startedAt := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	finishedAt := time.Date(2026, 6, 23, 12, 0, 5, 0, time.UTC)

	report := storage.NewCrawlReport("https://example.com", "test.json", startedAt, finishedAt, entries)

	// Verify summary counts
	if report.Summary.Total != 8 {
		t.Fatalf("Total = %d, want 8", report.Summary.Total)
	}

	if report.Summary.Passed != 2 {
		t.Fatalf("Passed = %d, want 2", report.Summary.Passed)
	}

	if report.Summary.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", report.Summary.Failed)
	}

	if report.Summary.Skipped != 4 {
		t.Fatalf("Skipped = %d, want 4", report.Summary.Skipped)
	}

	if report.Summary.Discovered != 1 {
		t.Fatalf("Discovered = %d, want 1", report.Summary.Discovered)
	}

	if report.Summary.MaxDepth != 5 {
		t.Fatalf("MaxDepth = %d, want 5", report.Summary.MaxDepth)
	}

	if report.DurationMS != 5000 {
		t.Fatalf("DurationMS = %d, want 5000", report.DurationMS)
	}

	// Verify new breakdown fields
	if report.Summary.SkippedByRobots != 1 {
		t.Fatalf("SkippedByRobots = %d, want 1", report.Summary.SkippedByRobots)
	}

	if report.Summary.SkippedByDuplicate != 1 {
		t.Fatalf("SkippedByDuplicate = %d, want 1", report.Summary.SkippedByDuplicate)
	}

	if report.Summary.SkippedByDomain != 1 {
		t.Fatalf("SkippedByDomain = %d, want 1", report.Summary.SkippedByDomain)
	}

	if report.Summary.SkippedByDepth != 1 {
		t.Fatalf("SkippedByDepth = %d, want 1", report.Summary.SkippedByDepth)
	}

	// Verify status code distribution
	if report.Summary.URLsByStatusCode[200] != 2 {
		t.Fatalf("URLsByStatusCode[200] = %d, want 2", report.Summary.URLsByStatusCode[200])
	}

	if report.Summary.URLsByStatusCode[404] != 1 {
		t.Fatalf("URLsByStatusCode[404] = %d, want 1", report.Summary.URLsByStatusCode[404])
	}

	// Verify skipped_by_reason breakdown
	if report.Summary.SkippedByReason["disallowed by robots.txt"] != 1 {
		t.Fatalf("SkippedByReason[disallowed] = %d, want 1", report.Summary.SkippedByReason["disallowed by robots.txt"])
	}

	if report.Summary.SkippedByReason["duplicate"] != 1 {
		t.Fatalf("SkippedByReason[duplicate] = %d, want 1", report.Summary.SkippedByReason["duplicate"])
	}
}

func TestCrawlReportRetryDistribution(t *testing.T) {
	entries := []storage.URLEntry{
		{URL: "https://example.com", Result: "passed", Attempts: 1},
		{URL: "https://example.com/retry1", Result: "passed", Attempts: 2},
		{URL: "https://example.com/retry2", Result: "passed", Attempts: 3},
		{URL: "https://example.com/retry3", Result: "passed", Attempts: 2},
		{URL: "https://example.com/failed", Result: "failed", Attempts: 4},
	}

	report := storage.NewCrawlReport("https://example.com", "", time.Time{}, time.Time{}, entries)

	// All entries with attempts > 1 are counted as retried
	if report.Summary.RetriedRequests != 4 {
		t.Fatalf("RetriedRequests = %d, want 4", report.Summary.RetriedRequests)
	}

	if report.Summary.RetryDistribution[2] != 2 {
		t.Fatalf("RetryDistribution[2] = %d, want 2", report.Summary.RetryDistribution[2])
	}

	if report.Summary.RetryDistribution[3] != 1 {
		t.Fatalf("RetryDistribution[3] = %d, want 1", report.Summary.RetryDistribution[3])
	}
}

func TestEnhancedJSONRoundtrip(t *testing.T) {
	dir := t.TempDir()
	filename := dir + "/test_report.json"

	entries := []storage.URLEntry{
		{URL: "https://example.com", Source: "href", Depth: 0, StatusCode: 200, Result: "passed", Attempts: 1},
		{URL: "https://example.com/skip", Source: "href", Depth: 1, Result: "skipped", SkippedReason: "duplicate"},
	}

	startedAt := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	finishedAt := time.Date(2026, 6, 23, 12, 0, 5, 0, time.UTC)
	report := storage.NewCrawlReport("https://example.com", filename, startedAt, finishedAt, entries)

	if err := storage.WriteJSONReportToFile(filename, report); err != nil {
		t.Fatalf("write JSON report: %v", err)
	}

	// Read back and verify
	loadedEntries, err := storage.ReadEntriesFromFile(filename)
	if err != nil {
		t.Fatalf("read entries from file: %v", err)
	}

	if len(loadedEntries) != 2 {
		t.Fatalf("loaded %d entries, want 2", len(loadedEntries))
	}

	// Verify the file contains enhanced schema fields
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var rawReport map[string]interface{}
	if err := json.Unmarshal(data, &rawReport); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	summary, ok := rawReport["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("summary field missing or not an object")
	}

	// Check enhanced fields exist
	if _, exists := summary["urls_by_status_code"]; !exists {
		t.Error("urls_by_status_code missing from summary")
	}

	if _, exists := summary["skipped_by_reason"]; !exists {
		t.Error("skipped_by_reason missing from summary")
	}

	if _, exists := summary["skipped_by_depth"]; !exists {
		t.Error("skipped_by_depth missing from summary")
	}
}
