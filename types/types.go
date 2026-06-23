package types

import "time"

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

// CrawlSummary holds aggregate statistics from a crawl.
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

// CrawlReport contains the complete output of a crawl session.
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