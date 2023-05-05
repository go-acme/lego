package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const apiBaseURL = "https://admin.vshosting.cloud/clouddns"

const authorizationHeader = "Authorization"

// Client handles all communication with CloudDNS API.
type Client struct {
	clientID string
	email    string
	password string
	ttl      int

	apiBaseURL *url.URL

	loginURL *url.URL

	HTTPClient *http.Client
}

// NewClient returns a Client instance configured to handle CloudDNS API communication.
func NewClient(clientID, email, password string, ttl int) *Client {
	baseURL, _ := url.Parse(apiBaseURL)
	loginBaseURL, _ := url.Parse(loginURL)

	return &Client{
		clientID:   clientID,
		email:      email,
		password:   password,
		ttl:        ttl,
		apiBaseURL: baseURL,
		loginURL:   loginBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddRecord is a high level method to add a new record into CloudDNS zone.
func (c *Client) AddRecord(ctx context.Context, zone, recordName, recordValue string) error {
	domain, err := c.getDomain(ctx, zone)
	if err != nil {
		return err
	}

	record := Record{DomainID: domain.ID, Name: recordName, Value: recordValue, Type: "TXT"}

	err = c.addTxtRecord(ctx, record)
	if err != nil {
		return err
	}

	return c.publishRecords(ctx, domain.ID)
}

// DeleteRecord is a high level method to remove a record from zone.
func (c *Client) DeleteRecord(ctx context.Context, zone, recordName string) error {
	domain, err := c.getDomain(ctx, zone)
	if err != nil {
		return err
	}

	record, err := c.getRecord(ctx, domain.ID, recordName)
	if err != nil {
		return err
	}

	err = c.deleteRecord(ctx, record)
	if err != nil {
		return err
	}

	return c.publishRecords(ctx, domain.ID)
}

func (c *Client) addTxtRecord(ctx context.Context, record Record) error {
	endpoint := c.apiBaseURL.JoinPath("record-txt")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) deleteRecord(ctx context.Context, record Record) error {
	endpoint := c.apiBaseURL.JoinPath("record", record.ID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) getDomain(ctx context.Context, zone string) (Domain, error) {
	searchQuery := SearchQuery{
		Search: []Search{
			{Name: "clientId", Operator: "eq", Value: c.clientID},
			{Name: "domainName", Operator: "eq", Value: zone},
		},
	}

	endpoint := c.apiBaseURL.JoinPath("domain", "search")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, searchQuery)
	if err != nil {
		return Domain{}, err
	}

	var result SearchResponse
	err = c.do(req, &result)
	if err != nil {
		return Domain{}, err
	}

	if len(result.Items) == 0 {
		return Domain{}, fmt.Errorf("domain not found: %s", zone)
	}

	return result.Items[0], nil
}

func (c *Client) getRecord(ctx context.Context, domainID, recordName string) (Record, error) {
	endpoint := c.apiBaseURL.JoinPath("domain", domainID)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Record{}, err
	}

	var result DomainInfo
	err = c.do(req, &result)
	if err != nil {
		return Record{}, err
	}

	for _, record := range result.LastDomainRecordList {
		if record.Name == recordName && record.Type == "TXT" {
			return record, nil
		}
	}

	return Record{}, fmt.Errorf("record not found: domainID %s, name %s", domainID, recordName)
}

func (c *Client) publishRecords(ctx context.Context, domainID string) error {
	endpoint := c.apiBaseURL.JoinPath("domain", domainID, "publish")

	payload := DomainInfo{SoaTTL: c.ttl}

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	at := getAccessToken(req.Context())
	if at != "" {
		req.Header.Set(authorizationHeader, "Bearer "+at)
	}

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

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var response APIError
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, response.Error)
}
