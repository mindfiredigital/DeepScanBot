package storage

import (
	"sync"

	"github.com/mindfiredigital/DeepScanBot/packages/types"
)

// Type aliases for backward compatibility — actual definitions are in types/types.go.
type (
	URLEntry     = types.URLEntry
	CrawlReport  = types.CrawlReport
	CrawlSummary = types.CrawlSummary
)

// PageStorage tracks visited URLs and stores crawl results with thread safety.
type PageStorage struct {
	visitedUrls map[string]bool
	urlSource   map[string]string
	mutex       sync.Mutex
	results     []URLEntry
}

// NewPageStorage creates a new PageStorage.
func NewPageStorage() *PageStorage {
	return &PageStorage{
		visitedUrls: make(map[string]bool),
		urlSource:   make(map[string]string),
		results:     []URLEntry{},
	}
}

// SeedEntries pre-populates the storage with existing entries (e.g., for resume mode).
func (ps *PageStorage) SeedEntries(entries []URLEntry) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, entry := range entries {
		if entry.URL == "" {
			continue
		}

		ps.visitedUrls[entry.URL] = true
		if entry.Source != "" {
			ps.urlSource[entry.URL] = entry.Source
		}

		ps.results = append(ps.results, entry)
	}
}

// MarkVisited marks a URL as visited.
func (ps *PageStorage) MarkVisited(url string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.visitedUrls[url] = true
}

// MarkVisitedIfNew marks a URL as visited and returns true if it wasn't already visited.
func (ps *PageStorage) MarkVisitedIfNew(url string) bool {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if ps.visitedUrls[url] {
		return false
	}

	ps.visitedUrls[url] = true

	return true
}

// HasVisited checks if a URL has been visited.
func (ps *PageStorage) HasVisited(url string) bool {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	return ps.visitedUrls[url]
}

// StoreSource associates a URL with its source.
func (ps *PageStorage) StoreSource(url, source string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.urlSource[url] = source
}

// StoreEntry stores a URL entry, automatically populating the source if available.
func (ps *PageStorage) StoreEntry(entry URLEntry) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if entry.Source == "" {
		entry.Source = ps.urlSource[entry.URL]
	}

	ps.results = append(ps.results, entry)
}

// StoreContent records a discovered URL with zero values (backwards-compatible test helper).
func (ps *PageStorage) StoreContent(url string) {
	ps.StoreResult(url, 0, 0, "")
}

// StoreResult records one discovered or fetched URL and its crawl outcome.
func (ps *PageStorage) StoreResult(url string, depth, statusCode int, resultError string) {
	result := "passed"
	if statusCode == 0 && resultError == "" {
		result = "discovered"
	}

	if resultError != "" {
		result = "failed"
	}

	ps.StoreEntry(URLEntry{
		URL:        url,
		Depth:      depth,
		StatusCode: statusCode,
		Result:     result,
		Error:      resultError,
	})
}

// StoreSkipped records a skipped URL (backwards-compatible test helper).
func (ps *PageStorage) StoreSkipped(url string, depth int, reason string) {
	ps.StoreEntry(URLEntry{
		URL:           url,
		Depth:         depth,
		Result:        "skipped",
		SkippedReason: reason,
	})
}

// Results returns a snapshot of all stored URL entries.
func (ps *PageStorage) Results() []URLEntry {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	return append([]URLEntry(nil), ps.results...)
}
