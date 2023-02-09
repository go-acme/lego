package internal

import (
	"bytes"
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
)

const defaultBaseURL = "https://rest.websupport.sk"

// StatusSuccess expected status text when success.
const StatusSuccess = "success"

// Client a Websupport DNS API client.
type Client struct {
	apiKey     string
	secretKey  string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, secretKey string) (*Client, error) {
	if apiKey == "" || secretKey == "" {
		return nil, errors.New("credentials missing")
	}

	return &Client{
		apiKey:     apiKey,
		secretKey:  secretKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GetUser gets a user detail.
// https://rest.websupport.sk/docs/v1.user#user
func (c *Client) GetUser(userID string) (*User, error) {
	endpoint, err := url.JoinPath(c.BaseURL, "v1", "user", userID)
	if err != nil {
		return nil, fmt.Errorf("base url parsing: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
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
func (c *Client) ListRecords(domainName string) (*ListResponse, error) {
	endpoint, err := url.JoinPath(c.BaseURL, "v1", "user", "self", "zone", domainName, "record")
	if err != nil {
		return nil, fmt.Errorf("base url parsing: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
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
func (c *Client) GetRecords(domainName string, recordID int) (*Record, error) {
	endpoint, err := url.JoinPath(c.BaseURL, "v1", "user", "self", "zone", domainName, "record", strconv.Itoa(recordID))
	if err != nil {
		return nil, fmt.Errorf("base url parsing: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
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
func (c *Client) AddRecord(domainName string, record Record) (*Response, error) {
	endpoint, err := url.JoinPath(c.BaseURL, "v1", "user", "self", "zone", domainName, "record")
	if err != nil {
		return nil, fmt.Errorf("base url parsing: %w", err)
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("request payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
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
func (c *Client) DeleteRecord(domainName string, recordID int) (*Response, error) {
	endpoint, err := url.JoinPath(c.BaseURL, "v1", "user", "self", "zone", domainName, "record", strconv.Itoa(recordID))
	if err != nil {
		return nil, fmt.Errorf("base url parsing: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("request payload: %w", err)
	}

	result := &Response{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
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
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > http.StatusBadRequest {
		all, _ := io.ReadAll(resp.Body)

		var e APIError
		err = json.Unmarshal(all, &e)
		if err != nil {
			return fmt.Errorf("%d: %s", resp.StatusCode, string(all))
		}

		return &e
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	err = json.Unmarshal(all, result)
	if err != nil {
		return fmt.Errorf("unmarshal response body: %w", err)
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
