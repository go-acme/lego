package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const baseURL = "https://api.wedos.com/wapi/json"

// Client the API client for Webos.
type Client struct {
	username string
	password string

	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username string, password string) *Client {
	return &Client{
		username:   username,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetRecords lists all the records in the zone.
// https://kb.wedos.com/en/wapi-api-interface/wapi-command-dns-rows-list/
func (c *Client) GetRecords(ctx context.Context, zone string) ([]DNSRow, error) {
	payload := map[string]any{
		"domain": dns01.UnFqdn(zone),
	}

	req, err := c.newRequest(ctx, commandDNSRowsList, payload)
	if err != nil {
		return nil, err
	}

	result := APIResponse[Rows]{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Response.Data.Rows, err
}

// AddRecord adds a record in the zone, either by updating existing records or creating new ones.
// https://kb.wedos.com/en/wapi-api-interface/wapi-command-dns-add-row/
// https://kb.wedos.com/en/wapi-api-interface/wapi-command-dns-row-update/
func (c *Client) AddRecord(ctx context.Context, zone string, record DNSRow) error {
	payload := DNSRowRequest{
		Domain: dns01.UnFqdn(zone),
		TTL:    record.TTL,
		Type:   record.Type,
		Data:   record.Data,
	}

	cmd := commandDNSRowAdd
	if record.ID == "" {
		payload.Name = record.Name
	} else {
		cmd = commandDNSRowUpdate
		payload.ID = record.ID
	}

	req, err := c.newRequest(ctx, cmd, payload)
	if err != nil {
		return err
	}

	return c.do(req, &APIResponse[json.RawMessage]{})
}

// DeleteRecord deletes a record from the zone.
// If a record does not have an ID, it will be looked up.
// https://kb.wedos.com/en/wapi-api-interface/wapi-command-dns-row-delete/
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordID string) error {
	payload := DNSRowRequest{
		Domain: dns01.UnFqdn(zone),
		ID:     recordID,
	}

	req, err := c.newRequest(ctx, commandDNSRowDelete, payload)
	if err != nil {
		return err
	}

	return c.do(req, &APIResponse[json.RawMessage]{})
}

// Commit not really required, all changes will be auto-committed after 5 minutes.
// https://kb.wedos.com/en/wapi-api-interface/wapi-command-dns-domain-commit/
func (c *Client) Commit(ctx context.Context, zone string) error {
	payload := map[string]any{
		"name": dns01.UnFqdn(zone),
	}

	req, err := c.newRequest(ctx, commandDNSDomainCommit, payload)
	if err != nil {
		return err
	}

	return c.do(req, &APIResponse[json.RawMessage]{})
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newRequest(ctx, commandPing, nil)
	if err != nil {
		return err
	}

	return c.do(req, &APIResponse[json.RawMessage]{})
}

func (c *Client) do(req *http.Request, result Response) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if result.GetCode() != codeOk {
		return fmt.Errorf("error %d: %s", result.GetCode(), result.GetResult())
	}

	return err
}

func (c *Client) newRequest(ctx context.Context, command string, payload any) (*http.Request, error) {
	requestObject := map[string]any{
		"request": APIRequest{
			User:    c.username,
			Auth:    authToken(c.username, c.password),
			Command: command,
			Data:    payload,
		},
	}

	object, err := json.Marshal(requestObject)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	form := url.Values{}
	form.Add("request", string(object))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}
