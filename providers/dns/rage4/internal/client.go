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
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://rage4.com/rapi"

// Client the Rage4 API client.
type Client struct {
	username string
	password string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username:   username,
		password:   password,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a TXT record.
// https://rage4.com/swagger/index.html?urls.primaryName=dns-legacy#/DNS/CreateRecordPost
func (c *Client) CreateRecord(ctx context.Context, record Record) (*CommonResponse, error) {
	endpoint := c.BaseURL.JoinPath("CreateRecord")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := &CommonResponse{}

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRecord deletes a TXT record.
// https://rage4.com/swagger/index.html?urls.primaryName=dns-legacy#/DNS/DeleteRecordDelete
func (c *Client) DeleteRecord(ctx context.Context, recordID int) error {
	endpoint := c.BaseURL.JoinPath("DeleteRecord")

	query := endpoint.Query()
	query.Set("id", strconv.Itoa(recordID))
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// GetDomains returns all domains.
// https://rage4.com/swagger/index.html?urls.primaryName=dns-legacy#/DNS/GetDomains
func (c *Client) GetDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.BaseURL.JoinPath("GetDomains")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var domains []Domain

	err = c.do(req, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		switch resp.StatusCode / 100 {
		case 4:
			return parseErrorNEW[*APIError](req, resp)

		case 5:
			return parseErrorNEW[*ServerErrorResponse](req, resp)

		default:
			raw, _ := io.ReadAll(resp.Body)

			return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
		}
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

func parseErrorNEW[T error](req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var errAPI T

	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}
