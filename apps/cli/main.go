package main

import (
	"flag"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mindfiredigital/DeepScanBot/packages/crawler"
	"github.com/mindfiredigital/DeepScanBot/packages/logger"
	"github.com/mindfiredigital/DeepScanBot/packages/storage"
)

var log = logger.New("info")

func main() {
	showHelp := flag.Bool("h", false, "Show this help message")
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
	contentTypes := flag.String("content-types", "text/html", "MIME types to download. Must be quoted as one argument, e.g. -content-types \"text/html application/pdf image/jpeg\" or -content-types \"text/html,application/pdf,image/jpeg\"")
	output := flag.String("output", "crawler_results", "Output filename without an extension")
	ignoreRobots := flag.Bool("ignore-robots", false, "Ignore robots.txt crawl restrictions")
	crossDomain := flag.Bool("cross-domain", false, "Follow links to hosts other than the starting URL")
	retries := flag.Int("retries", 0, "Number of retry attempts for transient fetch failures")
	retryBackoff := flag.Duration("retry-backoff", time.Second, "Base retry backoff duration, e.g. 500ms, 2s")
	crawlDelay := flag.Duration("delay", 0, "Politeness delay between requests to the same host, e.g. 500ms")
	includeSitemap := flag.Bool("sitemap", false, "Discover and crawl URLs from the starting host's /sitemap.xml")
	resume := flag.Bool("resume", false, "Load existing output file and avoid recrawling URLs already present")

	flag.Parse()

	if *showHelp {
		flag.Usage()
		return
	}

	startURL, err := validateStartURL(*url)
	if err != nil {
		log.Fatalf(err.Error())
	}

	timeoutDuration := time.Duration(*timeout) * time.Second

	outputFilename, err := buildOutputFilename(*output, *jsonOutput)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var resumeEntries []storage.URLEntry
	if *resume {
		resumeEntries, err = storage.ReadEntriesFromFile(outputFilename)
		if err != nil {
			log.Fatalf("load resume file: %v", err)
		}

		log.Infof("Resume mode loaded %d existing results from %s", len(resumeEntries), outputFilename)
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

	if *jsonOutput {
		err = storage.WriteJSONReportToFile(outputFilename, report)
	} else {
		err = storage.WriteTextToFile(outputFilename, report.URLs, *showSource)
	}

	if err != nil {
		log.Fatalf("write results: %v", err)
	}

	log.Infof("Results written to %s", outputFilename)
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

	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		if part = strings.TrimSpace(part); part != "" {
			contentTypes = append(contentTypes, part)
		}
	}

	return contentTypes
}
