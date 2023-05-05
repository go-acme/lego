package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const (
	removeAction = "rm"
	addAction    = "add"
)

const successCode = "successfully"

const defaultBaseURL = "https://www.dnshome.de/dyndns.php"

// Client the dnsHome.de client.
type Client struct {
	baseURL    string
	HTTPClient *http.Client

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
func (c *Client) Add(ctx context.Context, hostname, value string) error {
	domain := strings.TrimPrefix(hostname, "_acme-challenge.")

	return c.doAction(ctx, domain, addAction, value)
}

// Remove removes a TXT record.
// only one TXT record for ACME is allowed, so it will remove "all" the TXT records.
func (c *Client) Remove(ctx context.Context, hostname, value string) error {
	domain := strings.TrimPrefix(hostname, "_acme-challenge.")

	return c.doAction(ctx, domain, removeAction, value)
}

func (c *Client) doAction(ctx context.Context, domain, action, value string) error {
	endpoint, err := c.createEndpoint(domain, action, value)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	output := string(raw)

	if !strings.HasPrefix(output, successCode) {
		return errors.New(output)
	}

	return nil
}

func (c *Client) createEndpoint(domain, action, value string) (*url.URL, error) {
	if len(value) < 12 {
		return nil, fmt.Errorf("the TXT value must have more than 12 characters: %s", value)
	}

	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	c.credMu.Lock()
	password, ok := c.credentials[domain]
	c.credMu.Unlock()

	if !ok {
		return nil, fmt.Errorf("domain %s not found in credentials, check your credentials map", domain)
	}

	endpoint.User = url.UserPassword(domain, password)

	query := endpoint.Query()
	query.Set("acme", action)
	query.Set("txt", value)
	endpoint.RawQuery = query.Encode()

	return endpoint, nil
}
