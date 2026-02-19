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
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.leaseweb.com/hosting/v2"

const AuthHeader = "X-LSW-Auth"

// Client the Leaseweb API client.
type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRRSet creates a resource record set.
// https://developer.leaseweb.com/docs/#tag/DNS/operation/createResourceRecordSet
func (c *Client) CreateRRSet(ctx context.Context, domainName string, rrset RRSet) (*RRSet, error) {
	endpoint := c.BaseURL.JoinPath("domains", domainName, "resourceRecordSets")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, rrset)
	if err != nil {
		return nil, err
	}

	result := &RRSet{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetRRSet gets a resource record set.
// https://developer.leaseweb.com/docs/#tag/DNS/operation/getResourceRecordSet
func (c *Client) GetRRSet(ctx context.Context, domainName, name, rType string) (*RRSet, error) {
	endpoint := c.BaseURL.JoinPath("domains", domainName, "resourceRecordSets", name, rType)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &RRSet{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateRRSet updates a resource record set.
// https://developer.leaseweb.com/docs/#tag/DNS/operation/updateResourceRecordSet
func (c *Client) UpdateRRSet(ctx context.Context, domainName string, rrset RRSet) (*RRSet, error) {
	endpoint := c.BaseURL.JoinPath("domains", domainName, "resourceRecordSets", rrset.Name, rrset.Type)

	// Reset values that are not allowed to be updated.
	rrset.Name = ""
	rrset.Type = ""
	rrset.Editable = false

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, rrset)
	if err != nil {
		return nil, err
	}

	result := &RRSet{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRRSet deletes a resource record set.
// https://developer.leaseweb.com/docs/#tag/DNS/operation/deleteResourceRecordSet
func (c *Client) DeleteRRSet(ctx context.Context, domainName, name, rType string) error {
	endpoint := c.BaseURL.JoinPath("domains", domainName, "resourceRecordSets", name, rType)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Add(AuthHeader, c.apiKey)

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

	var errAPI APIError

	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return &NotFoundError{APIError{
				CorrelationID: resp.Header.Get("Correlation-Id"),
				ErrorCode:     strconv.Itoa(http.StatusNotFound),
				ErrorMessage:  string(raw),
			}}
		}

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	if errAPI.ErrorCode == strconv.Itoa(http.StatusNotFound) {
		return &NotFoundError{APIError: errAPI}
	}

	return &errAPI
}

// TTLRounder rounds the given TTL in seconds to the next accepted value.
// Accepted TTL values are: 60, 300, 1800, 3600, 14400, 28800, 43200, 86400.
func TTLRounder(ttl int) int {
	for _, validTTL := range []int{60, 300, 1800, 3600, 14400, 28800, 43200, 86400} {
		if ttl <= validTTL {
			return validTTL
		}
	}

	return 3600
}
