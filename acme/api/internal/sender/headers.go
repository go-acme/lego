package sender

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/log"
)

// GetLink get a rel into the `Link` header.
func GetLink(header http.Header, rel string) string {
	links := GetLinks(header, rel)
	if len(links) < 1 {
		return ""
	}

	return links[0]
}

// GetLinks get all rels into the `Link` header.
func GetLinks(header http.Header, rel string) []string {
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

// GetLocation get the value of the header `Location`.
func GetLocation(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	return resp.Header.Get("Location")
}

// GetRetryAfter get the value of the header `Retry-After`.
func GetRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}

	retryAfter, err := parseRetryAfter(resp.Header.Get("Retry-After"))
	if err != nil {
		log.Warn("Failed to parse Retry-After header.", log.ErrorAttr(err))
	}

	return retryAfter
}

// parseRetryAfter parses the Retry-After header value according to RFC 7231.
// The header can be either delay-seconds (numeric) or HTTP-date (RFC 1123 format).
// https://datatracker.ietf.org/doc/html/rfc7231#section-7.1.3
// Returns the duration until the retry time.
func parseRetryAfter(value string) (time.Duration, error) {
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

	return 0, fmt.Errorf("invalid Retry-After value: %q", value)
}
