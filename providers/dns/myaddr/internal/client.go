package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://myaddr.tools"

// Client the myaddr.{tools,dev,io} API client.
type Client struct {
	baseURL    *url.URL
	HTTPClient *http.Client

	credentials map[string]string
	credMu      sync.Mutex
}

// NewClient creates a new Client.
func NewClient(credentials map[string]string) (*Client, error) {
	if len(credentials) == 0 {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		baseURL:     baseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
		credentials: credentials,
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, subdomain, value string) error {
	c.credMu.Lock()
	privateKey, ok := c.credentials[subdomain]
	c.credMu.Unlock()

	if !ok {
		return fmt.Errorf("subdomain %s not found in credentials, check your credentials map", subdomain)
	}

	payload := ACMEChallenge{Key: privateKey, Data: value}

	req, err := newJSONRequest(ctx, http.MethodPost, c.baseURL.JoinPath("update"), payload)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
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
