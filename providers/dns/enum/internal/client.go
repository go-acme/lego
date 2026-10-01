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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://api.enum.co"

const authorizationHeader = "Authorization"

// Client the Enum API client.
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

func (c *Client) GetZoneByName(ctx context.Context, payload *GetZoneByNameRequest) (*GetZoneByNameResponse, error) {
	endpoint := c.BaseURL.JoinPath("enum.api.v1.DnsService", "GetZoneByName")

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}

	resp := &GetZoneByNameResponse{}

	err = c.do(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListProjects(ctx context.Context) (*ListProjectsResponse, error) {
	endpoint := c.BaseURL.JoinPath("enum.api.v1.ProjectService", "ListProjects")

	req, err := newJSONRequest(ctx, endpoint, struct{}{})
	if err != nil {
		return nil, err
	}

	resp := &ListProjectsResponse{}

	err = c.do(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) AddRecordSetValue(ctx context.Context, payload *AddRecordSetValueRequest) (*RecordSetResponse, error) {
	endpoint := c.BaseURL.JoinPath("enum.api.v1.DnsService", "AddRecordSetValue")

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}

	resp := &RecordSetResponse{}

	err = c.do(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) RemoveRecordSetValue(ctx context.Context, payload *RemoveRecordSetValueRequest) (*RecordSetResponse, error) {
	endpoint := c.BaseURL.JoinPath("enum.api.v1.DnsService", "RemoveRecordSetValue")

	req, err := newJSONRequest(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}

	resp := &RecordSetResponse{}

	err = c.do(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
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

func newJSONRequest(ctx context.Context, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), buf)
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
