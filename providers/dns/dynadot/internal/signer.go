package internal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"path"
	"strings"
)

// https://www.dynadot.com/domain/api-document?api-version=2.0.0#request-header
func (c *Client) sign(req *http.Request, xRequestID, body string) string {
	fpq := path.Clean("/" + req.URL.EscapedPath())

	if req.URL.RawQuery != "" {
		fpq += "?" + req.URL.RawQuery
	}

	elements := []string{c.apiKey, fpq, xRequestID, body}

	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(strings.Join(elements, "\n")))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
