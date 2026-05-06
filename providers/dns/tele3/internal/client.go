package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://www.tele3.cz/acme/"

// Client the Tele3 API client.
type Client struct {
	key    string
	secret string

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(key, secret string) (*Client, error) {
	if key == "" || secret == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		key:        key,
		secret:     secret,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, domain, value string) error {
	return c.apply(ctx, "add", domain, value)
}

func (c *Client) RemoveRecord(ctx context.Context, domain, value string) error {
	return c.apply(ctx, "rm", domain, value)
}

func (c *Client) apply(ctx context.Context, operation, domain, value string) error {
	ope := Operation{
		Key:       c.key,
		Secret:    c.secret,
		Operation: operation,
		Domain:    domain,
		Value:     value,
	}

	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(ope)
	if err != nil {
		return fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, buf)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	return c.do(req)
}

func (c *Client) do(req *http.Request) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if string(raw) != "success" {
		return fmt.Errorf("unexpected response: %s", string(raw))
	}

	return nil
}
