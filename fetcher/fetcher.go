package fetcher

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func Fetch(targetUrl string, timeout time.Duration, proxyUrl string, disableRedirects bool, insecure bool) ([]byte, int, error) {
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
			return nil, 0, err
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
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "DeepScanBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, 0, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, len(body), nil
}
