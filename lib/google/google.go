package google

import (
	"bytes"
	"net/http"
	"net/url"
	"research-toolkit/lib/getter"
	"strconv"

	"golang.org/x/net/html"
)

func GetAttr(node *html.Node, key string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

type Client struct {
	getter getter.Getter
}

func NewClient(getter getter.Getter) Client {
	return Client{getter: getter}
}

func (c Client) Search(query string, pageCount int) ([]*url.URL, error) {
	var results []*url.URL

	for p := 0; p < pageCount; p++ {
		offset := 10 * p

		var queryValues = make(url.Values)
		queryValues.Add("q", query)
		if offset > 0 {
			queryValues.Add("start", strconv.Itoa(offset))
		}
		u := &url.URL{
			Scheme:   "https",
			Host:     "www.google.com",
			Path:     "search",
			RawQuery: queryValues.Encode(),
		}

		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:109.0) Gecko/20100101 Firefox/119.0")

		res, err := c.getter.Do(getter.Request{
			URL: u,
		})
		if err != nil {
			return nil, err
		}

		reader := bytes.NewBuffer(res)
		node, err := html.Parse(reader)
		if err != nil {
			return nil, err
		}

		parsed := parseOldPage(node)
		if len(parsed) == 0 {
			parsed = parseNewPage(node)
		}
		results = append(results, parsed...)
	}

	return results, nil
}
