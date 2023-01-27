package internal

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	removeAction = "rm"
	addAction    = "add"
)

const successCode = "successfully"

const defaultBaseURL = "https://www.dnshome.de/dyndns.php"

// Client the dnsHome.de client.
type Client struct {
	HTTPClient *http.Client
	baseURL    string

	credentials map[string]string
	credMu      sync.Mutex
}

// NewClient Creates a new Client.
func NewClient(credentials map[string]string) *Client {
	return &Client{
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
		baseURL:     defaultBaseURL,
		credentials: credentials,
	}
}

// Add adds a TXT record.
// only one TXT record for ACME is allowed, so it will update the "current" TXT record.
func (c *Client) Add(hostname, value string) error {
	domain := strings.TrimPrefix(hostname, "_acme-challenge.")

	c.credMu.Lock()
	password, ok := c.credentials[domain]
	c.credMu.Unlock()

	if !ok {
		return fmt.Errorf("domain %s not found in credentials, check your credentials map", domain)
	}

	return c.do(url.UserPassword(domain, password), addAction, value)
}

// Remove removes a TXT record.
// only one TXT record for ACME is allowed, so it will remove "all" the TXT records.
func (c *Client) Remove(hostname, value string) error {
	domain := strings.TrimPrefix(hostname, "_acme-challenge.")

	c.credMu.Lock()
	password, ok := c.credentials[domain]
	c.credMu.Unlock()

	if !ok {
		return fmt.Errorf("domain %s not found in credentials, check your credentials map", domain)
	}

	return c.do(url.UserPassword(domain, password), removeAction, value)
}

func (c *Client) do(userInfo *url.Userinfo, action, value string) error {
	if len(value) < 12 {
		return fmt.Errorf("the TXT value must have more than 12 characters: %s", value)
	}

	apiEndpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}

	apiEndpoint.User = userInfo

	query := apiEndpoint.Query()
	query.Set("acme", action)
	query.Set("txt", value)
	apiEndpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodPost, apiEndpoint.String(), http.NoBody)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		all, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%d: %s", resp.StatusCode, string(all))
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	output := string(all)

	if !strings.HasPrefix(output, successCode) {
		return errors.New(output)
	}

	return nil
}
