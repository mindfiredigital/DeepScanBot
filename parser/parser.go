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
		var targetAttr string
		var sourceType string

		switch n.Data {
		case "a":
			targetAttr = "href"
			sourceType = "href"
		case "script":
			targetAttr = "src"
			sourceType = "script"
		case "img":
			targetAttr = "src"
			sourceType = "img"
		case "link":
			targetAttr = "href"
			sourceType = "link"
		case "iframe":
			targetAttr = "src"
			sourceType = "iframe"
		case "form":
			targetAttr = "action"
			sourceType = "form"
		}

		if targetAttr != "" {
			for _, attr := range n.Attr {
				if attr.Key == targetAttr {
					val := strings.TrimSpace(attr.Val)
					if val == "" || strings.HasPrefix(val, "#") {
						continue
					}
					if u, err := url.Parse(val); err == nil {
						resolved := base.ResolveReference(u)
						if resolved.Scheme == "http" || resolved.Scheme == "https" {
							links[resolved.String()] = sourceType
						}
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

