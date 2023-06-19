package internal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// Default API endpoints.
const (
	DefaultSandboxBaseURL = "https://api.sandbox.dnsmadeeasy.com/V2.0"
	DefaultProdBaseURL    = "https://api.dnsmadeeasy.com/V2.0"
)

// Client DNSMadeEasy client.
type Client struct {
	apiKey    string
	apiSecret string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a DNSMadeEasy client.
func NewClient(apiKey, apiSecret string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing: API key")
	}

	if apiSecret == "" {
		return nil, errors.New("credentials missing: API secret")
	}

	baseURL, _ := url.Parse(DefaultProdBaseURL)

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// GetDomain gets a domain.
func (c *Client) GetDomain(ctx context.Context, authZone string) (*Domain, error) {
	endpoint := c.BaseURL.JoinPath("dns", "managed", "name")

	domainName := authZone[0 : len(authZone)-1]

	query := endpoint.Query()
	query.Set("domainname", domainName)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	domain := &Domain{}
	err = c.do(req, domain)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

// GetRecords gets all TXT records.
func (c *Client) GetRecords(ctx context.Context, domain *Domain, recordName, recordType string) (*[]Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", "managed", strconv.Itoa(domain.ID), "records")

	query := endpoint.Query()
	query.Set("recordName", recordName)
	query.Set("type", recordType)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	records := &recordsResponse{}
	err = c.do(req, records)
	if err != nil {
		return nil, err
	}

	return records.Records, nil
}

// CreateRecord creates a TXT records.
func (c *Client) CreateRecord(ctx context.Context, domain *Domain, record *Record) error {
	endpoint := c.BaseURL.JoinPath("dns", "managed", strconv.Itoa(domain.ID), "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteRecord deletes a TXT records.
func (c *Client) DeleteRecord(ctx context.Context, record Record) error {
	endpoint := c.BaseURL.JoinPath("dns", "managed", strconv.Itoa(record.SourceID), "records", strconv.Itoa(record.ID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	err := c.sign(req, time.Now().UTC().Format(time.RFC1123))
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if err = json.Unmarshal(raw, result); err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func (c *Client) sign(req *http.Request, timestamp string) error {
	signature, err := computeHMAC(timestamp, c.apiSecret)
	if err != nil {
		return err
	}

	req.Header.Set("x-dnsme-apiKey", c.apiKey)
	req.Header.Set("x-dnsme-requestDate", timestamp)
	req.Header.Set("x-dnsme-hmac", signature)

	return nil
}

func computeHMAC(message, secret string) (string, error) {
	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	_, err := h.Write([]byte(message))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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
