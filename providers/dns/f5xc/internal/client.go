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
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultHost = "console.ves.volterra.io"

const authorizationHeader = "Authorization"

// Client the F5 XC API client.
type Client struct {
	apiToken string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiToken, tenantName string) (*Client, error) {
	if apiToken == "" {
		return nil, errors.New("credentials missing")
	}

	if tenantName == "" {
		return nil, errors.New("missing tenant name")
	}

	baseURL, err := url.Parse(fmt.Sprintf("https://%s.%s", tenantName, defaultHost))
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	return &Client{
		apiToken:   apiToken,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRRSet creates RRSet.
// https://docs.cloud.f5.com/docs-v2/api/dns-zone-rrset#operation/ves.io.schema.dns_zone.rrset.CustomAPI.Create
func (c *Client) CreateRRSet(ctx context.Context, dnsZoneName, groupName string, rrSet RRSet) (*APIRRSet, error) {
	endpoint := c.baseURL.JoinPath("api", "config", "dns", "namespaces", "system", "dns_zones", dnsZoneName, "rrsets", groupName)

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, APIRRSet{
		DNSZoneName: dnsZoneName,
		GroupName:   groupName,
		RRSet:       rrSet,
	})
	if err != nil {
		return nil, err
	}

	result := &APIRRSet{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetRRSet gets RRSets.
// https://docs.cloud.f5.com/docs-v2/api/dns-zone-rrset#operation/ves.io.schema.dns_zone.rrset.CustomAPI.Get
func (c *Client) GetRRSet(ctx context.Context, dnsZoneName, groupName, recordName, recordType string) (*APIRRSet, error) {
	endpoint := c.baseURL.JoinPath("api", "config", "dns", "namespaces", "system", "dns_zones", dnsZoneName, "rrsets", groupName, recordName, recordType)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &APIRRSet{}

	err = c.do(req, result)
	if err != nil {
		usce := &APIError{}
		if errors.As(err, &usce) && usce.StatusCode == http.StatusNotFound {
			return nil, nil
		}

		return nil, err
	}

	return result, nil
}

// DeleteRRSet deletes RRSet.
// https://docs.cloud.f5.com/docs-v2/api/dns-zone-rrset#operation/ves.io.schema.dns_zone.rrset.CustomAPI.Delete
func (c *Client) DeleteRRSet(ctx context.Context, dnsZoneName, groupName, recordName, recordType string) (*APIRRSet, error) {
	endpoint := c.baseURL.JoinPath("api", "config", "dns", "namespaces", "system", "dns_zones", dnsZoneName, "rrsets", groupName, recordName, recordType)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &APIRRSet{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ReplaceRRSet replaces RRSet.
// https://docs.cloud.f5.com/docs-v2/api/dns-zone-rrset#operation/ves.io.schema.dns_zone.rrset.CustomAPI.Replace
func (c *Client) ReplaceRRSet(ctx context.Context, dnsZoneName, groupName, recordName, recordType string, rrSet RRSet) (*APIRRSet, error) {
	endpoint := c.baseURL.JoinPath("api", "config", "dns", "namespaces", "system", "dns_zones", dnsZoneName, "rrsets", groupName, recordName, recordType)

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, APIRRSet{
		DNSZoneName: dnsZoneName,
		GroupName:   groupName,
		RRSet:       rrSet,
		Type:        recordType,
	})
	if err != nil {
		return nil, err
	}

	result := &APIRRSet{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set(authorizationHeader, "APIToken "+c.apiToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
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

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	apiErr := APIError{StatusCode: resp.StatusCode}

	err := json.Unmarshal(raw, &apiErr)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &apiErr
}
