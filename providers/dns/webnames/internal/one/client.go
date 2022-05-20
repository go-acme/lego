package one

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://www.webnames.ru/scripts/json_domain_zone_manager.pl"

// Client the Webnames client.
type Client struct {
	apikey string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient Creates a Webnames client.
func NewClient(apikey string) *Client {
	return &Client{
		apikey:     apikey,
		baseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// RemoveTxtRecord removes a TXT record.
// Inspired by https://github.com/regtime-ltd/certbot-dns-webnames/blob/master/cleanup.sh
func (c Client) RemoveTxtRecord(domain, subDomain, content string) error {
	data := url.Values{}
	data.Set("domain", domain)
	data.Set("type", "TXT")
	data.Set("record", fmt.Sprintf("%s:%s", subDomain, content))
	data.Set("action", "delete")

	return c.do(data)
}

// AddTXTRecord adds a TXT record.
// Inspired by https://github.com/regtime-ltd/certbot-dns-webnames/blob/master/authenticator.sh
func (c Client) AddTXTRecord(domain, subDomain, content string) error {
	data := url.Values{}
	data.Set("domain", domain)
	data.Set("type", "TXT")
	data.Set("record", fmt.Sprintf("%s:%s", subDomain, content))
	data.Set("action", "add")

	return c.do(data)
}

func (c Client) do(data url.Values) error {
	data.Set("apikey", c.apikey)

	req, err := http.NewRequest(http.MethodPost, c.baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		all, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status code: %d: %s", resp.StatusCode, string(all))
	}

	var r APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if r.Result == "OK" {
		return nil
	}

	return fmt.Errorf("%s: %s", r.Result, r.Details)
}
