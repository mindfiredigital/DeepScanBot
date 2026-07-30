package types

import (
	"testing"
	"time"
)

func TestNewMultiSiteReport(t *testing.T) {
	report := NewMultiSiteReport()

	if report.Sites == nil {
		t.Error("NewMultiSiteReport().Sites should not be nil")
	}
	if len(report.Sites) != 0 {
		t.Errorf("NewMultiSiteReport().Sites should be empty, got %d", len(report.Sites))
	}
	if report.Summary.TotalSites != 0 {
		t.Errorf("NewMultiSiteReport().Summary.TotalSites should be 0, got %d", report.Summary.TotalSites)
	}
}

func TestAddSiteReport(t *testing.T) {
	tests := []struct {
		name     string
		reports  []SiteReport
		expected MultiSiteSummary
	}{
		{
			name: "single site report",
			reports: []SiteReport{
				{
					StartURL: "https://example.com",
					Report: CrawlReport{
						Summary: CrawlSummary{
							Total:      10,
							Passed:     8,
							Failed:     1,
							Skipped:    1,
							Discovered: 5,
						},
					},
				},
			},
			expected: MultiSiteSummary{
				TotalSites:      1,
				TotalURLs:       10,
				TotalPassed:     8,
				TotalFailed:     1,
				TotalSkipped:    1,
				TotalDiscovered: 5,
			},
		},
		{
			name: "multiple site reports accumulate correctly",
			reports: []SiteReport{
				{
					StartURL: "https://site1.com",
					Report: CrawlReport{
						Summary: CrawlSummary{
							Total:      10,
							Passed:     8,
							Failed:     1,
							Skipped:    1,
							Discovered: 3,
						},
					},
				},
				{
					StartURL: "https://site2.com",
					Report: CrawlReport{
						Summary: CrawlSummary{
							Total:      20,
							Passed:     15,
							Failed:     3,
							Skipped:    2,
							Discovered: 7,
						},
					},
				},
				{
					StartURL: "https://site3.com",
					Report: CrawlReport{
						Summary: CrawlSummary{
							Total:      5,
							Passed:     5,
							Failed:     0,
							Skipped:    0,
							Discovered: 2,
						},
					},
				},
			},
			expected: MultiSiteSummary{
				TotalSites:      3,
				TotalURLs:       35,
				TotalPassed:     28,
				TotalFailed:     4,
				TotalSkipped:    3,
				TotalDiscovered: 12,
			},
		},
		{
			name: "empty site report",
			reports: []SiteReport{
				{
					StartURL: "https://empty.com",
					Report: CrawlReport{
						Summary: CrawlSummary{},
					},
				},
			},
			expected: MultiSiteSummary{
				TotalSites: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewMultiSiteReport()
			for _, sr := range tt.reports {
				report.AddSiteReport(sr)
			}

			if report.Summary.TotalSites != tt.expected.TotalSites {
				t.Errorf("TotalSites = %d, want %d", report.Summary.TotalSites, tt.expected.TotalSites)
			}
			if report.Summary.TotalURLs != tt.expected.TotalURLs {
				t.Errorf("TotalURLs = %d, want %d", report.Summary.TotalURLs, tt.expected.TotalURLs)
			}
			if report.Summary.TotalPassed != tt.expected.TotalPassed {
				t.Errorf("TotalPassed = %d, want %d", report.Summary.TotalPassed, tt.expected.TotalPassed)
			}
			if report.Summary.TotalFailed != tt.expected.TotalFailed {
				t.Errorf("TotalFailed = %d, want %d", report.Summary.TotalFailed, tt.expected.TotalFailed)
			}
			if report.Summary.TotalSkipped != tt.expected.TotalSkipped {
				t.Errorf("TotalSkipped = %d, want %d", report.Summary.TotalSkipped, tt.expected.TotalSkipped)
			}
			if report.Summary.TotalDiscovered != tt.expected.TotalDiscovered {
				t.Errorf("TotalDiscovered = %d, want %d", report.Summary.TotalDiscovered, tt.expected.TotalDiscovered)
			}

			if len(report.Sites) != tt.expected.TotalSites {
				t.Errorf("len(Sites) = %d, want %d", len(report.Sites), tt.expected.TotalSites)
			}
		})
	}
}

func TestFinalize(t *testing.T) {
	t.Run("zero timestamps does not update duration", func(t *testing.T) {
		report := NewMultiSiteReport()
		report.Finalize()
		if report.DurationMS != 0 {
			t.Errorf("DurationMS should be 0 with zero timestamps, got %d", report.DurationMS)
		}
	})

	t.Run("valid timestamps sets correct duration", func(t *testing.T) {
		report := NewMultiSiteReport()
		now := time.Now()
		report.StartedAt = now
		report.FinishedAt = now.Add(5 * time.Second)
		report.Finalize()
		if report.DurationMS != 5000 {
			t.Errorf("DurationMS should be 5000 for 5s difference, got %d", report.DurationMS)
		}
	})

	t.Run("only started_at set does not update duration", func(t *testing.T) {
		report := NewMultiSiteReport()
		report.StartedAt = time.Now()
		report.Finalize()
		if report.DurationMS != 0 {
			t.Errorf("DurationMS should be 0 when FinishedAt is zero, got %d", report.DurationMS)
		}
	})

	t.Run("only finished_at set does not update duration", func(t *testing.T) {
		report := NewMultiSiteReport()
		report.FinishedAt = time.Now()
		report.Finalize()
		if report.DurationMS != 0 {
			t.Errorf("DurationMS should be 0 when StartedAt is zero, got %d", report.DurationMS)
		}
	})
}
