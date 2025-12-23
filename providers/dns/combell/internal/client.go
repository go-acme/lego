package internal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	querystring "github.com/google/go-querystring/query"
	"github.com/hashicorp/go-retryablehttp"
)

const defaultBaseURL = "https://api.combell.com/v2"

// Client a Combell DNS API client.
type Client struct {
	apiKey    string
	apiSecret string

	nonce *Nonce

	httpClient *http.Client
	BaseURL    *url.URL
}

// NewClient creates a new Client.
func NewClient(apiKey, apiSecret string, httpClient *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	retryClient := retryablehttp.NewClient()

	retryClient.RetryMax = 5
	if httpClient != nil {
		retryClient.HTTPClient = httpClient
	}

	retryClient.Logger = log.Default()

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: retryClient.StandardClient(),
		BaseURL:    baseURL,
		nonce:      NewNonce(),
	}
}

// GetRecords gets the records of a domain.
// https://api.combell.com/v2/documentation#tag/DNS-records/paths/~1dns~1{domainName}~1records/get
func (c Client) GetRecords(ctx context.Context, domainName string, request *GetRecordsRequest) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", domainName, "records")

	if request != nil {
		values, err := querystring.Values(request)
		if err != nil {
			return nil, err
		}

		endpoint.RawQuery = values.Encode()
	}

	req, err := c.newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var results []Record

	err = c.do(req, http.StatusOK, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// CreateRecord creates a record.
// https://api.combell.com/v2/documentation#tag/DNS-records/paths/~1dns~1{domainName}~1records/post
func (c Client) CreateRecord(ctx context.Context, domainName string, record Record) error {
	endpoint := c.BaseURL.JoinPath("dns", domainName, "records")

	req, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// TODO(ldez) the "Location" header contains a reference to the created record.

	return c.do(req, http.StatusCreated, nil)
}

// GetRecord gets a specific record.
// https://api.combell.com/v2/documentation#tag/DNS-records/paths/~1dns~1{domainName}~1records~1{recordId}/get
func (c Client) GetRecord(ctx context.Context, domainName, recordID string) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("dns", domainName, "records", recordID)

	req, err := c.newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var result Record

	err = c.do(req, http.StatusOK, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteRecord deletes a record.
// https://api.combell.com/v2/documentation#tag/DNS-records/paths/~1dns~1{domainName}~1records~1{recordId}/delete
func (c Client) DeleteRecord(ctx context.Context, domainName, recordID string) error {
	endpoint := c.BaseURL.JoinPath("dns", domainName, "records", recordID)

	req, err := c.newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, http.StatusNoContent, nil)
}

func (c Client) do(req *http.Request, expectedStatus int, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != expectedStatus {
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

func (c Client) newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	var body []byte

	if payload != nil {
		var err error

		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}

		buf = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	sign, err := c.sign(req, body)
	if err != nil {
		return nil, fmt.Errorf("request signature: %w", err)
	}

	req.Header.Set("Authorization", sign)

	return req, nil
}

func (c Client) sign(req *http.Request, body []byte) (string, error) {
	unix := time.Now().Unix()
	nonce := c.nonce.Generate(10)

	var encodedBody string

	if len(body) > 0 {
		sum := md5.Sum(body)
		encodedBody = base64.StdEncoding.EncodeToString(sum[:])
	}

	h := hmac.New(sha256.New, []byte(c.apiSecret))

	_, err := h.Write([]byte(
		c.apiKey + strings.ToLower(req.Method) + url.QueryEscape(strings.ToLower(req.URL.RequestURI())) + strconv.FormatInt(unix, 10) + nonce + encodedBody),
	)
	if err != nil {
		return "", err
	}

	hSign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("hmac %s:%s:%s:%d", c.apiKey, hSign, nonce, unix), nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var response APIError

	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, response)
}
