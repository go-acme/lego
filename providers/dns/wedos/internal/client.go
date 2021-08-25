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
)

const baseURL = "https://api.wedos.com/wapi/json"

const codeOk = 1000

const (
	commandPing            = "ping"
	commandDNSDomainCommit = "dns-domain-commit"
	commandDNSRowsList     = "dns-rows-list"
	commandDNSRowDelete    = "dns-row-delete"
	commandDNSRowAdd       = "dns-row-add"
	commandDNSRowUpdate    = "dns-row-update"
)

type ResponsePayload struct {
	Code      int             `json:"code,omitempty"`
	Result    string          `json:"result,omitempty"`
	Timestamp int             `json:"timestamp,omitempty"`
	SvTRID    string          `json:"svTRID,omitempty"`
	Command   string          `json:"command,omitempty"`
	Data      json.RawMessage `json:"data"`
}

type DNSRow struct {
	ID   string      `json:"ID,omitempty"`
	Name string      `json:"name,omitempty"`
	TTL  json.Number `json:"ttl,omitempty" type:"integer"`
	Type string      `json:"rdtype,omitempty"`
	Data string      `json:"rdata"`
}

type DNSRowRequest struct {
	ID     string      `json:"row_id,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Name   string      `json:"name,omitempty"`
	TTL    json.Number `json:"ttl,omitempty" type:"integer"`
	Type   string      `json:"type,omitempty"`
	Data   string      `json:"rdata"`
}

type APIRequest struct {
	User    string      `json:"user,omitempty"`
	Auth    string      `json:"auth,omitempty"`
	Command string      `json:"command,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Client struct {
	username   string
	password   string
	baseURL    string
	HTTPClient *http.Client
}

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
	payload := map[string]interface{}{
		"domain": dns01.UnFqdn(zone),
	}

	resp, err := c.do(ctx, commandDNSRowsList, payload)
	if err != nil {
		return nil, err
	}

	arrayWrapper := struct {
		Rows []DNSRow `json:"row"`
	}{}

	err = json.Unmarshal(resp.Data, &arrayWrapper)
	if err != nil {
		return nil, err
	}

	return arrayWrapper.Rows, err
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

	_, err := c.do(ctx, cmd, payload)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecord deletes a record from the zone.
// If a record does not have an ID, it will be looked up.
// https://kb.wedos.com/en/wapi-api-interface/wapi-command-dns-row-delete/
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordID string) error {
	payload := DNSRowRequest{
		Domain: dns01.UnFqdn(zone),
		ID:     recordID,
	}

	_, err := c.do(ctx, commandDNSRowDelete, payload)
	if err != nil {
		return err
	}

	return nil
}

// Commit not really required, all changes will be auto-committed after 5 minutes.
// https://kb.wedos.com/en/wapi-api-interface/wapi-command-dns-domain-commit/
func (c *Client) Commit(ctx context.Context, zone string) error {
	payload := map[string]interface{}{
		"name": dns01.UnFqdn(zone),
	}

	_, err := c.do(ctx, commandDNSDomainCommit, payload)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.do(ctx, commandPing, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(ctx context.Context, command string, payload interface{}) (*ResponsePayload, error) {
	requestObject := map[string]interface{}{
		"request": APIRequest{
			User:    c.username,
			Auth:    authToken(c.username, c.password),
			Command: command,
			Data:    payload,
		},
	}

	jsonBytes, err := json.Marshal(requestObject)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("request", string(jsonBytes))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error, status code: %d", resp.StatusCode)
	}

	responseWrapper := struct {
		Response ResponsePayload `json:"response"`
	}{}

	err = json.Unmarshal(body, &responseWrapper)
	if err != nil {
		return nil, err
	}

	if responseWrapper.Response.Code != codeOk {
		return nil, fmt.Errorf("wedos responded with error code %d = %s", responseWrapper.Response.Code, responseWrapper.Response.Result)
	}

	return &responseWrapper.Response, err
}
