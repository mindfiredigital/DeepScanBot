package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PageStorage struct {
	visitedUrls map[string]bool
	urlSource   map[string]string
	mutex       sync.Mutex
	results     []URLEntry
}

// URLEntry is one crawled URL and, when known, the HTML element that referenced it.
type URLEntry struct {
	URL           string `json:"url"`
	Source        string `json:"source,omitempty"`
	Depth         int    `json:"depth"`
	StatusCode    int    `json:"status_code,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	Result        string `json:"result,omitempty"`
	Error         string `json:"error,omitempty"`
	SkippedReason string `json:"skipped_reason,omitempty"`
	Attempts      int    `json:"attempts,omitempty"`
}

func NewPageStorage() *PageStorage {
	return &PageStorage{
		visitedUrls: make(map[string]bool),
		urlSource:   make(map[string]string),
		results:     []URLEntry{},
	}
}

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

func (ps *PageStorage) MarkVisited(url string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.visitedUrls[url] = true
}

func (ps *PageStorage) MarkVisitedIfNew(url string) bool {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	if ps.visitedUrls[url] {
		return false
	}
	ps.visitedUrls[url] = true
	return true
}

func (ps *PageStorage) HasVisited(url string) bool {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	return ps.visitedUrls[url]
}

func (ps *PageStorage) StoreSource(url, source string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.urlSource[url] = source
}

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

func (ps *PageStorage) StoreSkipped(url string, depth int, reason string) {
	ps.StoreEntry(URLEntry{
		URL:           url,
		Depth:         depth,
		Result:        "skipped",
		SkippedReason: reason,
	})
}

func (ps *PageStorage) StoreEntry(entry URLEntry) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	if entry.Source == "" {
		entry.Source = ps.urlSource[entry.URL]
	}

	ps.results = append(ps.results, entry)
}

// Results returns a snapshot of the URLs collected during a crawl.
func (ps *PageStorage) Results() []URLEntry {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	return append([]URLEntry(nil), ps.results...)
}

func WriteJSONToFile(filename string, entries []URLEntry) error {
	return WriteJSONReportToFile(filename, NewCrawlReport("", "", time.Time{}, time.Time{}, entries))
}

type CrawlSummary struct {
	Total               int            `json:"total"`
	Passed              int            `json:"passed"`
	Failed              int            `json:"failed"`
	Skipped             int            `json:"skipped"`
	Discovered          int            `json:"discovered"`
	SkippedByRobots     int            `json:"skipped_by_robots"`
	SkippedByDomain     int            `json:"skipped_by_domain"`
	SkippedByDuplicate  int            `json:"skipped_by_duplicate"`
	SkippedByContent    int            `json:"skipped_by_content_type"`
	SkippedByDepth      int            `json:"skipped_by_depth"`
	SkippedByOther      int            `json:"skipped_by_other"`
	RetriedRequests     int            `json:"retried_requests"`
	MaxDepth            int            `json:"max_depth"`
	URLsByStatusCode    map[int]int    `json:"urls_by_status_code,omitempty"`
	SkippedByReason     map[string]int `json:"skipped_by_reason,omitempty"`
	RetryDistribution   map[int]int    `json:"retry_distribution,omitempty"`
}

type CrawlReport struct {
	StartURL   string       `json:"start_url,omitempty"`
	OutputFile string       `json:"output_file,omitempty"`
	StartedAt  time.Time    `json:"started_at,omitempty"`
	FinishedAt time.Time    `json:"finished_at,omitempty"`
	DurationMS int64        `json:"duration_ms,omitempty"`
	Summary    CrawlSummary `json:"summary"`
	URLs       []URLEntry   `json:"urls"`
	Skipped    []URLEntry   `json:"skipped,omitempty"`
}

func NewCrawlReport(startURL, outputFile string, startedAt, finishedAt time.Time, entries []URLEntry) CrawlReport {
	report := CrawlReport{
		StartURL:   startURL,
		OutputFile: outputFile,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		URLs:       make([]URLEntry, 0, len(entries)),
		Skipped:    []URLEntry{},
		Summary: CrawlSummary{
			URLsByStatusCode:  make(map[int]int),
			SkippedByReason:   make(map[string]int),
			RetryDistribution: make(map[int]int),
		},
	}
	if !startedAt.IsZero() && !finishedAt.IsZero() {
		report.DurationMS = finishedAt.Sub(startedAt).Milliseconds()
	}

	for _, entry := range entries {
		report.Summary.Total++
		if entry.Depth > report.Summary.MaxDepth {
			report.Summary.MaxDepth = entry.Depth
		}
		if entry.Attempts > 1 {
			report.Summary.RetriedRequests++
			report.Summary.RetryDistribution[entry.Attempts]++
		}
		if entry.StatusCode > 0 {
			report.Summary.URLsByStatusCode[entry.StatusCode]++
		}
		switch entry.Result {
		case "skipped":
			report.Summary.Skipped++
			switch entry.SkippedReason {
			case "disallowed by robots.txt":
				report.Summary.SkippedByRobots++
			case "outside domain scope":
				report.Summary.SkippedByDomain++
			case "duplicate":
				report.Summary.SkippedByDuplicate++
			case "content type not allowed":
				report.Summary.SkippedByContent++
			case "max depth exceeded":
				report.Summary.SkippedByDepth++
			default:
				report.Summary.SkippedByOther++
			}
			report.Summary.SkippedByReason[entry.SkippedReason]++
			report.Skipped = append(report.Skipped, entry)
		case "failed":
			report.Summary.Failed++
			report.URLs = append(report.URLs, entry)
		case "discovered":
			report.Summary.Discovered++
			report.URLs = append(report.URLs, entry)
		default:
			report.Summary.Passed++
			report.URLs = append(report.URLs, entry)
		}
	}
	if len(report.Skipped) == 0 {
		report.Skipped = nil
	}
	if len(report.Summary.URLsByStatusCode) == 0 {
		report.Summary.URLsByStatusCode = nil
	}
	if len(report.Summary.SkippedByReason) == 0 {
		report.Summary.SkippedByReason = nil
	}
	if len(report.Summary.RetryDistribution) == 0 {
		report.Summary.RetryDistribution = nil
	}

	return report
}

func WriteJSONReportToFile(filename string, report CrawlReport) error {
	report.OutputFile = filename
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func ReadEntriesFromFile(filename string) ([]URLEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var report CrawlReport
	if err := json.Unmarshal(data, &report); err == nil && (len(report.URLs) > 0 || len(report.Skipped) > 0) {
		return append(report.URLs, report.Skipped...), nil
	}

	var legacy struct {
		URLs []URLEntry `json:"urls"`
	}
	if err := json.Unmarshal(data, &legacy); err == nil {
		return legacy.URLs, nil
	}

	return readTextEntries(data), nil
}

func readTextEntries(data []byte) []URLEntry {
	var entries []URLEntry
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		source := ""
		if strings.HasPrefix(line, "[") {
			if idx := strings.Index(line, "] "); idx > 0 {
				source = strings.TrimPrefix(line[:idx], "[")
				line = line[idx+2:]
			}
		}
		if idx := strings.Index(line, " ["); idx >= 0 {
			line = line[:idx]
		}
		entries = append(entries, URLEntry{URL: line, Source: source})
	}
	return entries
}

func WriteTextToFile(filename string, entries []URLEntry, showSource bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range entries {
		line := entry.URL
		if showSource && entry.Source != "" {
			line = "[" + entry.Source + "] " + line
		}
		if entry.StatusCode != 0 {
			line += " [status=" + strconv.Itoa(entry.StatusCode) + "]"
		}
		if entry.Result != "" {
			line += " [result=" + entry.Result + "]"
		}
		if entry.SkippedReason != "" {
			line += " [skipped=" + entry.SkippedReason + "]"
		}
		if entry.Attempts > 1 {
			line += " [attempts=" + strconv.Itoa(entry.Attempts) + "]"
		}
		if entry.Error != "" {
			line += " [error=" + entry.Error + "]"
		}
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}
