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
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetDomain gets a domain.
// https://github.com/MrErikCodes/feno-api/blob/a28ccb0739cc42433095b14ea2c8774597dcf835/openapi.yaml#L201
func (c *Client) GetDomain(ctx context.Context, domain string) (*Domain, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(Domain)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListRecords lists the DNS records of a zone.
// https://github.com/MrErikCodes/feno-api/blob/a28ccb0739cc42433095b14ea2c8774597dcf835/openapi.yaml#L481
func (c *Client) ListRecords(ctx context.Context, zone string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", zone, "dns")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(RecordList)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Records, nil
}

// CreateRecord creates a new DNS record.
// https://github.com/MrErikCodes/feno-api/blob/a28ccb0739cc42433095b14ea2c8774597dcf835/openapi.yaml#L549
func (c *Client) CreateRecord(ctx context.Context, zone string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", zone, "dns")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(Record)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	if result.ID == 0 {
		return nil, errors.New("response carried no record id")
	}

	return result, nil
}

// DeleteRecord deletes a DNS record.
// https://github.com/MrErikCodes/feno-api/blob/a28ccb0739cc42433095b14ea2c8774597dcf835/openapi.yaml#L683
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

	res := new(APIResponse)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, res)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if !res.Success {
		return NewAPIError(resp.StatusCode, req.Header, res)
	}

	if result == nil {
		return nil
	}

	err = json.Unmarshal(res.Data, result)
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

	res := new(APIResponse)

	err := json.Unmarshal(raw, res)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return NewAPIError(resp.StatusCode, req.Header, res)
}
