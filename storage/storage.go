package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"sync"
)

type PageStorage struct {
	visitedUrls map[string]bool
	urlSource   map[string]string
	mutex       sync.Mutex
	results     []URLEntry
}

// URLEntry is one crawled URL and, when known, the HTML element that referenced it.
type URLEntry struct {
	URL        string `json:"url"`
	Source     string `json:"source,omitempty"`
	Depth      int    `json:"depth"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

func NewPageStorage() *PageStorage {
	return &PageStorage{
		visitedUrls: make(map[string]bool),
		urlSource:   make(map[string]string),
		results:     []URLEntry{},
	}
}

func (ps *PageStorage) MarkVisited(url string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.visitedUrls[url] = true
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
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.results = append(ps.results, URLEntry{
		URL:        url,
		Source:     ps.urlSource[url],
		Depth:      depth,
		StatusCode: statusCode,
		Error:      resultError,
	})
}

// Results returns a snapshot of the URLs collected during a crawl.
func (ps *PageStorage) Results() []URLEntry {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	return append([]URLEntry(nil), ps.results...)
}

func WriteJSONToFile(filename string, entries []URLEntry) error {
	data := struct {
		URLs []URLEntry `json:"urls"`
	}{
		URLs: entries,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
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
		if entry.Error != "" {
			line += " [error=" + entry.Error + "]"
		}
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}
