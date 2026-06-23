package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// WriteJSONToFile writes URL entries as a JSON report file.
func WriteJSONToFile(filename string, entries []URLEntry) error {
	return WriteJSONReportToFile(filename, NewCrawlReport("", "", zeroTime, zeroTime, entries))
}

// WriteJSONReportToFile writes a CrawlReport as a JSON file.
func WriteJSONReportToFile(filename string, report CrawlReport) error {
	report.OutputFile = filename

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0o644)
}

// ReadEntriesFromFile reads URL entries from a JSON or text file.
func ReadEntriesFromFile(filename string) ([]URLEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	// Try as new-style CrawlReport
	var report CrawlReport
	if err := json.Unmarshal(data, &report); err == nil && (len(report.URLs) > 0 || len(report.Skipped) > 0) {
		return append(report.URLs, report.Skipped...), nil
	}

	// Try as legacy JSON format
	var legacy struct {
		URLs []URLEntry `json:"urls"`
	}

	if err := json.Unmarshal(data, &legacy); err == nil {
		return legacy.URLs, nil
	}

	// Fallback to text format
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

// WriteTextToFile writes URL entries as a plain text file.
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

// zeroTime is a zero time.Time value for passing to NewCrawlReport when no timing is needed.
var zeroTime time.Time
