package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://simple-api.dondominio.net"

// Client the DonDominio/MrDomain API client.
type Client struct {
	apiUser     string
	apiPassword string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiUser, apiPassword string) (*Client, error) {
	if apiUser == "" || apiPassword == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiUser:     apiUser,
		apiPassword: apiPassword,
		BaseURL:     baseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a DNS record.
// https://dondominio.dev/en/api/docs/api/#dns-zone-create-service-dnscreate
func (c *Client) CreateRecord(ctx context.Context, payload CreateRecordRequest) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("service", "dnscreate")

	values, err := querystring.Values(payload)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, endpoint, values)
	if err != nil {
		return nil, err
	}

	var result DNSResponse

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.DNS, nil
}

// DeleteRecord deletes a DNS record.
// https://dondominio.dev/en/api/docs/api/#dns-zone-delete-service-dnsdelete
func (c *Client) DeleteRecord(ctx context.Context, payload DeleteRecordRequest) error {
	endpoint := c.BaseURL.JoinPath("service", "dnsdelete")

	values, err := querystring.Values(payload)
	if err != nil {
		return err
	}

	req, err := c.newRequest(ctx, endpoint, values)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

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

	response := APIResponse{}

	err = json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if !response.Success {
		return fmt.Errorf("%d: %s: %s: %s", response.ErrorCode, response.Action, response.ErrorCodeMsg, strings.Join(response.Messages, ", "))
	}

	err = json.Unmarshal(response.Data, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, response.Data, err)
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, endpoint *url.URL, values url.Values) (*http.Request, error) {
	values.Set("apiuser", c.apiUser)
	values.Set("apipasswd", c.apiPassword)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}
