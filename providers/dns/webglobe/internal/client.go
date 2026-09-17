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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://api.webglobe.com/"

const authorizationHeader = "Authorization"

// Client the Webglobe API client.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient() (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) GetDomainDetails(ctx context.Context, domainID string) (*Domain, error) {
	endpoint := c.BaseURL.JoinPath("domains", domainID)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[*Domain])

	err = c.do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("failure: %s", result.Status)
	}

	return result.Data, nil
}

func (c *Client) CreateRecord(ctx context.Context, domainID int64, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath(strconv.FormatInt(domainID, 10), "dns")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(RecordResponse)

	err = c.do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int64) error {
	endpoint := c.BaseURL.JoinPath(strconv.FormatInt(domainID, 10), "dns", strconv.FormatInt(recordID, 10))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, nil)
}

func (c *Client) do(ctx context.Context, req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set(authorizationHeader, "Bearer "+getToken(ctx))

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

	var envelope APIResponse[any]

	err := json.Unmarshal(raw, &envelope)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return envelope.Error
}
