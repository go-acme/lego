package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

const ddnsScriptPath = "/ddns/update.php"

type Client struct {
	endpoint   *url.URL
	token      string
	httpClient *http.Client
}

func NewClient(endpoint string, token string, httpClient *http.Client) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, ddnsScriptPath)

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		endpoint:   u,
		token:      token,
		httpClient: httpClient,
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, zone, recordFQDN, recordContent string) error {
	return c.do(ctx, "add", zone, recordFQDN, recordContent)
}

func (c *Client) DeleteRecord(ctx context.Context, zone, recordFQDN, recordContent string) error {
	return c.do(ctx, "delete", zone, recordFQDN, recordContent)
}

func (c *Client) do(ctx context.Context, action, zone, recordFQDN, recordContent string) error {
	query := url.Values{}
	query.Set("action", action)
	query.Set("zone", zone)
	query.Set("type", "TXT")
	query.Set("record", recordFQDN)
	query.Set("data", recordContent)

	reqURL := *c.endpoint
	reqURL.RawQuery = query.Encode()

	method := http.MethodPost
	if action == "delete" {
		method = http.MethodDelete
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth("anonymous", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
