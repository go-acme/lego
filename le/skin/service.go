package skin

import (
	"net/http"
	"regexp"
)

var linkExpr = regexp.MustCompile(`<(.+?)>;\s*rel="(.+?)"`)

type service struct {
	core *Core
}

// getLink get a rel into the Link header
func getLink(header http.Header, rel string) string {
	for _, link := range header["Link"] {
		for _, m := range linkExpr.FindAllStringSubmatch(link, -1) {
			if len(m) != 3 {
				continue
			}
			if m[2] == rel {
				return m[1]
			}
		}
	}
	return ""
}

// getLocation get the value of the header Location
func getLocation(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	return resp.Header.Get("Location")
}

// getRetryAfter get the value of the header Retry-After
func getRetryAfter(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	return resp.Header.Get("Retry-After")
}
