package google

import (
	"net/url"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func parseNewPage(node *html.Node) []*url.URL {
	resultNodes := htmlquery.Find(node, "//span[@jsaction]//a[@href]")
	var results []*url.URL
	for _, result := range resultNodes {
		link, ok := GetAttr(result, "href")
		if !ok {
			continue
		}

		resultUrl, err := url.Parse(link)
		if err != nil {
			continue
		}
		results = append(results, resultUrl)
	}
	return results
}

func parseOldPage(node *html.Node) []*url.URL {
	resultNodes := htmlquery.Find(node, "//a//h3")
	var results []*url.URL
	for _, result := range resultNodes {
		current := result
		for current != nil {
			current = current.Parent
			if current.Type == html.ElementNode && current.Data == "a" {
				break
			}
		}
		if current == nil {
			continue
		}

		link, ok := GetAttr(current, "href")
		if !ok {
			continue
		}

		wrappedUrl, err := url.Parse(link)
		if err != nil {
			continue
		}

		resultUrl, err := url.Parse(wrappedUrl.Query().Get("q"))
		if err != nil {
			continue
		}
		results = append(results, resultUrl)
	}
	return results
}
