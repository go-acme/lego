// Package internal inspired by https://gist.github.com/dankrause/5585907
package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://www.hover.com/api"

// Client is the client context for communicating with Hover DNS API;
// should only need one of these but keeping state isolated to instances rather than global where possible.
type Client struct {
	username string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a Hover client using plaintext passwords against plain username.
// Consider the risk of where the text is stored.
func NewClient(username, password string) (*Client, error) {
	if username == "" {
		return nil, errors.New("incomplete credentials, missing username")
	}
	if password == "" {
		return nil, errors.New("incomplete credentials, missing password")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username: username,
		password: password,
		baseURL:  baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) AddTXTRecord(ctx context.Context, domainID, fqdn, value string) error {
	values := url.Values{}
	values.Set("name", fqdn)
	values.Set("type", "TXT")
	values.Set("content", value)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.JoinPath("domains", domainID, "dns").String(), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return c.do(req, nil)
}

func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL.JoinPath("domains", domainID, "dns", recordID).String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("httpDelete: creating new request: %w", err)
	}

	return c.do(req, nil)
}

func (c *Client) GetDomains(ctx context.Context) ([]Domain, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.JoinPath("domains").String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	list := APIResponse{}

	err = c.do(req, &list)
	if err != nil {
		return nil, err
	}

	return list.Domains, nil
}

// GetRecords gets the records for a specific domain.
func (c *Client) GetRecords(ctx context.Context, domainID string) ([]Record, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.JoinPath("domains", domainID, "dns").String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	var list APIResponse

	err = c.do(req, &list)
	if err != nil {
		return nil, err
	}

	var records []Record
	for _, d := range list.Domains {
		if d.ID == domainID {
			records = append(records, d.Records...)
		}
	}

	return records, nil
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
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

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	response := &APIResponse{}
	err := json.Unmarshal(raw, response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("%d: %w", resp.StatusCode, response)
}
