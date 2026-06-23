package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
	"web-crawler-assignment/crawler"
	"web-crawler-assignment/storage"
)

func main() {
	url := flag.String("url", "", "The starting URL")
	depth := flag.Int("depth", 2, "The maximum depth to crawl")
	timeout := flag.Int("timeout", 2, "The timeout for each request in seconds")
	proxy := flag.String("proxy", "", "Proxy URL. E.g. -proxy http://127.0.0.1:8080")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	maxSize := flag.Int("size", -1, "Page size limit in KB. Default is -1 (no limit)")
	disableRedirects := flag.Bool("dr", false, "Disable following redirects")
	showSource := flag.Bool("s", false, "Show the source of the URL based on where it was found")
	insecure := flag.Bool("insecure", false, "Disable TLS verification")
	uniqueUrls := flag.Bool("u", false, "Ensure unique URLs")
	concurrency := flag.Int("concurrency", 0, "Maximum concurrent requests; 0 uses available CPU capacity")
	hostConcurrency := flag.Int("host-concurrency", 0, "Maximum concurrent requests per host; 0 uses -concurrency")
	contentTypes := flag.String("content-types", "text/html", "Comma-separated MIME types to download, e.g. text/html,application/pdf,image/jpeg")
	output := flag.String("output", "crawler_results", "Output filename without an extension")
	ignoreRobots := flag.Bool("ignore-robots", false, "Ignore robots.txt crawl restrictions")
	crossDomain := flag.Bool("cross-domain", false, "Follow links to hosts other than the starting URL")
	retries := flag.Int("retries", 0, "Number of retry attempts for transient fetch failures")
	retryBackoff := flag.Duration("retry-backoff", time.Second, "Base retry backoff duration, e.g. 500ms, 2s")
	crawlDelay := flag.Duration("delay", 0, "Politeness delay between requests to the same host, e.g. 500ms")
	includeSitemap := flag.Bool("sitemap", false, "Discover and crawl URLs from the starting host's /sitemap.xml")
	resume := flag.Bool("resume", false, "Load existing output file and avoid recrawling URLs already present")
	showHelp := flag.Bool("h", false, "Show this help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s web crawler:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDeepScanBot is a feature-rich web crawler for scanning websites, extracting links,\n")
		fmt.Fprintf(os.Stderr, "and generating comprehensive reports. It supports retries, rate-limiting,\n")
		fmt.Fprintf(os.Stderr, "sitemap discovery, resume mode, robots.txt compliance, and more.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Basic crawl\n")
		fmt.Fprintf(os.Stderr, "  deepscanbot -url https://example.com -depth 2\n\n")
		fmt.Fprintf(os.Stderr, "  # Crawl with JSON output, 3 retries, and 500ms delay\n")
		fmt.Fprintf(os.Stderr, "  deepscanbot -url https://example.com -json -retries 3 -retry-backoff 500ms -delay 1s\n\n")
		fmt.Fprintf(os.Stderr, "  # Cross-domain crawl with sitemap discovery\n")
		fmt.Fprintf(os.Stderr, "  deepscanbot -url https://example.com -cross-domain -sitemap -concurrency 10 -host-concurrency 2\n\n")
		fmt.Fprintf(os.Stderr, "  # Resume an interrupted crawl\n")
		fmt.Fprintf(os.Stderr, "  deepscanbot -url https://example.com -resume -output my_results.json -json\n\n")
		fmt.Fprintf(os.Stderr, "  # Advanced: Goodreads crawl with rate-limit handling\n")
		fmt.Fprintf(os.Stderr, "  deepscanbot -url https://www.goodreads.com -depth 2 -delay 2s -retries 5 -retry-backoff 2s -concurrency 2 -host-concurrency 1\n")
	}

	flag.Parse()

	if *showHelp {
		flag.Usage()
		return
	}

	startURL, err := validateStartURL(*url)
	if err != nil {
		log.Fatal(err)
	}

	timeoutDuration := time.Duration(*timeout) * time.Second

	outputFilename, err := buildOutputFilename(*output, *jsonOutput)
	if err != nil {
		log.Fatal(err)
	}
	var resumeEntries []storage.URLEntry
	if *resume {
		resumeEntries, err = storage.ReadEntriesFromFile(outputFilename)
		if err != nil {
			log.Fatalf("load resume file: %v", err)
		}
		log.Printf("Resume mode loaded %d existing results from %s", len(resumeEntries), outputFilename)
	}

	c := crawler.NewCrawlerWithOptions(startURL, *depth, timeoutDuration, *proxy, *maxSize, *disableRedirects, *insecure, *uniqueUrls, *concurrency, parseContentTypes(*contentTypes), *ignoreRobots, *crossDomain, crawler.Options{
		Retries:            *retries,
		RetryBackoff:       *retryBackoff,
		CrawlDelay:         *crawlDelay,
		PerHostConcurrency: *hostConcurrency,
		IncludeSitemap:     *includeSitemap,
		ResumeEntries:      resumeEntries,
	})

	report, err := c.StartReport()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	allEntries := append(append([]storage.URLEntry{}, report.URLs...), report.Skipped...)
	if *jsonOutput {
		err = storage.WriteJSONReportToFile(outputFilename, report)
	} else {
		err = storage.WriteTextToFile(outputFilename, allEntries, *showSource)
	}
	if err != nil {
		log.Fatalf("write results: %v", err)
	}
	log.Printf("Results written to %s", outputFilename)

	// Clearer result summary in CLI output
	printResultSummary(report)
}

