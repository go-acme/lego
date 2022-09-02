package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
)

const defaultBaseURL = "https://api.cloudns.net/dns/"

// Client the ClouDNS client.
type Client struct {
	authID       string
	subAuthID    string
	authPassword string
	HTTPClient   *http.Client
	BaseURL      *url.URL
}

// NewClient creates a ClouDNS client.
func NewClient(authID, subAuthID, authPassword string) (*Client, error) {
	if authID == "" && subAuthID == "" {
		return nil, errors.New("credentials missing: authID or subAuthID")
	}

	if authPassword == "" {
		return nil, errors.New("credentials missing: authPassword")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		authID:       authID,
		subAuthID:    subAuthID,
		authPassword: authPassword,
		HTTPClient:   &http.Client{},
		BaseURL:      baseURL,
	}, nil
}

// GetZone Get domain name information for a FQDN.
func (c *Client) GetZone(authFQDN string) (*Zone, error) {
	authZone, err := dns01.FindZoneByFqdn(authFQDN)
	if err != nil {
		return nil, err
	}

	authZoneName := dns01.UnFqdn(authZone)

	endpoint, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, "get-zone-info.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	q := endpoint.Query()
	q.Set("domain-name", authZoneName)
	endpoint.RawQuery = q.Encode()

	result, err := c.doRequest(http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}

	var zone Zone

	if len(result) > 0 {
		if err = json.Unmarshal(result, &zone); err != nil {
			return nil, fmt.Errorf("failed to unmarshal zone: %w", err)
		}
	}

	if zone.Name == authZoneName {
		return &zone, nil
	}

	return nil, fmt.Errorf("zone %s not found for authFQDN %s", authZoneName, authFQDN)
}

// FindTxtRecord returns the TXT record a zone ID and a FQDN.
func (c *Client) FindTxtRecord(zoneName, fqdn string) (*TXTRecord, error) {
	host := dns01.UnFqdn(strings.TrimSuffix(dns01.UnFqdn(fqdn), zoneName))

	reqURL, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, "records.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	q := reqURL.Query()
	q.Set("domain-name", zoneName)
	q.Set("host", host)
	q.Set("type", "TXT")
	reqURL.RawQuery = q.Encode()

	result, err := c.doRequest(http.MethodGet, reqURL)
	if err != nil {
		return nil, err
	}

	// the API returns [] when there is no records.
	if string(result) == "[]" {
		return nil, nil
	}

	var records map[string]TXTRecord
	if err = json.Unmarshal(result, &records); err != nil {
		return nil, fmt.Errorf("failed to unmarshall TXT records: %w: %s", err, string(result))
	}

	for _, record := range records {
		if record.Host == host && record.Type == "TXT" {
			return &record, nil
		}
	}

	return nil, nil
}

// ListTxtRecords returns the TXT records a zone ID and a FQDN.
func (c *Client) ListTxtRecords(zoneName, fqdn string) ([]TXTRecord, error) {
	host := dns01.UnFqdn(strings.TrimSuffix(dns01.UnFqdn(fqdn), zoneName))

	reqURL, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, "records.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	q := reqURL.Query()
	q.Set("domain-name", zoneName)
	q.Set("host", host)
	q.Set("type", "TXT")
	reqURL.RawQuery = q.Encode()

	result, err := c.doRequest(http.MethodGet, reqURL)
	if err != nil {
		return nil, err
	}

	// the API returns [] when there is no records.
	if string(result) == "[]" {
		return nil, nil
	}

	var raw map[string]TXTRecord
	if err = json.Unmarshal(result, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshall TXT records: %w: %s", err, string(result))
	}

	var records []TXTRecord
	for _, record := range raw {
		if record.Host == host && record.Type == "TXT" {
			records = append(records, record)
		}
	}

	return records, nil
}

