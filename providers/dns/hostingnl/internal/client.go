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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultBaseURL = "https://api.hosting.nl"

type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c Client) AddRecord(ctx context.Context, domain string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain, "dns")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, []Record{record})
	if err != nil {
		return nil, err
	}

	var result APIResponse[Record]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Data) != 1 {
		return nil, fmt.Errorf("unexpected response data: %v", result.Data)
	}

	return &result.Data[0], nil
}

func (c Client) DeleteRecord(ctx context.Context, domain, recordID string) error {
	endpoint := c.BaseURL.JoinPath("domains", domain, "dns")

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, []Record{{ID: recordID}})
	if err != nil {
		return err
	}

	var result APIResponse[Record]

	err = c.do(req, &result)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("API-TOKEN", c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
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

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var apiErr APIError

	err := json.Unmarshal(raw, &apiErr)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, apiErr)
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
