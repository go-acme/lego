package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-acme/lego/v5/internal/errutils"
)

const APIBaseURL = "https://api.mythic-beasts.com/dns/v2"

// Client the Mythic Beasts API client.
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(APIBaseURL)

	return &Client{
		BaseURL:    baseURL,
		httpClient: hc,
	}
}

// CreateRecord creates a TXT record.
func (c *Client) CreateRecord(ctx context.Context, zone string, record Record) error {
	resp, err := c.createRecord(ctx, zone, record)
	if err != nil {
		return err
	}

	if resp.Added != 1 {
		return fmt.Errorf("did not add TXT record for some reason: %s", resp.Message)
	}

	// Success
	return nil
}

// RemoveTXTRecord removes a TXT records.
func (c *Client) RemoveTXTRecord(ctx context.Context, zone, leaf, value string) error {
	resp, err := c.removeTXTRecord(ctx, zone, leaf, "TXT", value)
	if err != nil {
		return err
	}

	if resp.Removed != 1 {
		return fmt.Errorf("did not remove TXT record for some reason: %s", resp.Message)
	}

	// Success
	return nil
}

// https://www.mythic-beasts.com/support/api/dnsv2#ep-post-zoneszonerecords
func (c *Client) createRecord(ctx context.Context, zone string, record Record) (*recordResponse, error) {
	endpoint := c.BaseURL.JoinPath("zones", zone, "records", record.Host, record.Type)

	payload := createRecordsRequest{
		Records: []Record{record},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	resp := new(recordResponse)

	err = c.do(req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// https://www.mythic-beasts.com/support/api/dnsv2#ep-delete-zoneszonerecords
func (c *Client) removeTXTRecord(ctx context.Context, zone, leaf, recordType, value string) (*recordResponse, error) {
	endpoint := c.BaseURL.JoinPath("zones", zone, "records", leaf, recordType)

	query := endpoint.Query()
	query.Add("data", value)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp := new(recordResponse)

	err = c.do(req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return parseError(req, resp)
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

	errAPI := &APIError{StatusCode: resp.StatusCode}

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}
