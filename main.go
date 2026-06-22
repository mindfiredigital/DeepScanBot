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
	concurrency := flag.Int("concurrency", 10, "The maximum number of concurrent requests")
	contentTypes := flag.String("content-types", "text/html", "Comma-separated MIME types to download, e.g. text/html,application/pdf,image/jpeg")
	showHelp := flag.Bool("h", false, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s web crawler:\n", os.Args[0])
		flag.PrintDefaults()
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

	c := crawler.NewCrawler(startURL, *depth, timeoutDuration, *proxy, *maxSize, *disableRedirects, *insecure, *uniqueUrls, *concurrency, parseContentTypes(*contentTypes))

	results, err := c.Start()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if *jsonOutput {
		err = storage.WriteJSONToFile("crawler_results.json", results)
	} else {
		err = storage.WriteTextToFile("crawler_results.txt", results, *showSource)
	}
	if err != nil {
		log.Fatalf("write results: %v", err)
	}
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
