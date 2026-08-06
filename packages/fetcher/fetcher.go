package fetcher

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mindfiredigital/DeepScanBot/packages/logger"
	"github.com/mindfiredigital/DeepScanBot/packages/types"
)

// FetchResult alias for backward compatibility.
type FetchResult = types.FetchResult

// Fetch performs an HTTP GET request and returns the response body, size, content type, and status code.
func Fetch(targetURL string, timeout time.Duration, proxyURL string, disableRedirects bool, insecure bool, maxSize int, allowedContentTypes []string) ([]byte, int, string, int, error) {
	result := FetchWithDetails(targetURL, timeout, proxyURL, disableRedirects, insecure, maxSize, allowedContentTypes)
	return result.Body, result.Size, result.ContentType, result.StatusCode, result.Err
}

// FetchWithDetails performs an HTTP GET request and returns a FetchResult with detailed information
// including the Retry-After duration from the response headers.
func FetchWithDetails(targetURL string, timeout time.Duration, proxyURL string, disableRedirects bool, insecure bool, maxSize int, allowedContentTypes []string) FetchResult {
	client := &http.Client{
		Timeout: timeout,
	}

	if disableRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	transport := &http.Transport{}
	hasCustomTransport := false

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return FetchResult{Err: err}
		}

		transport.Proxy = http.ProxyURL(proxy)
		hasCustomTransport = true
	}

	if insecure {
		fetcherLog := logger.New("info")
		fetcherLog.Infof("-insecure flag, disable TLS verification")

		//nolint:gosec // InsecureSkipVerify is intentional for the -insecure flag
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		hasCustomTransport = true
	}

	if hasCustomTransport {
		client.Transport = transport
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", targetURL, nil)
	if err != nil {
		return FetchResult{Err: err}
	}

	req.Header.Set("User-Agent", "DeepScanBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		wrappedErr := wrapProxyError(err, proxyURL)
		return FetchResult{StatusCode: 0, Err: wrappedErr}
	}
	defer resp.Body.Close()

	result := FetchResult{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		RetryAfter:  parseRetryAfter(resp.Header.Get("Retry-After")),
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		result.Err = fmt.Errorf("bad status code: %d", resp.StatusCode)
		return result
	}

	if !isAllowedContentType(result.ContentType, allowedContentTypes) {
		contentLength := resp.ContentLength

		result.Size = int(contentLength)
		if result.Size < 0 {
			result.Size = 0
		}

		result.Body = nil

		return result
	}

	var bodyReader io.Reader = resp.Body
	if maxSize > 0 {
		bodyReader = io.LimitReader(resp.Body, int64(maxSize)*1024+1)
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		result.Err = err
		return result
	}

	result.Size = len(body)
	if maxSize > 0 && len(body) > maxSize*1024 {
		result.Err = fmt.Errorf("page exceeds size limit (%d bytes)", len(body))
		return result
	}

	result.Body = body

	return result
}

// parseRetryAfter parses the Retry-After HTTP header and returns the duration to wait.
// It supports both seconds (integer) and HTTP-date formats.
// wrapProxyError wraps low-level proxy errors with user-friendly messages
func wrapProxyError(err error, proxyURL string) error {
	if proxyURL == "" {
		return err
	}

	errStr := err.Error()

	// Check for specific error patterns and provide user-friendly messages
	if strings.Contains(errStr, "connection refused") {
		return fmt.Errorf("failed to connect to proxy at %s: connection refused\nHint: Verify the proxy server is running and accessible", proxyURL)
	}

	if strings.Contains(errStr, "dial tcp") {
		return fmt.Errorf("failed to establish connection to proxy at %s\nHint: Check proxy address and network connectivity", proxyURL)
	}

	if strings.Contains(errStr, "proxyconnect") {
		return fmt.Errorf("proxy connection failed for %s\nHint: Verify proxy configuration and credentials", proxyURL)
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "i/o timeout") {
		return fmt.Errorf("connection to proxy at %s timed out\nHint: Check network connectivity and proxy performance", proxyURL)
	}

	if strings.Contains(errStr, "unsupported protocol scheme") {
		return fmt.Errorf("invalid proxy protocol in %s\nHint: Use http:// or https:// for proxy URLs", proxyURL)
	}

	// Generic proxy error
	return fmt.Errorf("proxy error: %s\nHint: Check proxy configuration", errStr)
}

func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return 0
	}

	val = strings.TrimSpace(val)

	// Try parsing as seconds (integer)
	if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP-date (e.g., Wed, 21 Oct 2015 07:28:00 GMT)
	if t, err := time.Parse(time.RFC1123, val); err == nil {
		wait := time.Until(t)
		if wait > 0 {
			return wait
		}

		return 0
	}

	// Try other common date formats
	if t, err := time.Parse(http.TimeFormat, val); err == nil {
		wait := time.Until(t)
		if wait > 0 {
			return wait
		}

		return 0
	}

	return 0
}

func isAllowedContentType(contentType string, allowedContentTypes []string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	}

	if len(allowedContentTypes) == 0 {
		return strings.EqualFold(mediaType, "text/html")
	}

	for _, allowed := range allowedContentTypes {
		allowed = strings.ToLower(strings.TrimSpace(allowed))
		if allowed == "" {
			continue
		}

		if allowed == "*/*" || strings.EqualFold(mediaType, allowed) {
			return true
		}

		if strings.HasSuffix(allowed, "/*") && strings.HasPrefix(strings.ToLower(mediaType), strings.TrimSuffix(allowed, "*")) {
			return true
		}
	}

	return false
}
