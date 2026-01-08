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
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://api.noip.com"

// Client the No-IP API client.
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		BaseURL:    baseURL,
		httpClient: hc,
	}, nil
}

// CreateRData creates a record.
// https://developer.noip.com/reference/v1-dns-records-create-rdata
func (c *Client) CreateRData(ctx context.Context, zone, name, dnsType string, data RData) error {
	endpoint := c.BaseURL.JoinPath("v1", "dns", "records", zone, name, "rrsets", dnsType, "rdata")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, []RData{data})
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteRDataByLabel deletes a record by label.
// https://developer.noip.com/reference/v1-dns-records-delete-rdata-by-label
func (c *Client) DeleteRDataByLabel(ctx context.Context, zone, name, dnsType, label string) error {
	endpoint := c.BaseURL.JoinPath("v1", "dns", "records", zone, name, "rrsets", dnsType, "rdata", label)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.httpClient.Do(req)
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

	var apiResp APIResponse[any]

	err := json.Unmarshal(raw, &apiResp)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, apiResp.Errors)
}

func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
