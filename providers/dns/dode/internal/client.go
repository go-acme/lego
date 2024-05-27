package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://my.do.de/api"

// Client the do.de API client.
type Client struct {
	token string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(token string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// UpdateTxtRecord Update the domains TXT record
// To update the TXT record we just need to make one simple get request.
func (c Client) UpdateTxtRecord(ctx context.Context, fqdn, txt string, clear bool) error {
	endpoint := c.baseURL.JoinPath("letsencrypt")

	query := endpoint.Query()
	query.Set("token", c.token)
	query.Set("domain", dns01.UnFqdn(fqdn))

	// api call differs per set/delete
	if clear {
		query.Set("action", "delete")
	} else {
		query.Set("value", txt)
	}

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response apiResponse
	err = json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	body := string(raw)
	if !response.Success {
		return fmt.Errorf("request to change TXT record for do.de returned the following error result (%s); used url [%s]", body, endpoint)
	}

	return nil
}