func printResultSummary(report storage.CrawlReport) {
	sep := strings.Repeat("=", 60)
	fmt.Fprintf(os.Stderr, "\n%s\n", sep)
	fmt.Fprintf(os.Stderr, "  CRAWL RESULT SUMMARY\n")
	fmt.Fprintf(os.Stderr, "%s\n", sep)
	fmt.Fprintf(os.Stderr, "  Target URL:     %s\n", report.StartURL)
	fmt.Fprintf(os.Stderr, "  Output file:    %s\n", report.OutputFile)
	fmt.Fprintf(os.Stderr, "  Duration:       %d ms\n", report.DurationMS)
	fmt.Fprintf(os.Stderr, "%s\n", sep)
	fmt.Fprintf(os.Stderr, "  Total URLs:     %d\n", report.Summary.Total)
	fmt.Fprintf(os.Stderr, "  ── Passed:      %d\n", report.Summary.Passed)
	fmt.Fprintf(os.Stderr, "  ── Failed:      %d\n", report.Summary.Failed)
	fmt.Fprintf(os.Stderr, "  ── Discovered:  %d\n", report.Summary.Discovered)
	fmt.Fprintf(os.Stderr, "  ── Skipped:     %d\n", report.Summary.Skipped)
	fmt.Fprintf(os.Stderr, "%s\n", sep)
	if report.Summary.Skipped > 0 {
		fmt.Fprintf(os.Stderr, "  Skip Breakdown:\n")
		if report.Summary.SkippedByRobots > 0 {
			fmt.Fprintf(os.Stderr, "    • Robots.txt:    %d\n", report.Summary.SkippedByRobots)
		}
		if report.Summary.SkippedByDomain > 0 {
			fmt.Fprintf(os.Stderr, "    • Domain scope:  %d\n", report.Summary.SkippedByDomain)
		}
		if report.Summary.SkippedByDuplicate > 0 {
			fmt.Fprintf(os.Stderr, "    • Duplicate:     %d\n", report.Summary.SkippedByDuplicate)
		}
		if report.Summary.SkippedByContent > 0 {
			fmt.Fprintf(os.Stderr, "    • Content-type:  %d\n", report.Summary.SkippedByContent)
		}
		if report.Summary.SkippedByDepth > 0 {
			fmt.Fprintf(os.Stderr, "    • Max depth:     %d\n", report.Summary.SkippedByDepth)
		}
		if report.Summary.SkippedByOther > 0 {
			fmt.Fprintf(os.Stderr, "    • Other:         %d\n", report.Summary.SkippedByOther)
		}
		fmt.Fprintf(os.Stderr, "%s\n", sep)
	}
	if report.Summary.RetriedRequests > 0 {
		fmt.Fprintf(os.Stderr, "  Retries:        %d requests were retried\n", report.Summary.RetriedRequests)
	}
	fmt.Fprintf(os.Stderr, "  Max Depth:      %d\n", report.Summary.MaxDepth)
	if len(report.Skipped) > 0 {
		fmt.Fprintf(os.Stderr, "  Skipped URLs:   %d (separate list in output)\n", len(report.Skipped))
	}
	fmt.Fprintf(os.Stderr, "%s\n\n", sep)
}

func buildOutputFilename(baseName string, jsonOutput bool) (string, error) {
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		return "", fmt.Errorf("output filename must not be empty")
	}
	if jsonOutput {
		return baseName + ".json", nil
	}
	return baseName + ".txt", nil
}

func validateStartURL(rawURL string) (string, error) {
	startURL := strings.TrimSpace(rawURL)
	if startURL == "" {
		return "", fmt.Errorf("you must specify a starting URL with the -url flag")
	}

	parsedURL, err := url.ParseRequestURI(startURL)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", fmt.Errorf("invalid URL %q: must be an absolute http:// or https:// URL", rawURL)
	}

	return parsedURL.String(), nil
}

func parseContentTypes(value string) []string {
	var contentTypes []string
	for _, contentType := range strings.Split(value, ",") {
		if contentType = strings.TrimSpace(contentType); contentType != "" {
			contentTypes = append(contentTypes, contentType)
		}
	}
	return contentTypes
}