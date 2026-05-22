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
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
)

const defaultBaseURL = "https://api.dynadot.com"

// Client the Dynadot RESTful v2 API client.
type Client struct {
	apiKey    string
	apiSecret string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, apiSecret string) (*Client, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// SetDNS adds DNS records for the specified domain.
// https://www.dynadot.com/domain/api-document?api-version=2.0.0#set_dns
func (c *Client) SetDNS(ctx context.Context, domain string, payload *SetDNSRequest) error {
	endpoint := c.BaseURL.JoinPath("restful", "v2", "domains", domain, "records")

	req, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req)
}

// RemoveDNS removes DNS records for the specified domain.
// Currently not documented.
func (c *Client) RemoveDNS(ctx context.Context, domain string, payload *RemoveDNSRequest) error {
	endpoint := c.BaseURL.JoinPath("restful", "v2", "domains", domain, "records")

	req, err := c.newJSONRequest(ctx, http.MethodDelete, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req)
}

func (c *Client) do(req *http.Request) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	envelope := new(APIResponse)

	err = json.Unmarshal(raw, envelope)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if resp.StatusCode/100 != 2 || (envelope.Code != 0 && envelope.Code != 200) {
		return envelope
	}

	return nil
}

func (c *Client) newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
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

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Signature", c.sign(req, "", buf.String()))

	return req, nil
}
