package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://ipv64.net"

type Client struct {
	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{
		baseURL:    baseURL,
		HTTPClient: hc,
	}
}

func (c *Client) GetDomains(ctx context.Context) (*Domains, error) {
	endpoint := c.baseURL.JoinPath("api")

	query := endpoint.Query()
	query.Set("get_domains", "")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	results := &Domains{}

	err = c.do(req, results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (c *Client) AddRecord(ctx context.Context, domain, prefix, recordType, content string) error {
	endpoint := c.baseURL.JoinPath("api")

	data := make(url.Values)
	data.Set("add_record", domain)
	data.Set("praefix", prefix)
	data.Set("type", recordType)
	data.Set("content", content)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	return c.do(req, nil)
}

func (c *Client) DeleteRecord(ctx context.Context, domain, prefix, recordType, content string) error {
	endpoint := c.baseURL.JoinPath("api")

	data := make(url.Values)
	data.Set("del_record", domain)
	data.Set("praefix", prefix)
	data.Set("type", recordType)
	data.Set("content", content)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	if req.Method != http.MethodGet {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

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

	if string(raw) == "null" {
		return fmt.Errorf("unexpected response: %s", string(raw))
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errAPI := &APIError{}

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}

func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
