// Package dmapi Client for DMAPI joker.com.
// https://joker.com/faq/category/39/22-dmapi.html
package dmapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://dmapi.joker.com/request/"

// Response Joker DMAPI Response.
type Response struct {
	Headers    url.Values
	Body       string
	StatusCode int
	StatusText string
	AuthSid    string
}

type AuthInfo struct {
	APIKey   string
	Username string
	Password string
}

// Client a DMAPI Client.
type Client struct {
	apiKey   string
	username string
	password string

	token   *Token
	muToken sync.Mutex

	Debug      bool
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new DMAPI Client.
func NewClient(authInfo AuthInfo) *Client {
	return &Client{
		apiKey:     authInfo.APIKey,
		username:   authInfo.Username,
		password:   authInfo.Password,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetZone returns content of DNS zone for domain.
func (c *Client) GetZone(ctx context.Context, domain string) (*Response, error) {
	if getSessionID(ctx) == "" {
		return nil, errors.New("must be logged in to get zone")
	}

	return c.postRequest(ctx, "dns-zone-get", url.Values{"domain": {dns01.UnFqdn(domain)}})
}

// PutZone uploads DNS zone to Joker DMAPI.
func (c *Client) PutZone(ctx context.Context, domain, zone string) (*Response, error) {
	if getSessionID(ctx) == "" {
		return nil, errors.New("must be logged in to put zone")
	}

	return c.postRequest(ctx, "dns-zone-put", url.Values{"domain": {dns01.UnFqdn(domain)}, "zone": {strings.TrimSpace(zone)}})
}

// postRequest performs actual HTTP request.
func (c *Client) postRequest(ctx context.Context, cmd string, data url.Values) (*Response, error) {
	endpoint, err := url.JoinPath(c.BaseURL, cmd)
	if err != nil {
		return nil, err
	}

	if getSessionID(ctx) != "" {
		data.Set("auth-sid", getSessionID(ctx))
	}

	if c.Debug {
		log.Infof("postRequest:\n\tURL: %q\n\tData: %v", endpoint, data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	return parseResponse(string(raw)), nil
}

// parseResponse parses HTTP response body.
func parseResponse(message string) *Response {
	r := &Response{Headers: url.Values{}, StatusCode: -1}

	lines, body, _ := strings.Cut(message, "\n\n")

	for line := range strings.Lines(lines) {
		if strings.TrimSpace(line) == "" {
			continue
		}

		k, v, _ := strings.Cut(line, ":")

		val := strings.TrimSpace(v)

		r.Headers.Add(k, val)

		switch k {
		case "Status-Code":
			i, err := strconv.Atoi(val)
			if err == nil {
				r.StatusCode = i
			}
		case "Status-Text":
			r.StatusText = val
		case "Auth-Sid":
			r.AuthSid = val
		}
	}

	r.Body = body

	return r
}

// Temporary workaround, until it get fixed on API side.
func fixTxtLines(line string) string {
	fields := strings.Fields(line)

	if len(fields) < 6 || fields[1] != "TXT" {
		return line
	}

	if fields[3][0] == '"' && fields[4] == `"` {
		fields[3] = strings.TrimSpace(fields[3]) + `"`
		fields = append(fields[:4], fields[5:]...)
	}

	return strings.Join(fields, " ")
}

// RemoveTxtEntryFromZone clean-ups all TXT records with given name.
func RemoveTxtEntryFromZone(zone, relative string) (string, bool) {
	prefix := fmt.Sprintf("%s TXT 0 ", relative)

	modified := false

	var zoneEntries []string

	for line := range strings.Lines(zone) {
		if strings.HasPrefix(line, prefix) {
			modified = true
			continue
		}

		zoneEntries = append(zoneEntries, line)
	}

	return strings.TrimSpace(strings.Join(zoneEntries, "\n")), modified
}

// AddTxtEntryToZone returns DNS zone with added TXT record.
func AddTxtEntryToZone(zone, relative, value string, ttl int) string {
	var zoneEntries []string

	for line := range strings.Lines(zone) {
		zoneEntries = append(zoneEntries, fixTxtLines(line))
	}

	newZoneEntry := fmt.Sprintf("%s TXT 0 %q %d", relative, value, ttl)
	zoneEntries = append(zoneEntries, newZoneEntry)

	return strings.TrimSpace(strings.Join(zoneEntries, "\n"))
}
