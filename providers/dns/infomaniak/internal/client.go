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
	"golang.org/x/oauth2"
)

const DefaultBaseURL = "https://api.infomaniak.com"

const statusSuccess = "success"

// Client the Infomaniak API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client, apiEndpoint string) (*Client, error) {
	baseURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}

	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Second}
	}

	return &Client{baseURL: baseURL, httpClient: hc}, nil
}

// CreateRecord creates a new DNS record.
// https://developer.infomaniak.com/docs/api/post/2/zones/%7Bzone%7D/records
func (c *Client) CreateRecord(ctx context.Context, zone string, payload RecordRequest) (*Record, error) {
	endpoint := c.baseURL.JoinPath("2", "zones", zone, "records")

	query := endpoint.Query()
	query.Set("with", "idn")
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[*Record])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	if result.Result != statusSuccess {
		return nil, fmt.Errorf("%s: %w", result.Result, result.Error)
	}

	return result.Data, nil
}

// DeleteRecord deletes a DNS record.
// https://developer.infomaniak.com/docs/api/delete/2/zones/%7Bzone%7D/records/%7Brecord%7D
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordID int) error {
	endpoint := c.baseURL.JoinPath("2", "zones", zone, "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	result := new(APIResponse[any])

	err = c.do(req, result)
	if err != nil {
		return err
	}

	if result.Result != statusSuccess {
		return fmt.Errorf("%s: %w", result.Result, result.Error)
	}

	return nil
}

// ZoneExists checks if a zone exists.
// https://developer.infomaniak.com/docs/api/get/2/zones/%7Bzone%7D/exists
func (c *Client) ZoneExists(ctx context.Context, zone string) (bool, error) {
	endpoint := c.baseURL.JoinPath("2", "zones", zone, "exists")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}

	result := new(APIResponse[bool])

	err = c.do(req, result)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Code == "object_not_found" {
			return false, nil
		}

		return false, err
	}

	if result.Result != statusSuccess {
		return false, fmt.Errorf("%s: %s: %w", zone, result.Result, result.Error)
	}

	return result.Data, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.httpClient.Do(req)
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

	result := new(APIResponse[any])

	err := json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("%s: %w", result.Result, result.Error)
}

func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
