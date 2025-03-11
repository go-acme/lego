package active24

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
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://rest.%s"

// Client the Active24 API client.
type Client struct {
	apiKey string
	secret string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(baseAPIDomain, apiKey, secret string) (*Client, error) {
	if apiKey == "" || secret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(fmt.Sprintf(defaultBaseURL, baseAPIDomain))

	return &Client{
		apiKey:     apiKey,
		secret:     secret,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetServices lists of all services.
// https://rest.active24.cz/docs/v1.service#services
func (c *Client) GetServices(ctx context.Context) ([]Service, error) {
	endpoint := c.baseURL.JoinPath("v1", "user", "self", "service")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result OldAPIResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Items, err
}

// GetRecords lists of DNS records.
// https://rest.active24.cz/v2/docs#/DNS/rest.v2.dns.record_f94908d4e0e48489468498fce87cb90b
func (c *Client) GetRecords(ctx context.Context, service string, filter RecordFilter) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("v2", "service", service, "dns", "record")

	encodedFilter, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("marshal records filter: %w", err)
	}

	query := endpoint.Query()
	query.Add("filters", string(encodedFilter))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result APIResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, err
}

// CreateRecord creates a new DNS record.
// https://rest.active24.cz/v2/docs#/DNS/rest.v2.dns.create-record_6773d572235be9a72646bf6c54863573
func (c *Client) CreateRecord(ctx context.Context, service string, record Record) error {
	endpoint := c.baseURL.JoinPath("v2", "service", service, "dns", "record")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteRecord deletes a DNS record.
// https://rest.active24.cz/v2/docs#/DNS/rest.v2.dns.delete-record_fc6603c14848e547f8d0b967842f0a2c
func (c *Client) DeleteRecord(ctx context.Context, service, recordID string) error {
	endpoint := c.baseURL.JoinPath("v2", "service", service, "dns", "record", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("Accept-Language", "en_us")

	err := c.sign(req, time.Now())
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
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

	var errAPI APIError
	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}

// sign creates and sets request signature and date.
// https://rest.active24.cz/v2/docs/intro
func (c *Client) sign(req *http.Request, now time.Time) error {
	if req.URL.Path == "" {
		req.URL.Path += "/"
	}

	canonicalRequest := fmt.Sprintf("%s %s %d", req.Method, req.URL.Path, now.Unix())

	mac := hmac.New(sha1.New, []byte(c.secret))
	_, err := mac.Write([]byte(canonicalRequest))
	if err != nil {
		return err
	}

	hashed := mac.Sum(nil)
	signature := hex.EncodeToString(hashed)

	req.SetBasicAuth(c.apiKey, signature)

	req.Header.Set("Date", now.Format(time.RFC3339))

	return nil
}
