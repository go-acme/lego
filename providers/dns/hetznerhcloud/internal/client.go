package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const DefaultBaseURL = "https://api.hetzner.cloud"

type Client struct {
	token      string
	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(token string) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)
	return &Client{
		token:      token,
		baseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) SetBaseURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	c.baseURL = u
	return nil
}

func (c *Client) GetZone(ctx context.Context, name string) (string, error) {
	query := url.Values{"name": {name}}
	var resp struct {
		Zones []struct {
			ID   json.RawMessage `json:"id"`
			Name string          `json:"name"`
		} `json:"zones"`
	}
	if err := c.get(ctx, "/v1/zones", query, &resp); err != nil {
		return "", err
	}
	for _, z := range resp.Zones {
		if z.Name == name {
			return string(bytes.Trim(z.ID, "\"")), nil
		}
	}
	return "", fmt.Errorf("zone %q not found", name)
}

func (c *Client) CreateRecord(ctx context.Context, zoneID, name, value string, ttl int) (string, error) {
	payload := map[string]any{"name": name, "type": "TXT", "value": value, "ttl": ttl}
	var resp struct {
		Record struct {
			ID json.RawMessage `json:"id"`
		} `json:"record"`
	}
	if err := c.post(ctx, fmt.Sprintf("/v1/zones/%s/records", zoneID), payload, &resp); err != nil {
		return "", err
	}
	return string(bytes.Trim(resp.Record.ID, "\"")), nil
}

func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	return c.delete(ctx, fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID))
}

func (c *Client) get(ctx context.Context, path string, query url.Values, result any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, result)
}

func (c *Client) post(ctx context.Context, path string, payload, result any) error {
	body, _ := json.Marshal(payload)
	return c.do(ctx, http.MethodPost, path, nil, body, result)
}

func (c *Client) delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body []byte, result any) error {
	u := c.baseURL.ResolveReference(&url.URL{Path: path, RawQuery: query.Encode()})
	req, _ := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	if result != nil && len(data) > 0 {
		return json.Unmarshal(data, result)
	}
	return nil
}
