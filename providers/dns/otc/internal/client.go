package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

type Client struct {
	username    string
	password    string
	domainName  string
	projectName string

	IdentityEndpoint string
	token            string
	muToken          sync.Mutex

	baseURL   *url.URL
	muBaseURL sync.Mutex

	HTTPClient *http.Client
}

func NewClient(username, password, domainName, projectName string) *Client {
	return &Client{
		username:         username,
		password:         password,
		domainName:       domainName,
		projectName:      projectName,
		IdentityEndpoint: DefaultIdentityEndpoint,
		HTTPClient:       &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetZoneID(ctx context.Context, zone string) (string, error) {
	zonesResp, err := c.getZones(ctx, zone)
	if err != nil {
		return "", err
	}

	if len(zonesResp.Zones) < 1 {
		return "", fmt.Errorf("zone %s not found", zone)
	}

	for _, z := range zonesResp.Zones {
		if z.Name == zone {
			return z.ID, nil
		}
	}

	return "", fmt.Errorf("zone %s not found", zone)
}

// https://docs.otc.t-systems.com/domain-name-service/api-ref/apis/public_zone_management/querying_public_zones.html
func (c *Client) getZones(ctx context.Context, zone string) (*ZonesResponse, error) {
	c.muBaseURL.Lock()
	endpoint := c.baseURL.JoinPath("zones")
	c.muBaseURL.Unlock()

	query := endpoint.Query()
	query.Set("name", zone)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zones ZonesResponse
	err = c.do(req, &zones)
	if err != nil {
		return nil, err
	}

	return &zones, nil
}

func (c *Client) GetRecordSetID(ctx context.Context, zoneID, fqdn string) (string, error) {
	recordSetsRes, err := c.getRecordSet(ctx, zoneID, fqdn)
	if err != nil {
		return "", err
	}

	if len(recordSetsRes.RecordSets) < 1 {
		return "", errors.New("record not found")
	}

	if len(recordSetsRes.RecordSets) > 1 {
		return "", errors.New("to many records found")
	}

	if recordSetsRes.RecordSets[0].ID == "" {
		return "", errors.New("id not found")
	}

	return recordSetsRes.RecordSets[0].ID, nil
}

// https://docs.otc.t-systems.com/domain-name-service/api-ref/apis/record_set_management/querying_all_record_sets.html
func (c *Client) getRecordSet(ctx context.Context, zoneID, fqdn string) (*RecordSetsResponse, error) {
	c.muBaseURL.Lock()
	endpoint := c.baseURL.JoinPath("zones", zoneID, "recordsets")
	c.muBaseURL.Unlock()

	query := endpoint.Query()
	query.Set("type", "TXT")
	query.Set("name", fqdn)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var recordSetsRes RecordSetsResponse
	err = c.do(req, &recordSetsRes)
	if err != nil {
		return nil, err
	}

	return &recordSetsRes, nil
}

// CreateRecordSet creates a record.
// https://docs.otc.t-systems.com/domain-name-service/api-ref/apis/record_set_management/creating_a_record_set.html
func (c *Client) CreateRecordSet(ctx context.Context, zoneID string, record RecordSets) error {
	c.muBaseURL.Lock()
	endpoint := c.baseURL.JoinPath("zones", zoneID, "recordsets")
	c.muBaseURL.Unlock()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteRecordSet delete a record set.
// https://docs.otc.t-systems.com/domain-name-service/api-ref/apis/record_set_management/deleting_a_record_set.html
func (c *Client) DeleteRecordSet(ctx context.Context, zoneID, recordID string) error {
	c.muBaseURL.Lock()
	endpoint := c.baseURL.JoinPath("zones", zoneID, "recordsets", recordID)
	c.muBaseURL.Unlock()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	c.muToken.Lock()
	if c.token != "" {
		req.Header.Set("X-Auth-Token", c.token)
	}
	c.muToken.Unlock()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
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
