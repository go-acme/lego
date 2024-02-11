package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.cloudns.net/dns/"

// Client the ClouDNS client.
type Client struct {
	authID       string
	subAuthID    string
	authPassword string

	BaseURL    *url.URL
	HTTPClient *http.Client
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
		BaseURL:      baseURL,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetZone Get domain name information for a FQDN.
func (c *Client) GetZone(ctx context.Context, authFQDN string) (*Zone, error) {
	authZone, err := dns01.FindZoneByFqdn(authFQDN)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	authZoneName := dns01.UnFqdn(authZone)

	endpoint := c.BaseURL.JoinPath("get-zone-info.json")

	q := endpoint.Query()
	q.Set("domain-name", authZoneName)
	endpoint.RawQuery = q.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}

	rawMessage, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var zone Zone

	if len(rawMessage) > 0 {
		if err = json.Unmarshal(rawMessage, &zone); err != nil {
			return nil, errutils.NewUnmarshalError(req, http.StatusOK, rawMessage, err)
		}
	}

	if zone.Name == authZoneName {
		return &zone, nil
	}

	return nil, fmt.Errorf("zone %s not found for authFQDN %s", authZoneName, authFQDN)
}

// FindTxtRecord returns the TXT record a zone ID and a FQDN.
func (c *Client) FindTxtRecord(ctx context.Context, zoneName, fqdn string) (*TXTRecord, error) {
	subDomain, err := dns01.ExtractSubDomain(fqdn, zoneName)
	if err != nil {
		return nil, err
	}

	endpoint := c.BaseURL.JoinPath("records.json")

	q := endpoint.Query()
	q.Set("domain-name", zoneName)
	q.Set("host", subDomain)
	q.Set("type", "TXT")
	endpoint.RawQuery = q.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}

	rawMessage, err := c.do(req)
	if err != nil {
		return nil, err
	}

	// the API returns [] when there is no records.
	if string(rawMessage) == "[]" {
		return nil, nil
	}

	var records map[string]TXTRecord
	if err = json.Unmarshal(rawMessage, &records); err != nil {
		return nil, errutils.NewUnmarshalError(req, http.StatusOK, rawMessage, err)
	}

	for _, record := range records {
		if record.Host == subDomain && record.Type == "TXT" {
			return &record, nil
		}
	}

	return nil, nil
}

// ListTxtRecords returns the TXT records a zone ID and a FQDN.
func (c *Client) ListTxtRecords(ctx context.Context, zoneName, fqdn string) ([]TXTRecord, error) {
	subDomain, err := dns01.ExtractSubDomain(fqdn, zoneName)
	if err != nil {
		return nil, err
	}

	endpoint := c.BaseURL.JoinPath("records.json")

	q := endpoint.Query()
	q.Set("domain-name", zoneName)
	q.Set("host", subDomain)
	q.Set("type", "TXT")
	endpoint.RawQuery = q.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}

	rawMessage, err := c.do(req)
	if err != nil {
		return nil, err
	}

	// the API returns [] when there is no records.
	if string(rawMessage) == "[]" {
		return nil, nil
	}

	var raw map[string]TXTRecord
	if err = json.Unmarshal(rawMessage, &raw); err != nil {
		return nil, errutils.NewUnmarshalError(req, http.StatusOK, rawMessage, err)
	}

	var records []TXTRecord
	for _, record := range raw {
		if record.Host == subDomain && record.Type == "TXT" {
			records = append(records, record)
		}
	}

	return records, nil
}

// AddTxtRecord adds a TXT record.
func (c *Client) AddTxtRecord(ctx context.Context, zoneName, fqdn, value string, ttl int) error {
	subDomain, err := dns01.ExtractSubDomain(fqdn, zoneName)
	if err != nil {
		return err
	}

	endpoint := c.BaseURL.JoinPath("add-record.json")

	q := endpoint.Query()
	q.Set("domain-name", zoneName)
	q.Set("host", subDomain)
	q.Set("record", value)
	q.Set("ttl", strconv.Itoa(ttlRounder(ttl)))
	q.Set("record-type", "TXT")
	endpoint.RawQuery = q.Encode()

	req, err := c.newRequest(ctx, http.MethodPost, endpoint)
	if err != nil {
		return err
	}

	rawMessage, err := c.do(req)
	if err != nil {
		return err
	}

	resp := apiResponse{}
	if err = json.Unmarshal(rawMessage, &resp); err != nil {
		return errutils.NewUnmarshalError(req, http.StatusOK, rawMessage, err)
	}

	if resp.Status != "Success" {
		return fmt.Errorf("failed to add TXT record: %s %s", resp.Status, resp.StatusDescription)
	}

	return nil
}

// RemoveTxtRecord removes a TXT record.
func (c *Client) RemoveTxtRecord(ctx context.Context, recordID int, zoneName string) error {
	endpoint := c.BaseURL.JoinPath("delete-record.json")

	q := endpoint.Query()
	q.Set("domain-name", zoneName)
	q.Set("record-id", strconv.Itoa(recordID))
	endpoint.RawQuery = q.Encode()

	req, err := c.newRequest(ctx, http.MethodPost, endpoint)
	if err != nil {
		return err
	}

	rawMessage, err := c.do(req)
	if err != nil {
		return err
	}

	resp := apiResponse{}
	if err = json.Unmarshal(rawMessage, &resp); err != nil {
		return errutils.NewUnmarshalError(req, http.StatusOK, rawMessage, err)
	}

	if resp.Status != "Success" {
		return fmt.Errorf("failed to remove TXT record: %s %s", resp.Status, resp.StatusDescription)
	}

	return nil
}

// GetUpdateStatus gets sync progress of all CloudDNS NS servers.
func (c *Client) GetUpdateStatus(ctx context.Context, zoneName string) (*SyncProgress, error) {
	endpoint := c.BaseURL.JoinPath("update-status.json")

	q := endpoint.Query()
	q.Set("domain-name", zoneName)
	endpoint.RawQuery = q.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}

	rawMessage, err := c.do(req)
	if err != nil {
		return nil, err
	}

	// the API returns [] when there is no records.
	if string(rawMessage) == "[]" {
		return nil, errors.New("no nameservers records returned")
	}

	var records []UpdateRecord
	if err = json.Unmarshal(rawMessage, &records); err != nil {
		return nil, errutils.NewUnmarshalError(req, http.StatusOK, rawMessage, err)
	}

	updatedCount := 0
	for _, record := range records {
		if record.Updated {
			updatedCount++
		}
	}

	return &SyncProgress{Complete: updatedCount == len(records), Updated: updatedCount, Total: len(records)}, nil
}

func (c *Client) newRequest(ctx context.Context, method string, endpoint *url.URL) (*http.Request, error) {
	q := endpoint.Query()

	if c.subAuthID != "" {
		q.Set("sub-auth-id", c.subAuthID)
	} else {
		q.Set("auth-id", c.authID)
	}

	q.Set("auth-password", c.authPassword)

	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	return req, nil
}

func (c *Client) do(req *http.Request) (json.RawMessage, error) {
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

	return raw, nil
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
