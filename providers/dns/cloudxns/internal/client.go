package internal

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://www.cloudxns.net/api2/"

// Client CloudXNS client.
type Client struct {
	apiKey    string
	secretKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a CloudXNS client.
func NewClient(apiKey, secretKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing: apiKey")
	}

	if secretKey == "" {
		return nil, errors.New("credentials missing: secretKey")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		secretKey:  secretKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetDomainInformation Get domain name information for a FQDN.
func (c *Client) GetDomainInformation(ctx context.Context, fqdn string) (*Data, error) {
	endpoint := c.baseURL.JoinPath("domain")

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	var domains []Data
	err = c.do(req, &domains)
	if err != nil {
		return nil, err
	}

	for _, data := range domains {
		if data.Domain == authZone {
			return &data, nil
		}
	}

	return nil, fmt.Errorf("zone %s not found for domain %s", authZone, fqdn)
}

// FindTxtRecord return the TXT record a zone ID and a FQDN.
func (c *Client) FindTxtRecord(ctx context.Context, zoneID, fqdn string) (*TXTRecord, error) {
	endpoint := c.baseURL.JoinPath("record", zoneID)

	query := endpoint.Query()
	query.Set("host_id", "0")
	query.Set("offset", "0")
	query.Set("row_num", "2000")
	endpoint.RawQuery = query.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records []TXTRecord
	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.Host == dns01.UnFqdn(fqdn) && record.Type == "TXT" {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("no existing record found for %q", fqdn)
}

// AddTxtRecord add a TXT record.
func (c *Client) AddTxtRecord(ctx context.Context, info *Data, fqdn, value string, ttl int) error {
	id, err := strconv.Atoi(info.ID)
	if err != nil {
		return fmt.Errorf("invalid zone ID: %w", err)
	}

	endpoint := c.baseURL.JoinPath("record")

	subDomain, err := dns01.ExtractSubDomain(fqdn, info.Domain)
	if err != nil {
		return err
	}

	record := TXTRecord{
		ID:     id,
		Host:   subDomain,
		Value:  value,
		Type:   "TXT",
		LineID: 1,
		TTL:    ttl,
	}

	req, err := c.newRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// RemoveTxtRecord remove a TXT record.
func (c *Client) RemoveTxtRecord(ctx context.Context, recordID, zoneID string) error {
	endpoint := c.baseURL.JoinPath("record", recordID, zoneID)

	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
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

	if response.Code != 1 {
		return fmt.Errorf("[status code %d] invalid code (%v) error: %s", resp.StatusCode, response.Code, response.Message)
	}

	if result == nil {
		return nil
	}

	if len(response.Data) == 0 {
		return nil
	}

	err = json.Unmarshal(response.Data, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
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

	requestDate := time.Now().Format(time.RFC1123Z)

	req.Header.Set("API-KEY", c.apiKey)
	req.Header.Set("API-REQUEST-DATE", requestDate)
	req.Header.Set("API-HMAC", c.hmac(endpoint.String(), requestDate, buf.String()))
	req.Header.Set("API-FORMAT", "json")

	return req, nil
}

func (c *Client) hmac(endpoint, date, body string) string {
	sum := md5.Sum([]byte(c.apiKey + endpoint + body + date + c.secretKey))
	return hex.EncodeToString(sum[:])
}
