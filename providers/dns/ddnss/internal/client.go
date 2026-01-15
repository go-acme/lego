package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	"golang.org/x/net/html"
)

const defaultBaseURL = "https://ddnss.de/upd.php"

// Client the DDns API client.
type Client struct {
	auth *Authentication

	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(auth *Authentication) (*Client, error) {
	if auth == nil {
		return nil, errors.New("credentials missing")
	}

	err := auth.validate()
	if err != nil {
		return nil, err
	}

	return &Client{
		auth:       auth,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, host, value string) error {
	return c.update(ctx, map[string]string{
		"host": host,
		"txt":  value,
		"txtm": "1",
	})
}

func (c *Client) RemoveTXTRecord(ctx context.Context, host string) error {
	return c.update(ctx, map[string]string{
		"host": host,
		"txtm": "2",
	})
}

func (c *Client) update(ctx context.Context, params map[string]string) error {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	query := endpoint.Query()

	for k, v := range params {
		query.Set(k, v)
	}

	c.auth.set(query)

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

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

	content, err := readPage(raw)
	if err != nil {
		return err
	}

	if strings.Contains(content, "Updated 1 hostname.") {
		return nil
	}

	return fmt.Errorf("unexpected response: %s", content)
}

func readPage(raw []byte) (string, error) {
	page, err := html.Parse(strings.NewReader(string(raw)))
	if err != nil {
		return "", err
	}

	var b strings.Builder
	extractText(page, &b)

	return strings.TrimSpace(b.String()), nil
}

func extractText(n *html.Node, b *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			b.WriteString(text + " ")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, b)
	}
}
