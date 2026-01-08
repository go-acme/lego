package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://todapi.now.cn:2443"

// Client the TodayNIC API client.
type Client struct {
	authUserID string
	apiKey     string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(authUserID, apiKey string) (*Client, error) {
	if authUserID == "" || apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		authUserID: authUserID,
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, record Record) (int, error) {
	endpoint := c.BaseURL.JoinPath("api", "dns", "add-domain-record.json")

	query, err := querystring.Values(record)
	if err != nil {
		return 0, err
	}

	req, err := c.newRequest(ctx, endpoint, query)
	if err != nil {
		return 0, err
	}

	var result APIResponse

	err = c.do(req, &result)
	if err != nil {
		return 0, err
	}

	return result.ID, nil
}

func (c *Client) DeleteRecord(ctx context.Context, recordID int) error {
	endpoint := c.BaseURL.JoinPath("api", "dns", "delete-domain-record.json")

	query := endpoint.Query()
	query.Set("Id", strconv.Itoa(recordID))

	req, err := c.newRequest(ctx, endpoint, query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

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

func (c *Client) newRequest(ctx context.Context, endpoint *url.URL, query url.Values) (*http.Request, error) {
	query.Set("auth-userid", c.authUserID)
	query.Set("api-key", c.apiKey)

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

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
