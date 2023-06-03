package netspy

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func includes(array []string, item string) bool {
	for _, array_item := range array {
		if strings.Contains(array_item, item) {
			return true
		}
	}
	return false
}

func baseURLof(rawUrl string) (string, error) {
	parsedURL, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	return parsedURL.Scheme + "://" + parsedURL.Host, nil
}

func normalizeUrl(rawURL string) (string, error) {
	parsedURL, err := url.Parse(strings.ToLower(rawURL))
	if err != nil {
		return "", err
	}

	hostPath := "https://" + parsedURL.Host + parsedURL.Path
	last := len(hostPath) - 1
	if last != -1 && hostPath[last] == '/' {
		return hostPath[:last], nil
	}

	return hostPath, nil
}

func searchForLinks(node *html.Node, baseURL string) []string {
	var links []string

	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key != "href" {
				continue
			}

			link := attr.Val
			var err error
			if len(link) > 1 && link[0] == '/' {
				base, berr := baseURLof(baseURL)
				if berr != nil {
					continue
				}
				link, err = url.JoinPath(base, link)
			} else {
				_, err = url.ParseRequestURI(link)
				if err != nil {
					link, err = url.JoinPath(baseURL, link)
				}
			}
			if err != nil {
				continue
			}

			link, err = normalizeUrl(link)
			if err != nil {
				continue
			}

			links = append(links, link)
			break
		}
	}

	for childNode := node.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		links = append(links, searchForLinks(childNode, baseURL)...)
	}

	return links
}
