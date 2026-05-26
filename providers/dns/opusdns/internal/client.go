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

const AuthenticationHeader = "X-Api-Key"

const (
	RecordOperationUpset  = "upsert"
	RecordOperationRemove = "remove"
)

const defaultBaseURL = "https://api.opusdns.com"

// Client the OpusDNS API client.
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

// PatchRecords applies record operations to a zone.
// https://developers.opusdns.com/products/dns/manage-records#add-a-single-record
// https://developers.opusdns.com/products/dns/manage-records#remove-a-single-record
func (c *Client) PatchRecords(ctx context.Context, zone string, ops []RecordOperation) error {
	endpoint := c.BaseURL.JoinPath("v1", "dns", zone, "records")

	payload := PatchRequest{Ops: ops}

	req, err := newJSONRequest(ctx, http.MethodPatch, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req)
}

func (c *Client) do(req *http.Request) error {
	useragent.SetHeader(req.Header)

	req.Header.Set(AuthenticationHeader, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
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

	apiErr := &APIError{StatusCode: resp.StatusCode}

	err := json.Unmarshal(raw, apiErr)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return apiErr
}
