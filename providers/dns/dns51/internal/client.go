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

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://www.51dns.com/api/"

// Client the 51DNS API client.
type Client struct {
	apiKey    string
	apiSecret string

	clock func() time.Time

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, apiSecret string) (*Client, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		clock:      time.Now,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a record.
// https://www.51dns.com/document/api/4/12.html
func (c *Client) CreateRecord(ctx context.Context, record RecordRequest) (*RecordData, error) {
	endpoint := c.BaseURL.JoinPath("record", "create")

	req, err := c.newSignedRequest(ctx, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := &RecordData{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRecord deletes a record.
// https://www.51dns.com/document/api/4/27.html
func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int64) error {
	endpoint := c.BaseURL.JoinPath("record", "remove")

	payload := struct {
		DomainID int64 `json:"domainID" url:"domainID"`
		RecordID int64 `json:"recordID" url:"recordID"`
	}{
		DomainID: domainID,
		RecordID: recordID,
	}

	req, err := c.newSignedRequest(ctx, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// ListDomains lists domains.
// https://www.51dns.com/document/api/74/88.html
func (c *Client) ListDomains(ctx context.Context, request DomainRequest) (*DomainData, error) {
	endpoint := c.BaseURL.JoinPath("domain", "list")

	req, err := c.newSignedRequest(ctx, endpoint, request)
	if err != nil {
		return nil, err
	}

	result := &DomainData{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	response := &APIResponse{}

	err = json.Unmarshal(raw, response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	// https://www.51dns.com/document/api/70/73.html
	if response.Code != 0 {
		return fmt.Errorf("%d: %s", response.Code, response.Message)
	}

	if result == nil {
		return nil
	}

	err = json.Unmarshal(response.Data, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, response.Data, err)
	}

	return nil
}

func (c *Client) newSignedRequest(ctx context.Context, endpoint *url.URL, data any) (*http.Request, error) {
	values, err := querystring.Values(data)
	if err != nil {
		return nil, err
	}

	values.Set("apiKey", c.apiKey)
	values.Set("timestamp", strconv.FormatInt(c.clock().UTC().Unix(), 10))

	signature, err := c.sign(values)
	if err != nil {
		return nil, err
	}

	values.Set("hash", signature)

	payload := map[string]any{}

	for k, v := range values {
		payload[k] = v[0]
	}

	buf := new(bytes.Buffer)

	err = json.NewEncoder(buf).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) sign(values url.Values) (string, error) {
	unescape, err := url.QueryUnescape(values.Encode())
	if err != nil {
		return "", err
	}

	m := md5.New()
	m.Write([]byte(unescape + c.apiSecret))

	return hex.EncodeToString(m.Sum(nil)), nil
}
