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

const defaultBaseURL = "https://api.rackcorp.net/api/v2.8"

// Client the RackCorp API client.
type Client struct {
	apiUUID   string
	apiSecret string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiUUID, apiSecret string) (*Client, error) {
	if apiUUID == "" || apiSecret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiUUID:    apiUUID,
		apiSecret:  apiSecret,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a new DNS record.
// https://app.swaggerhub.com/apis/RackCorp/Rackcorp-REST-API/2.8#/default/dns.record.create
func (c *Client) CreateRecord(ctx context.Context, record Record) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", "records")

	payload := map[string]any{
		"data": []Record{record},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var result []Record

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response data: %s, %w", string(data), err)
	}

	return result, nil
}

// DeleteRecord deletes a DNS record.
// https://app.swaggerhub.com/apis/RackCorp/Rackcorp-REST-API/2.8#/default/dns.record.delete
func (c *Client) DeleteRecord(ctx context.Context, recordID int64) error {
	endpoint := c.BaseURL.JoinPath("dns", "records", strconv.FormatInt(recordID, 10))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// GetDomains gets all domain details.
// https://app.swaggerhub.com/apis/RackCorp/Rackcorp-REST-API/2.8#/default/dns.domain.getall
func (c *Client) GetDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.BaseURL.JoinPath("dns", "domain")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	data, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var result []Domain

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response data: %s, %w", string(data), err)
	}

	return result, nil
}

func (c *Client) do(req *http.Request) (json.RawMessage, error) {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth(c.apiUUID, c.apiSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return nil, errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	result := &APIResponse{}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if result.Code != "OK" {
		return nil, &APIError{Code: result.Code, Message: result.Message}
	}

	return result.Data, nil
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
