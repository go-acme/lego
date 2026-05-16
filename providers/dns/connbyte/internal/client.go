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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://api.connbyte.com"

// Client the Connbyte API client.
type Client struct {
	token string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) ListZones(ctx context.Context) ([]ZoneListResponse, error) {
	endpoint := c.BaseURL.JoinPath("ns", "zone", "list")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[[]ZoneListResponse])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) CreateRecord(ctx context.Context, record Record) (*RecordCreateResponse, error) {
	endpoint := c.BaseURL.JoinPath("ns", "record", "create")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[*RecordCreateResponse])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) DeleteRecord(ctx context.Context, zoneID string, recordID int64) error {
	endpoint := c.BaseURL.JoinPath("ns", "record", "delete")

	payload := map[string]any{
		"zoneId":   zoneID,
		"recordId": recordID,
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "Bearer "+c.token)

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

	result := new(APIError)

	err := json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return result
}