// AddTxtRecord adds a TXT record.
func (c *Client) AddTxtRecord(zoneName, fqdn, value string, ttl int) error {
	host := dns01.UnFqdn(strings.TrimSuffix(dns01.UnFqdn(fqdn), zoneName))

	reqURL, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, "add-record.json"))
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	q := reqURL.Query()
	q.Set("domain-name", zoneName)
	q.Set("host", host)
	q.Set("record", value)
	q.Set("ttl", strconv.Itoa(ttlRounder(ttl)))
	q.Set("record-type", "TXT")
	reqURL.RawQuery = q.Encode()

	raw, err := c.doRequest(http.MethodPost, reqURL)
	if err != nil {
		return err
	}

	resp := apiResponse{}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %w: %s", err, string(raw))
	}

	if resp.Status != "Success" {
		return fmt.Errorf("failed to add TXT record: %s %s", resp.Status, resp.StatusDescription)
	}

	return nil
}

// RemoveTxtRecord removes a TXT record.
func (c *Client) RemoveTxtRecord(recordID int, zoneName string) error {
	reqURL, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, "delete-record.json"))
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	q := reqURL.Query()
	q.Set("domain-name", zoneName)
	q.Set("record-id", strconv.Itoa(recordID))
	reqURL.RawQuery = q.Encode()

	raw, err := c.doRequest(http.MethodPost, reqURL)
	if err != nil {
		return err
	}

	resp := apiResponse{}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %w: %s", err, string(raw))
	}

	if resp.Status != "Success" {
		return fmt.Errorf("failed to remove TXT record: %s %s", resp.Status, resp.StatusDescription)
	}

	return nil
}

// GetUpdateStatus gets sync progress of all CloudDNS NS servers.
func (c *Client) GetUpdateStatus(zoneName string) (*SyncProgress, error) {
	reqURL, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, "update-status.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	q := reqURL.Query()
	q.Set("domain-name", zoneName)
	reqURL.RawQuery = q.Encode()

	result, err := c.doRequest(http.MethodGet, reqURL)
	if err != nil {
		return nil, err
	}

	// the API returns [] when there is no records.
	if string(result) == "[]" {
		return nil, errors.New("no nameservers records returned")
	}

	var records []UpdateRecord
	if err = json.Unmarshal(result, &records); err != nil {
		return nil, fmt.Errorf("failed to unmarshal UpdateRecord: %w: %s", err, string(result))
	}

	updatedCount := 0
	for _, record := range records {
		if record.Updated {
			updatedCount++
		}
	}

	return &SyncProgress{Complete: updatedCount == len(records), Updated: updatedCount, Total: len(records)}, nil
}

func (c *Client) doRequest(method string, uri *url.URL) (json.RawMessage, error) {
	req, err := c.buildRequest(method, uri)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid code (%d), error: %s", resp.StatusCode, content)
	}

	return content, nil
}

func (c *Client) buildRequest(method string, uri *url.URL) (*http.Request, error) {
	q := uri.Query()

	if c.subAuthID != "" {
		q.Set("sub-auth-id", c.subAuthID)
	} else {
		q.Set("auth-id", c.authID)
	}

	q.Set("auth-password", c.authPassword)

	uri.RawQuery = q.Encode()

	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	return req, nil
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}

// Rounds the given TTL in seconds to the next accepted value.
// Accepted TTL values are:
//   - 60 = 1 minute
//   - 300 = 5 minutes
//   - 900 = 15 minutes
//   - 1800 = 30 minutes
//   - 3600 = 1 hour
//   - 21600 = 6 hours
//   - 43200 = 12 hours
//   - 86400 = 1 day
//   - 172800 = 2 days
//   - 259200 = 3 days
//   - 604800 = 1 week
//   - 1209600 = 2 weeks
//   - 2592000 = 1 month
//
// See https://www.cloudns.net/wiki/article/58/ for details.
func ttlRounder(ttl int) int {
	for _, validTTL := range []int{60, 300, 900, 1800, 3600, 21600, 43200, 86400, 172800, 259200, 604800, 1209600} {
		if ttl <= validTTL {
			return validTTL
		}
	}

	return 2592000
}
