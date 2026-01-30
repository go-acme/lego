package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
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

// ParseRetryAfter parses the Retry-After header value according to RFC 7231.
// The header can be either delay-seconds (numeric) or HTTP-date (RFC 1123 format).
// Returns the duration until the retry time, or an error if parsing fails.
// If the HTTP-date is in the past, returns 0 duration.
func ParseRetryAfter(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}

	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	if retryTime, err := time.Parse(time.RFC1123, value); err == nil {
		duration := time.Until(retryTime)
		if duration < 0 {
			return 0, nil
		}

		return duration, nil
	}

	return 0, fmt.Errorf("invalid Retry-After value: %q (expected delay-seconds or HTTP-date)", value)
}
