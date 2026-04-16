package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is an HTTP client for the DNScale API.
type Client struct {
	baseURL    string
	apiToken   string
	HTTPClient *http.Client
}

// NewClient creates a new DNScale API client.
func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiToken:   apiToken,
		HTTPClient: &http.Client{},
	}
}

// FindZoneByFQDN finds the zone that manages the given FQDN.
// It strips labels from the left until a matching zone is found.
// Returns the zone ID, zone name, and any error.
func (c *Client) FindZoneByFQDN(ctx context.Context, fqdn string) (string, string, error) {
	zones, err := c.listZones(ctx)
	if err != nil {
		return "", "", fmt.Errorf("list zones: %w", err)
	}

	// Strip the trailing dot from the FQDN.
	name := strings.TrimSuffix(fqdn, ".")

	// Walk up the domain labels to find the zone.
	for {
		for _, z := range zones {
			zoneName := strings.TrimSuffix(z.Name, ".")
			if strings.EqualFold(zoneName, name) {
				return z.ID, z.Name, nil
			}
		}

		// Remove the leftmost label.
		idx := strings.Index(name, ".")
		if idx < 0 {
			break
		}
		name = name[idx+1:]
	}

	return "", "", fmt.Errorf("no zone found for %s", fqdn)
}

// CreateTXTRecord creates a TXT record in the given zone.
func (c *Client) CreateTXTRecord(ctx context.Context, zoneID, name, value string, ttl int) error {
	payload := RecordRequest{
		Name:    name,
		Type:    "TXT",
		Content: value,
		TTL:     ttl,
	}

	endpoint := fmt.Sprintf("%s/v1/zones/%s/records", c.baseURL, zoneID)

	req, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// DeleteTXTRecord deletes a specific TXT record by name and content value.
func (c *Client) DeleteTXTRecord(ctx context.Context, zoneID, name, value string) error {
	endpoint := fmt.Sprintf("%s/v1/zones/%s/records/by-name/%s/TXT?content=%s",
		c.baseURL, zoneID, name, url.QueryEscape(value))

	req, err := c.newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) listZones(ctx context.Context) ([]Zone, error) {
	var allZones []Zone
	offset := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("%s/v1/zones?offset=%d&limit=%d", c.baseURL, offset, limit)

		req, err := c.newJSONRequest(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		var result ZonesResponse
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		allZones = append(allZones, result.Data.Zones...)

		if len(result.Data.Zones) < limit {
			break
		}
		offset += limit
	}

	return allZones, nil
}

func (c *Client) newJSONRequest(ctx context.Context, method, endpoint string, payload any) (*http.Request, error) {
	var body io.Reader

	if payload != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiErr APIError
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Code != "" {
		return fmt.Errorf("API error %d: %s - %s", resp.StatusCode, apiErr.Error.Code, apiErr.Error.Message)
	}

	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}
