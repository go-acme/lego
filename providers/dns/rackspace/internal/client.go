package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

type Client struct {
	token string

	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(endpoint, token string) (*Client, error) {
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		token:      token,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// AddRecord Adds one  record to a specified domain.
// https://docs.rackspace.com/docs/cloud-dns/v1/api-reference/records#add-records
func (c *Client) AddRecord(ctx context.Context, zoneID string, record Record) error {
	endpoint := c.baseURL.JoinPath("domains", zoneID, "records")

	records := Records{Records: []Record{record}}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, records)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecord Deletes a record from the domain.
// https://docs.rackspace.com/docs/cloud-dns/v1/api-reference/records#delete-records
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.baseURL.JoinPath("domains", zoneID, "records")

	query := endpoint.Query()
	query.Set("id", recordID)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetHostedZoneID performs a lookup to get the DNS zone which needs modifying for a given FQDN.
func (c *Client) GetHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone: %w", err)
	}

	zoneSearchResponse, err := c.listDomainsByName(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return "", err
	}

	// If nothing was returned, or for whatever reason more than 1 was returned (the search uses exact match, so should not occur)
	if zoneSearchResponse.TotalEntries != 1 {
		return "", fmt.Errorf("found %d zones for %s in Rackspace for domain %s", zoneSearchResponse.TotalEntries, authZone, fqdn)
	}

	return zoneSearchResponse.HostedZones[0].ID, nil
}

// listDomainsByName Filters domains by domain name.
// https://docs.rackspace.com/docs/cloud-dns/v1/api-reference/domains#list-domains-by-name
func (c *Client) listDomainsByName(ctx context.Context, domain string) (*ZoneSearchResponse, error) {
	endpoint := c.baseURL.JoinPath("domains")

	query := endpoint.Query()
	query.Set("name", domain)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zoneSearchResponse ZoneSearchResponse
	err = c.do(req, &zoneSearchResponse)
	if err != nil {
		return nil, err
	}

	return &zoneSearchResponse, nil
}

// FindTxtRecord searches a DNS zone for a TXT record with a specific name.
func (c *Client) FindTxtRecord(ctx context.Context, fqdn, zoneID string) (*Record, error) {
	records, err := c.searchRecords(ctx, zoneID, dns01.UnFqdn(fqdn), "TXT")
	if err != nil {
		return nil, err
	}

	switch len(records.Records) {
	case 1:
	case 0:
		return nil, fmt.Errorf("no TXT record found for %s", fqdn)
	default:
		return nil, fmt.Errorf("more than 1 TXT record found for %s", fqdn)
	}

	return &records.Records[0], nil
}

// https://docs.rackspace.com/docs/cloud-dns/v1/api-reference/records#search-records
func (c *Client) searchRecords(ctx context.Context, zoneID, recordName, recordType string) (*Records, error) {
	endpoint := c.baseURL.JoinPath("domains", zoneID, "records")

	query := endpoint.Query()
	query.Set("type", recordType)
	query.Set("name", recordName)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records Records
	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	return &records, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("X-Auth-Token", c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func newJSONRequest[T string | *url.URL](ctx context.Context, method string, endpoint T, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s", endpoint), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
