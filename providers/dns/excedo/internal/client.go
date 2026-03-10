package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

type responseChecker interface {
	Check() error
}

// Client the Excedo API client.
type Client struct {
	apiKey string

	baseURL    *url.URL
	HTTPClient *http.Client

	token   *ExpirableToken
	muToken sync.Mutex
}

// NewClient creates a new Client.
func NewClient(apiURL, apiKey string) (*Client, error) {
	if apiURL == "" || apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, record Record) (int64, error) {
	payload, err := querystring.Values(record)
	if err != nil {
		return 0, err
	}

	endpoint := c.baseURL.JoinPath("/dns/addrecord/")

	req, err := newFormRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return 0, err
	}

	result := new(AddRecordResponse)

	err = c.doAuthenticated(ctx, req, result)
	if err != nil {
		return 0, err
	}

	return result.RecordID, nil
}

func (c *Client) DeleteRecord(ctx context.Context, zone, recordID string) error {
	endpoint := c.baseURL.JoinPath("/dns/deleterecord/")

	data := map[string]string{
		"domainname": dns01.UnFqdn(zone),
		"recordid":   recordID,
	}

	req, err := newMultipartRequest(ctx, http.MethodPost, endpoint, data)
	if err != nil {
		return err
	}

	result := new(BaseResponse)

	err = c.doAuthenticated(ctx, req, result)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetRecords(ctx context.Context, zone string) (map[string]Zone, error) {
	endpoint := c.baseURL.JoinPath("/dns/getrecords/")

	query := endpoint.Query()
	query.Set("domainname", zone)

	endpoint.RawQuery = query.Encode()

	req, err := newFormRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(GetRecordsResponse)

	err = c.doAuthenticated(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result.DNS, nil
}

func (c *Client) do(req *http.Request, result responseChecker) error {
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

	return result.Check()
}

func newMultipartRequest(ctx context.Context, method string, endpoint *url.URL, data map[string]string) (*http.Request, error) {
	buf := new(bytes.Buffer)

	writer := multipart.NewWriter(buf)

	for k, v := range data {
		err := writer.WriteField(k, v)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(buf.Bytes())

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func newFormRequest(ctx context.Context, method string, endpoint *url.URL, form url.Values) (*http.Request, error) {
	var body io.Reader

	if len(form) > 0 {
		body = bytes.NewReader([]byte(form.Encode()))
	} else {
		body = http.NoBody
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}
