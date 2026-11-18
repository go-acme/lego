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

	"github.com/go-acme/lego/v5/internal/errutils"
)

// defaultBaseURL endpoint is the Gandi API endpoint used by Present and CleanUp.
const defaultBaseURL = "https://api.gandi.net/v5/livedns"

// Related to Personal Access Token.
const authorizationHeader = "Authorization"

// Client the Gandi API v5 client.
type Client struct {
	apiKey string
	pat    string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(apiKey, pat string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		pat:        pat,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddVerificationRecord adds a verification record.
// https://api.gandi.net/docs/livedns/#post-v5-livedns-domains-fqdn-verification-rrset_name-rrset_type
func (c *Client) AddVerificationRecord(ctx context.Context, domain, name, value string) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain, "verification", name, "TXT")

	newRecord := &Record{Values: []string{value}}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, newRecord)
	if err != nil {
		return nil, err
	}

	result := new(Record)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteTXTRecord deletes a TXT record.
// https://api.gandi.net/docs/livedns/#delete-v5-livedns-domains-fqdn-records-rrset_name-rrset_type
func (c *Client) DeleteTXTRecord(ctx context.Context, domain, name string) error {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records", name, "TXT")

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	var result apiResponse

	err = c.do(req, &result)
	if err != nil {
		return err
	}

	return nil
}

// GetTXTRecord gets a record set.
// https://api.gandi.net/docs/livedns/#get-v5-livedns-domains-fqdn-records-rrset_name-rrset_type
func (c *Client) GetTXTRecord(ctx context.Context, domain, name string) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records", name, "TXT")

	// Get exiting values for the TXT records
	// Needed to create challenges for both wildcard and base name domains
	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(Record)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// OverwriteTXTRecord overwrites a single record.
// https://api.gandi.net/docs/livedns/#put-v5-livedns-domains-fqdn-records-rrset_name-rrset_type
func (c *Client) OverwriteTXTRecord(ctx context.Context, domain, name string, newRecord *Record) error {
	endpoint := c.BaseURL.JoinPath("domains", domain, "records", name, "TXT")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, newRecord)
	if err != nil {
		return err
	}

	var result apiResponse

	err = c.do(req, &result)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	if c.apiKey != "" {
		req.Header.Set(authorizationHeader, "Apikey "+c.apiKey)
	}

	if c.pat != "" {
		req.Header.Set(authorizationHeader, "Bearer "+c.pat)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	err = checkResponse(req, resp)
	if err != nil {
		return err
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if len(raw) > 0 {
		err = json.Unmarshal(raw, result)
		if err != nil {
			return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
		}
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

func checkResponse(req *http.Request, resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound && resp.Request.Method == http.MethodGet {
		return nil
	}

	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	return parseError(req, resp)
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errAPI := new(APIError)

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	if errAPI.Code == 0 {
		errAPI.Code = resp.StatusCode
	}

	return errAPI
}
