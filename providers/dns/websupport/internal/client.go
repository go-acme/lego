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

const defaultBaseURL = "https://rest.websupport.sk"

// StatusSuccess expected status text when success.
const StatusSuccess = "success"

// Client a Websupport DNS API client.
type Client struct {
	apiKey    string
	secretKey string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, secretKey string) (*Client, error) {
	if apiKey == "" || secretKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		secretKey:  secretKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetUser gets a user detail.
// https://rest.websupport.sk/docs/v1.user#user
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	endpoint := c.baseURL.JoinPath("v1", "user", userID)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request payload: %w", err)
	}

	result := &User{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListRecords lists all records.
// https://rest.websupport.sk/docs/v1.zone#records
func (c *Client) ListRecords(ctx context.Context, domainName string) (*ListResponse, error) {
	endpoint := c.baseURL.JoinPath("v1", "user", "self", "zone", domainName, "record")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request payload: %w", err)
	}

	result := &ListResponse{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetRecords gets a DNS record.
func (c *Client) GetRecords(ctx context.Context, domainName string, recordID int) (*Record, error) {
	endpoint := c.baseURL.JoinPath("v1", "user", "self", "zone", domainName, "record", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &Record{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// AddRecord adds a DNS record.
// https://rest.websupport.sk/docs/v1.zone#post-record
func (c *Client) AddRecord(ctx context.Context, domainName string, record Record) (*Response, error) {
	endpoint := c.baseURL.JoinPath("v1", "user", "self", "zone", domainName, "record")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := &Response{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRecord deletes a DNS record.
// https://rest.websupport.sk/docs/v1.zone#delete-record
func (c *Client) DeleteRecord(ctx context.Context, domainName string, recordID int) (*Response, error) {
	endpoint := c.baseURL.JoinPath("v1", "user", "self", "zone", domainName, "record", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result := &Response{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("Accept-Language", "en_us")

	location, err := time.LoadLocation("GMT")
	if err != nil {
		return fmt.Errorf("time location: %w", err)
	}

	err = c.sign(req, time.Now().In(location))
	if err != nil {
		return fmt.Errorf("signature: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > http.StatusBadRequest {
		return parseError(req, resp)
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

func (c *Client) sign(req *http.Request, now time.Time) error {
	if req.URL.Path == "" {
		req.URL.Path += "/"
	}

	canonicalRequest := fmt.Sprintf("%s %s %d", req.Method, req.URL.Path, now.Unix())

	mac := hmac.New(sha1.New, []byte(c.secretKey))
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
