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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://dns-service.iran.liara.ir"

// Client a Liara DNS API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{httpClient: hc, baseURL: baseURL}
}

// GetRecords gets the records of a domain.
// https://openapi.liara.ir/?urls.primaryName=DNS
func (c *Client) GetRecords(ctx context.Context, domainName string) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("api", "v1", "zones", domainName, "dns-records")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response Response[[]Record]
	err = json.Unmarshal(raw, &response)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return response.Data, nil
}

// CreateRecord creates a record.
func (c *Client) CreateRecord(ctx context.Context, domainName string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("api", "v1", "zones", domainName, "dns-records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response Response[*Record]
	err = json.Unmarshal(raw, &response)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return response.Data, nil
}

// GetRecord gets a specific record.
func (c *Client) GetRecord(ctx context.Context, domainName, recordID string) (*Record, error) {
	endpoint := c.baseURL.JoinPath("api", "v1", "zones", domainName, "dns-records", recordID)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response Response[*Record]
	err = json.Unmarshal(raw, &response)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return response.Data, nil
}

// DeleteRecord deletes a record.
func (c *Client) DeleteRecord(ctx context.Context, domainName, recordID string) error {
	endpoint := c.baseURL.JoinPath("api", "v1", "zones", domainName, "dns-records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return parseError(req, resp)
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

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, &errAPI)
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
