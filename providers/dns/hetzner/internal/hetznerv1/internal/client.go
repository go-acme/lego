package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://api.hetzner.cloud/v1"

const (
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusError   = "error"
)

// Client the Hetzner API client.
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		BaseURL:    baseURL,
		httpClient: hc,
	}, nil
}

// AddRRSetRecords adds records to an RRSet.
// https://docs.hetzner.cloud/reference/cloud#zone-rrset-actions-add-records-to-an-rrset
func (c *Client) AddRRSetRecords(ctx context.Context, zoneIDName, recordType, recordName string, ttl int, records []Record) (*Action, error) {
	endpoint := c.BaseURL.JoinPath("zones", zoneIDName, "rrsets", recordName, recordType, "actions", "add_records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, RRSet{TTL: ttl, Records: records})
	if err != nil {
		return nil, err
	}

	var result ActionResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Action, nil
}

// RemoveRRSetRecords removes records from an RRSet.
// https://docs.hetzner.cloud/reference/cloud#zone-rrset-actions-remove-records-from-an-rrset
func (c *Client) RemoveRRSetRecords(ctx context.Context, zoneIDName, recordType, recordName string, records []Record) (*Action, error) {
	endpoint := c.BaseURL.JoinPath("zones", zoneIDName, "rrsets", recordName, recordType, "actions", "remove_records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, RRSet{Records: records})
	if err != nil {
		return nil, err
	}

	var result ActionResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Action, nil
}

// GetAction gets an action.
// https://docs.hetzner.cloud/reference/cloud#actions-get-an-action
func (c *Client) GetAction(ctx context.Context, id int) (*Action, error) {
	endpoint := c.BaseURL.JoinPath("actions", strconv.Itoa(id))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result ActionResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Action, nil
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
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
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}

func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
