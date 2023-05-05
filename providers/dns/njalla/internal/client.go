package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const apiEndpoint = "https://njal.la/api/1/"

const authorizationHeader = "Authorization"

// Client is a Njalla API client.
type Client struct {
	token string

	apiEndpoint string
	HTTPClient  *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) *Client {
	return &Client{
		token:       token,
		apiEndpoint: apiEndpoint,
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

// AddRecord adds a record.
func (c *Client) AddRecord(ctx context.Context, record Record) (*Record, error) {
	data := APIRequest{
		Method: "add-record",
		Params: record,
	}

	req, err := newJSONRequest(ctx, http.MethodPost, c.apiEndpoint, data)
	if err != nil {
		return nil, err
	}

	var result APIResponse[*Record]
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

// RemoveRecord removes a record.
func (c *Client) RemoveRecord(ctx context.Context, id string, domain string) error {
	data := APIRequest{
		Method: "remove-record",
		Params: Record{
			ID:     id,
			Domain: domain,
		},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, c.apiEndpoint, data)
	if err != nil {
		return err
	}

	err = c.do(req, &APIResponse[json.RawMessage]{})
	if err != nil {
		return err
	}

	return nil
}

// ListRecords list the records for one domain.
func (c *Client) ListRecords(ctx context.Context, domain string) ([]Record, error) {
	data := APIRequest{
		Method: "list-records",
		Params: Record{
			Domain: domain,
		},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, c.apiEndpoint, data)
	if err != nil {
		return nil, err
	}

	var result APIResponse[Records]
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Result.Records, nil
}

func (c *Client) do(req *http.Request, result Response) error {
	req.Header.Set(authorizationHeader, "Njalla "+c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return result.GetError()
}

func newJSONRequest(ctx context.Context, method string, endpoint string, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
