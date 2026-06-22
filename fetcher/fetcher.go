package fetcher

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Fetch(targetUrl string, timeout time.Duration, proxyUrl string, disableRedirects bool, insecure bool, maxSize int) ([]byte, int, string, error) {
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

	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err != nil {
			return nil, 0, "", err
		}
		transport.Proxy = http.ProxyURL(proxy)
		hasCustomTransport = true
	}

	if insecure {
		log.Println("-insecure flag, disable TLS verification")
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		hasCustomTransport = true
	}

	if hasCustomTransport {
		client.Transport = transport
	}

	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("User-Agent", "DeepScanBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, 0, "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		// Avoid downloading the body for non-HTML resources
		contentLength := resp.ContentLength
		size := int(contentLength)
		if size < 0 {
			size = 0
		}
		return nil, size, contentType, nil
	}

	var bodyReader io.Reader = resp.Body
	if maxSize > 0 {
		bodyReader = io.LimitReader(resp.Body, int64(maxSize)*1024+1)
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, 0, contentType, err
	}

	if maxSize > 0 && len(body) > maxSize*1024 {
		return nil, len(body), contentType, fmt.Errorf("page exceeds size limit (%d bytes)", len(body))
	}

	return body, len(body), contentType, nil
}
