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
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://apiv2.xinnet.com"

type RequestSigner interface {
	Sign(req *http.Request) error
}

// Client the Xinnet API client.
type Client struct {
	signer RequestSigner

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(signer RequestSigner) (*Client, error) {
	if signer == nil {
		return nil, errors.New("request signer missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		signer:     signer,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord creates a record.
// https://apidoc.xin.cn/doc-7283900
func (c *Client) CreateRecord(ctx context.Context, record Record) (int64, error) {
	endpoint := c.BaseURL.JoinPath("api", "dns", "create", "/")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return 0, err
	}

	var result int64

	err = c.do(req, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// DeleteRecord deletes a record.
// https://apidoc.xin.cn/doc-7283901
func (c *Client) DeleteRecord(ctx context.Context, domain string, recordID int64) error {
	endpoint := c.BaseURL.JoinPath("api", "dns", "delete", "/")

	payload := map[string]any{
		"domainName": domain,
		"recordId":   recordID,
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	err := c.signer.Sign(req)
	if err != nil {
		return err
	}

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

	var response APIResponse

	err = json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if response.Code != "0" {
		return fmt.Errorf("%s: %s (%s)", response.Code, response.Message, response.RequestID)
	}

	if result == nil {
		return nil
	}

	err = json.Unmarshal(response.Data, &result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, response.Data, err)
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
