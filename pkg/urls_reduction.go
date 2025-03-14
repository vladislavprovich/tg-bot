package pkg

import (
	"net/url"
	"strings"
)

func ShortInfo(shortURL string) string {
	extractLastSegment := func(link string) string {
		parsedURL, err := url.Parse(link)
		if err != nil {
			return link
		}
		pathSegments := strings.Split(parsedURL.Path, "/")
		if len(pathSegments) > 0 {
			return pathSegments[len(pathSegments)-1]
		}
		return ""
	}
	nameShort := extractLastSegment(shortURL)
	return nameShort
}

func OriginalInfo(originalURL string) string {
	extractDomain := func(link string) string {
		parsedURL, err := url.Parse(link)
		if err != nil {
			return link
		}
		return parsedURL.Hostname()
	}
	nameOrig := extractDomain(originalURL)
	return nameOrig
}
