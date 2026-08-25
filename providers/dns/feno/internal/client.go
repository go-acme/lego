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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

// defaultBaseURL the default API endpoint.
const defaultBaseURL = "https://api.feno.no/v1"

// Client the FENO API client.
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
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetDomain gets one domain.
// https://github.com/mrerikcodes/feno-api/blob/main/openapi.yaml
func (c *Client) GetDomain(ctx context.Context, domain string) (*Domain, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[*Domain])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// ListRecords lists the DNS records of a zone.
// An `acme:write` key only sees the TXT records at `_acme-challenge*`.
// https://github.com/mrerikcodes/feno-api/blob/main/openapi.yaml
func (c *Client) ListRecords(ctx context.Context, zone string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", zone, "dns")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[*RecordList])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	if result.Data == nil {
		return nil, nil
	}

	return result.Data.Records, nil
}

// CreateRecord creates a DNS record inside a zone.
// https://github.com/mrerikcodes/feno-api/blob/main/openapi.yaml
func (c *Client) CreateRecord(ctx context.Context, zone string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", zone, "dns")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[*Record])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	if result.Data == nil || result.Data.ID == 0 {
		return nil, errors.New("response carried no record id")
	}

	return result.Data, nil
}

// DeleteRecord deletes a DNS record from a zone.
// https://github.com/mrerikcodes/feno-api/blob/main/openapi.yaml
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordID int) error {
	endpoint := c.BaseURL.JoinPath("domains", zone, "dns", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	var envelope APIResponse[json.RawMessage]

	err = json.Unmarshal(raw, &envelope)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if !envelope.Success {
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       envelope.Code,
			Message:    envelope.Error,
			RequestID:  resp.Header.Get("X-Request-Id"),
		}
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

	var envelope APIResponse[json.RawMessage]

	err := json.Unmarshal(raw, &envelope)
	if err != nil || envelope.Error == "" {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Code:       envelope.Code,
		Message:    envelope.Error,
		RequestID:  resp.Header.Get("X-Request-Id"),
		RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
	}
}

// parseRetryAfter parses a Retry-After header (delta-seconds or HTTP-date).
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	if secs, err := strconv.Atoi(value); err == nil {
		return max(time.Duration(secs)*time.Second, 0)
	}

	if t, err := http.ParseTime(value); err == nil {
		return max(time.Until(t), 0)
	}

	return 0
}
