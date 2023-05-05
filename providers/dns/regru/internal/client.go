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

const defaultBaseURL = "https://api.reg.ru/api/regru2/"

// Client the reg.ru client.
type Client struct {
	username string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a reg.ru client.
func NewClient(username, password string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username:   username,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// RemoveTxtRecord removes a TXT record.
// https://www.reg.ru/support/help/api2#zone_remove_record
func (c Client) RemoveTxtRecord(ctx context.Context, domain, subDomain, content string) error {
	request := RemoveRecordRequest{
		Username:          c.username,
		Password:          c.password,
		Domains:           []Domain{{DName: domain}},
		SubDomain:         subDomain,
		Content:           content,
		RecordType:        "TXT",
		OutputContentType: "plain",
	}

	resp, err := c.doRequest(ctx, request, "zone", "remove_record")
	if err != nil {
		return err
	}

	return resp.HasError()
}

// AddTXTRecord adds a TXT record.
// https://www.reg.ru/support/help/api2#zone_add_txt
func (c Client) AddTXTRecord(ctx context.Context, domain, subDomain, content string) error {
	request := AddTxtRequest{
		Username:          c.username,
		Password:          c.password,
		Domains:           []Domain{{DName: domain}},
		SubDomain:         subDomain,
		Text:              content,
		OutputContentType: "plain",
	}

	resp, err := c.doRequest(ctx, request, "zone", "add_txt")
	if err != nil {
		return err
	}

	return resp.HasError()
}

func (c Client) doRequest(ctx context.Context, request any, fragments ...string) (*APIResponse, error) {
	endpoint := c.baseURL.JoinPath(fragments...)

	inputData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to create input data: %w", err)
	}

	query := endpoint.Query()
	query.Add("input_data", string(inputData))
	query.Add("input_format", "json")
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

	var errAPI APIResponse
	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("status code: %d, %w", resp.StatusCode, errAPI)
}
