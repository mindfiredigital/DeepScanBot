package types

import "time"

// MultiSiteReport contains crawl results for multiple seed URLs.
type MultiSiteReport struct {
	StartedAt  time.Time        `json:"started_at"`
	FinishedAt time.Time        `json:"finished_at"`
	DurationMS int64            `json:"duration_ms"`
	Sites      []SiteReport    `json:"sites"`
	Summary    MultiSiteSummary `json:"summary"`
}

// SiteReport contains the crawl report for a single seed URL.
type SiteReport struct {
	StartURL   string       `json:"start_url"`
	OutputFile string       `json:"output_file,omitempty"`
	StartedAt  time.Time    `json:"started_at,omitempty"`
	FinishedAt time.Time    `json:"finished_at,omitempty"`
	DurationMS int64        `json:"duration_ms"`
	Report     CrawlReport  `json:"report"`
}

// MultiSiteSummary holds aggregate statistics across all crawled sites.
type MultiSiteSummary struct {
	TotalSites     int `json:"total_sites"`
	TotalURLs      int `json:"total_urls"`
	TotalPassed    int `json:"total_passed"`
	TotalFailed    int `json:"total_failed"`
	TotalSkipped   int `json:"total_skipped"`
	TotalDiscovered int `json:"total_discovered"`
}

// NewMultiSiteReport creates a new empty MultiSiteReport.
func NewMultiSiteReport() MultiSiteReport {
	return MultiSiteReport{
		Sites: []SiteReport{},
		Summary: MultiSiteSummary{
			TotalSites: 0,
		},
	}
}

// AddSiteReport adds a site report and updates the aggregate summary.
func (msr *MultiSiteReport) AddSiteReport(siteReport SiteReport) {
	msr.Sites = append(msr.Sites, siteReport)
	msr.Summary.TotalSites++
	msr.Summary.TotalURLs += siteReport.Report.Summary.Total
	msr.Summary.TotalPassed += siteReport.Report.Summary.Passed
	msr.Summary.TotalFailed += siteReport.Report.Summary.Failed
	msr.Summary.TotalSkipped += siteReport.Report.Summary.Skipped
	msr.Summary.TotalDiscovered += siteReport.Report.Summary.Discovered
}

// Finalize sets the finished time and duration.
func (msr *MultiSiteReport) Finalize() {
	if !msr.FinishedAt.IsZero() && !msr.StartedAt.IsZero() {
		msr.DurationMS = msr.FinishedAt.Sub(msr.StartedAt).Milliseconds()
	}
}