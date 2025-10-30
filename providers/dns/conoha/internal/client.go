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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const dnsServiceBaseURL = "https://dns-service.%s.conoha.io"

// Client is a ConoHa API client.
type Client struct {
	token string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient returns a client instance logged into the ConoHa service.
func NewClient(region, token string) (*Client, error) {
	baseURL, err := url.Parse(fmt.Sprintf(dnsServiceBaseURL, region))
	if err != nil {
		return nil, err
	}

	return &Client{
		token:      token,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// GetDomainID returns an ID of specified domain.
func (c *Client) GetDomainID(ctx context.Context, domainName string) (string, error) {
	domainList, err := c.getDomains(ctx)
	if err != nil {
		return "", err
	}

	for _, domain := range domainList.Domains {
		if domain.Name == domainName {
			return domain.ID, nil
		}
	}

	return "", fmt.Errorf("no such domain: %s", domainName)
}

// https://doc.conoha.jp/reference/api-vps2/api-dns-vps2/paas-dns-list-domains-v2/?btn_id=reference-api-vps2--sidebar_reference-paas-dns-list-domains-v2
func (c *Client) getDomains(ctx context.Context) (*DomainListResponse, error) {
	endpoint := c.baseURL.JoinPath("v1", "domains")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	domainList := &DomainListResponse{}

	err = c.do(req, domainList)
	if err != nil {
		return nil, err
	}

	return domainList, nil
}

// GetRecordID returns an ID of specified record.
func (c *Client) GetRecordID(ctx context.Context, domainID, recordName, recordType, data string) (string, error) {
	recordList, err := c.getRecords(ctx, domainID)
	if err != nil {
		return "", err
	}

	for _, record := range recordList.Records {
		if record.Name == recordName && record.Type == recordType && record.Data == data {
			return record.ID, nil
		}
	}

	return "", errors.New("no such record")
}

// https://doc.conoha.jp/reference/api-vps2/api-dns-vps2/paas-dns-list-records-in-a-domain-v2/?btn_id=reference-paas-dns-list-domains-v2--sidebar_reference-paas-dns-list-records-in-a-domain-v2
func (c *Client) getRecords(ctx context.Context, domainID string) (*RecordListResponse, error) {
	endpoint := c.baseURL.JoinPath("v1", "domains", domainID, "records")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	recordList := &RecordListResponse{}

	err = c.do(req, recordList)
	if err != nil {
		return nil, err
	}

	return recordList, nil
}

// CreateRecord adds new record.
func (c *Client) CreateRecord(ctx context.Context, domainID string, record Record) error {
	_, err := c.createRecord(ctx, domainID, record)
	return err
}

// https://doc.conoha.jp/reference/api-vps2/api-dns-vps2/paas-dns-create-record-v2/?btn_id=reference-paas-dns-list-records-in-a-domain-v2--sidebar_reference-paas-dns-create-record-v2
func (c *Client) createRecord(ctx context.Context, domainID string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("v1", "domains", domainID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	newRecord := &Record{}

	err = c.do(req, newRecord)
	if err != nil {
		return nil, err
	}

	return newRecord, nil
}

// DeleteRecord removes specified record.
// https://doc.conoha.jp/reference/api-vps2/api-dns-vps2/paas-dns-delete-a-record-v2/?btn_id=reference-paas-dns-create-record-v2--sidebar_reference-paas-dns-delete-a-record-v2
func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID string) error {
	endpoint := c.baseURL.JoinPath("v1", "domains", domainID, "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	if c.token != "" {
		req.Header.Set("X-Auth-Token", c.token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
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
