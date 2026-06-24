package crawler

import (
	"net/http"
	"time"

	"web-crawler-assignment/fetcher"
)

// isRetryable checks if a status code indicates a retryable error.
func isRetryable(statusCode int) bool {
	return statusCode == 0 || statusCode == http.StatusRequestTimeout || statusCode == http.StatusTooManyRequests || statusCode >= 500
}

// retryDelay calculates the delay before the next retry attempt.
func (c *Crawler) retryDelay(attempt int, statusCode int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}

	delay := time.Duration(attempt) * c.retryBackoff
	if statusCode == http.StatusTooManyRequests {
		delay *= 3
	}

	if delay > 30*time.Second {
		delay = 30 * time.Second
	}

	return delay
}

// fetchWithRetry fetches a URL with retries, using exponential backoff and Retry-After headers.
func (c *Crawler) fetchWithRetry(targetURL string) ([]byte, int, string, int, int, error) {
	maxAttempts := c.retries + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		c.waitForHostDelay(targetURL)

		result := fetcher.FetchWithDetails(targetURL, c.timeout, c.proxyURL, c.disableRedirects, c.insecure, c.maxSize, c.contentTypes)

		if result.Err == nil {
			return result.Body, result.Size, result.ContentType, result.StatusCode, attempt, nil
		}

		if attempt == maxAttempts || !isRetryable(result.StatusCode) {
			return result.Body, result.Size, result.ContentType, result.StatusCode, attempt, result.Err
		}

		delay := c.retryDelay(attempt, result.StatusCode, result.RetryAfter)
		c.log.Infof("Retrying URL %s after %s because status=%d error=%v", targetURL, delay, result.StatusCode, result.Err)
		time.Sleep(delay)
	}

	return nil, 0, "", 0, maxAttempts, nil
}
