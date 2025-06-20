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
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.derak.cloud/v1.0"

type Client struct {
	baseURL      *url.URL
	HTTPClient   *http.Client
	zoneEndpoint string

	apiKey string
}

func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
		baseURL:      baseURL,
		zoneEndpoint: "https://api.derak.cloud/api/v2/service/cdn/zones",
		apiKey:       apiKey,
	}
}

// GetRecords gets all records.
// Note: the response is not influenced by the query parameters, so the documentation seems wrong.
func (c Client) GetRecords(ctx context.Context, zoneID string, params *GetRecordsParameters) (*GetRecordsResponse, error) {
	endpoint := c.baseURL.JoinPath("zones", zoneID, "dnsrecords")

	v, err := querystring.Values(params)
	if err != nil {
		return nil, err
	}
	endpoint.RawQuery = v.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &GetRecordsResponse{}
	err = c.do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetRecord gets a record by ID.
func (c Client) GetRecord(ctx context.Context, zoneID, recordID string) (*Record, error) {
	endpoint := c.baseURL.JoinPath("zones", zoneID, "dnsrecords", recordID)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &Record{}
	err = c.do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// CreateRecord creates a new record.
func (c Client) CreateRecord(ctx context.Context, zoneID string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("zones", zoneID, "dnsrecords")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, record)
	if err != nil {
		return nil, err
	}

	response := &Record{}
	err = c.do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// EditRecord edits an existing record.
func (c Client) EditRecord(ctx context.Context, zoneID, recordID string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("zones", zoneID, "dnsrecords", recordID)

	req, err := newJSONRequest(ctx, http.MethodPatch, endpoint, record)
	if err != nil {
		return nil, err
	}

	response := &Record{}
	err = c.do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// DeleteRecord deletes an existing record.
func (c Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := c.baseURL.JoinPath("zones", zoneID, "dnsrecords", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	response := &APIResponse[any]{}

	err = c.do(req, response)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("API error: %d %s", response.Error, codeText(response.Error))
	}

	return nil
}

// GetZones gets zones.
// Note: it's not a part of the official API, there is no documentation about this.
// The endpoint comes from UI calls analysis.
func (c Client) GetZones(ctx context.Context) ([]Zone, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.zoneEndpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	response := &APIResponse[[]Zone]{}
	err = c.do(req, response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %d %s", response.Error, codeText(response.Error))
	}

	return response.Result, nil
}

func (c Client) do(req *http.Request, result any) error {
	req.SetBasicAuth("api", c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch req.Method {
	case http.MethodPut:
		if resp.StatusCode != http.StatusCreated {
			return parseError(req, resp)
		}
	default:
		if resp.StatusCode != http.StatusOK {
			return parseError(req, resp)
		}
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

	var response APIResponse[any]
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %d: %s", resp.StatusCode, response.Error, codeText(response.Error))
}
