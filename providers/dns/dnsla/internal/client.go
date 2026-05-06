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
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.dns.la"

const (
	TypeA             = 1
	TypeNS            = 2
	TypeCNAME         = 5
	TypeMX            = 15
	TypeTXT           = 16
	TypeAAAA          = 28
	TypeSRV           = 33
	TypeCAA           = 257
	TypeURLForwarding = 256
)

// Client the dns.la API client.
type Client struct {
	apiID     string
	apiSecret string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiID, apiSecret string) (*Client, error) {
	if apiID == "" || apiSecret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiID:      apiID,
		apiSecret:  apiSecret,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) ListDomains(ctx context.Context, pager Pager) ([]Domain, error) {
	endpoint := c.BaseURL.JoinPath("api", "domainList")

	values, err := querystring.Values(pager)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	result := BaseResponse[Domains]{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("code: %d, msg: %s", result.Code, result.Msg)
	}

	return result.Data.Results, nil
}

func (c *Client) AddRecord(ctx context.Context, record Record) (string, error) {
	endpoint := c.BaseURL.JoinPath("api", "record")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return "", err
	}

	result := BaseResponse[RecordID]{}

	err = c.do(req, &result)
	if err != nil {
		return "", err
	}

	if result.Code != 200 {
		return "", fmt.Errorf("code: %d, msg: %s", result.Code, result.Msg)
	}

	return result.Data.ID, nil
}

func (c *Client) DeleteRecord(ctx context.Context, recordID string) error {
	endpoint := c.BaseURL.JoinPath("api", "record")

	query := endpoint.Query()
	query.Set("id", recordID)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, http.NoBody)
	if err != nil {
		return err
	}

	result := BaseResponse[any]{}

	err = c.do(req, &result)
	if err != nil {
		return err
	}

	if result.Code != 200 {
		return fmt.Errorf("code: %d, msg: %s", result.Code, result.Msg)
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth(c.apiID, c.apiSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
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
