package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"web-crawler-assignment/crawler"
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

	if *url == "" {
		log.Fatal("You must specify a starting URL with the -url flag")
	}

	timeoutDuration := time.Duration(*timeout) * time.Second

	c := crawler.NewCrawler(*url, *depth, timeoutDuration, *proxy, *jsonOutput, *maxSize, *disableRedirects, *showSource, *insecure, *uniqueUrls, *concurrency, parseContentTypes(*contentTypes))

	if err := c.Start(); err != nil {
		log.Fatalf("error: %v", err)
	}
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
