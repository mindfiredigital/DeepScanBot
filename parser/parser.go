package parser

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func Parse(body []byte, baseURLStr string) map[string]string {
	links := make(map[string]string)
	
	base, err := url.Parse(baseURLStr)
	if err != nil {
		return links
	}

	htmlData, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return links
	}

	links = extractLinks(htmlData, base, links)
	return links
}

func extractLinks(n *html.Node, base *url.URL, links map[string]string) map[string]string {
	if n.Type == html.ElementNode {
		if n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if u, err := url.Parse(attr.Val); err == nil {
						resolved := base.ResolveReference(u)
						links[resolved.String()] = "href"
					}
				}
			}
		}
		if n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					if u, err := url.Parse(attr.Val); err == nil {
						resolved := base.ResolveReference(u)
						links[resolved.String()] = "script"
					}
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = extractLinks(c, base, links)
	}
	return links
}

