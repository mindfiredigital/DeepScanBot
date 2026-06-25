package storage

import "time"

// NewCrawlReport builds a CrawlReport from crawled URL entries.
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
