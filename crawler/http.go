package crawler

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

// buildTransport creates an http.Transport with proxy and TLS settings, or nil if none are needed.
func (c *Crawler) buildTransport() *http.Transport {
	if c.proxyUrl == "" && !c.insecure {
		return nil
	}

	transport := &http.Transport{}

	if c.proxyUrl != "" {
		proxy, err := url.Parse(c.proxyUrl)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	if c.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return transport
}

// httpClient builds an http.Client with proxy, TLS, and redirect settings.
func (c *Crawler) httpClient() *http.Client {
	client := &http.Client{Timeout: c.timeout}
	if c.disableRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	if transport := c.buildTransport(); transport != nil {
		client.Transport = transport
	}

	return client
}
