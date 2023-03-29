package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.beget.com/api/"

// Client the beget.com client.
type Client struct {
	login    string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a beget.com client.
func NewClient(login, password string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		login:      login,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddTXTRecord adds a TXT record.
// https://beget.com/ru/kb/api/funkczii-upravleniya-dns
func (c Client) AddTXTRecord(ctx context.Context, domain, subDomain, content string) error {
	request := ChangeRecordsRequest{
		Fqdn: fmt.Sprintf("%s.%s", subDomain, domain),
		Records: RecordList{
			TXT: []Record{
				{Priority: 10, Value: content},
			},
		},
	}

	resp, err := c.doRequest(ctx, request, "dns", "changeRecords")
	if err != nil {
		return err
	}

	return resp.HasError()
}

// RemoveTxtRecord removes a TXT record.
// https://beget.com/ru/kb/api/funkczii-upravleniya-dns
func (c Client) RemoveTxtRecord(ctx context.Context, domain, subDomain string) error {
	request := ChangeRecordsRequest{
		Fqdn:    fmt.Sprintf("%s.%s", subDomain, domain),
		Records: RecordList{},
	}

	resp, err := c.doRequest(ctx, request, "dns", "changeRecords")
	if err != nil {
		return err
	}

	return resp.HasError()
}

func (c Client) doRequest(ctx context.Context, data any, fragments ...string) (*APIResponse, error) {
	endpoint := c.baseURL.JoinPath(fragments...)

	inputData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to mashall input data: %w", err)
	}

	query := endpoint.Query()
	query.Add("input_data", string(inputData))
	query.Add("login", c.login)
	query.Add("passwd", c.password)
	query.Add("input_format", "json")
	query.Add("output_format", "json")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return nil, parseError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var apiResp APIResponse
	err = json.Unmarshal(raw, &apiResp)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return &apiResp, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var apiResp APIResponse
	err := json.Unmarshal(raw, &apiResp)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, apiResp)
}
