package api

import (
	"net/http"
	"regexp"
)

type service struct {
	core *Core
}

// getLink get a rel into the Link header.
func getLink(header http.Header, rel string) string {
	links := getLinks(header, rel)
	if len(links) < 1 {
		return ""
	}

	return links[0]
}

func getLinks(header http.Header, rel string) []string {
	linkExpr := regexp.MustCompile(`<(.+?)>(?:;[^;]+)*?;\s*rel="(.+?)"`)

	var links []string
	for _, link := range header["Link"] {
		for _, m := range linkExpr.FindAllStringSubmatch(link, -1) {
			if len(m) != 3 {
				continue
			}
			if m[2] == rel {
				links = append(links, m[1])
			}
		}
	}

	return links
}

// getLocation get the value of the header Location.
func getLocation(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	return resp.Header.Get("Location")
}

// getRetryAfter get the value of the header Retry-After.
func getRetryAfter(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	return resp.Header.Get("Retry-After")
}
