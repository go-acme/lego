// Package dmapi Client for DMAPI joker.com.
// https://joker.com/faq/category/39/22-dmapi.html
package dmapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
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
	authSid  string
}

// Client a DMAPI Client.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string

	Debug bool

	auth AuthInfo
}

// NewClient creates a new DMAPI Client.
func NewClient(auth AuthInfo) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    defaultBaseURL,
		Debug:      false,
		auth:       auth,
	}
}

// Login performs a login to Joker's DMAPI.
func (c *Client) Login() (*Response, error) {
	if c.auth.authSid != "" {
		// already logged in
		return nil, nil
	}

	var values url.Values
	switch {
	case c.auth.Username != "" && c.auth.Password != "":
		values = url.Values{
			"username": {c.auth.Username},
			"password": {c.auth.Password},
		}
	case c.auth.APIKey != "":
		values = url.Values{"api-key": {c.auth.APIKey}}
	default:
		return nil, errors.New("no username and password or api-key")
	}

	response, err := c.postRequest("login", values)
	if err != nil {
		return response, err
	}

	if response == nil {
		return nil, errors.New("login returned nil response")
	}

	if response.AuthSid == "" {
		return response, errors.New("login did not return valid Auth-Sid")
	}

	c.auth.authSid = response.AuthSid

	return response, nil
}

// Logout closes authenticated session with Joker's DMAPI.
func (c *Client) Logout() (*Response, error) {
	if c.auth.authSid == "" {
		return nil, errors.New("already logged out")
	}

	response, err := c.postRequest("logout", url.Values{})
	if err == nil {
		c.auth.authSid = ""
	}
	return response, err
}

// GetZone returns content of DNS zone for domain.
func (c *Client) GetZone(domain string) (*Response, error) {
	if c.auth.authSid == "" {
		return nil, errors.New("must be logged in to get zone")
	}

	return c.postRequest("dns-zone-get", url.Values{"domain": {dns01.UnFqdn(domain)}})
}

// PutZone uploads DNS zone to Joker DMAPI.
func (c *Client) PutZone(domain, zone string) (*Response, error) {
	if c.auth.authSid == "" {
		return nil, errors.New("must be logged in to put zone")
	}

	return c.postRequest("dns-zone-put", url.Values{"domain": {dns01.UnFqdn(domain)}, "zone": {strings.TrimSpace(zone)}})
}

// postRequest performs actual HTTP request.
func (c *Client) postRequest(cmd string, data url.Values) (*Response, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	endpoint, err := baseURL.Parse(path.Join(baseURL.Path, cmd))
	if err != nil {
		return nil, err
	}

	if c.auth.authSid != "" {
		data.Set("auth-sid", c.auth.authSid)
	}

	if c.Debug {
		log.Infof("postRequest:\n\tURL: %q\n\tData: %v", endpoint.String(), data)
	}

	resp, err := c.HTTPClient.PostForm(endpoint.String(), data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d [%s]: %v", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	return parseResponse(string(body)), nil
}

// parseResponse parses HTTP response body.
func parseResponse(message string) *Response {
	r := &Response{Headers: url.Values{}, StatusCode: -1}

	parts := strings.SplitN(message, "\n\n", 2)

	for _, line := range strings.Split(parts[0], "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		kv := strings.SplitN(line, ":", 2)

		val := ""
		if len(kv) == 2 {
			val = strings.TrimSpace(kv[1])
		}

		r.Headers.Add(kv[0], val)

		switch kv[0] {
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

	if len(parts) > 1 {
		r.Body = parts[1]
	}

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
	for _, line := range strings.Split(zone, "\n") {
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

	for _, line := range strings.Split(zone, "\n") {
		zoneEntries = append(zoneEntries, fixTxtLines(line))
	}

	newZoneEntry := fmt.Sprintf("%s TXT 0 %q %d", relative, value, ttl)
	zoneEntries = append(zoneEntries, newZoneEntry)

	return strings.TrimSpace(strings.Join(zoneEntries, "\n"))
}
