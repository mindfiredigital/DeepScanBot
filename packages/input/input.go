package input

import (
	"bufio"
	"os"
)

// ReadInput reads URLs from stdin or a file based on the provided flags.
// Returns a slice of URLs and an error if reading fails.
// Precedence: file input takes priority over stdin if both are provided.
func ReadInput(filePath string, useStdin bool) ([]string, error) {
	// File input takes precedence
	if filePath != "" {
		return readFromFile(filePath)
	}

	// Fall back to stdin if requested
	if useStdin {
		return readFromStdin()
	}

	return nil, nil
}

// readFromFile reads URLs from a file (one per line)
func readFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// readFromStdin reads URLs from standard input (one per line)
func readFromStdin() ([]string, error) {
	var urls []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// HasStdinData checks if there is data available on stdin
func HasStdinData() bool {
	// Check if stdin is not a terminal (i.e., it's piped)
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// If stdin is a pipe or redirect, it has data
	return (fileInfo.Mode() & os.ModeCharDevice) == 0
}
