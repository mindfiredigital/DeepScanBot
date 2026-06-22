package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type PageStorage struct {
	visitedUrls map[string]bool
	pageContent map[string][]byte
	urlSource   map[string]string
	jsonOutput  bool
	mutex       sync.Mutex
	urls        []string
	maxSize     int
	file        *os.File
	writer      *bufio.Writer
}

// URLEntry is one crawled URL and, when known, the HTML element that referenced it.
type URLEntry struct {
	URL    string `json:"url"`
	Source string `json:"source,omitempty"`
}

func NewPageStorage(jsonOutput bool, maxSize int) *PageStorage {
	ps := &PageStorage{
		visitedUrls: make(map[string]bool),
		pageContent: make(map[string][]byte),
		urlSource:   make(map[string]string),
		jsonOutput:  jsonOutput,
		maxSize:     maxSize,
		urls:        []string{},
	}
	if !jsonOutput {
		f, err := os.Create("crawler_results.txt")
		if err != nil {
			log.Println("Error creating crawler_results.txt:", err)
		} else {
			ps.file = f
			ps.writer = bufio.NewWriter(f)
		}
	}
	return ps
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

func (ps *PageStorage) GetContent(url string) ([]byte, bool) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	content, exists := ps.pageContent[url]
	return content, exists
}

func (ps *PageStorage) IsJSONOutput() bool {
	return ps.jsonOutput
}

func (ps *PageStorage) StoreContent(url string, content []byte, showSource bool) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.pageContent[url] = content

	if ps.jsonOutput {
		ps.urls = append(ps.urls, url)
	} else {
		source := ps.urlSource[url]
		if ps.writer != nil {
			var err error
			if showSource && source != "" {
				_, err = ps.writer.WriteString("[" + source + "] " + url + "\n")
			} else {
				_, err = ps.writer.WriteString(url + "\n")
			}
			if err != nil {
				log.Println("Error writing URL to file:", err)
			}
		}
	}
}

func (ps *PageStorage) Close() error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	var errs []error
	if ps.writer != nil {
		if err := ps.writer.Flush(); err != nil {
			errs = append(errs, err)
		}
		ps.writer = nil
	}
	if ps.file != nil {
		if err := ps.file.Close(); err != nil {
			errs = append(errs, err)
		}
		ps.file = nil
	}

	return errors.Join(errs...)
}

func (ps *PageStorage) WriteJSONToFile(filename string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	entries := make([]URLEntry, 0, len(ps.urls))
	for _, url := range ps.urls {
		entry := URLEntry{URL: url}
		entry.Source = ps.urlSource[url]
		entries = append(entries, entry)
	}

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
