package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// Default API endpoints.
const (
	APIBaseURL  = "https://console.cloud.ru/api/clouddns/v1"
	AuthBaseURL = "https://auth.iam.cloud.ru/auth/system/openid/token"
)

// Client the Cloud.ru API client.
type Client struct {
	keyID  string
	secret string

	APIEndpoint  *url.URL
	AuthEndpoint *url.URL
	HTTPClient   *http.Client

	token   *Token
	muToken sync.Mutex
}

// NewClient Creates a new Client.
func NewClient(login, secret string) *Client {
	apiEndpoint, _ := url.Parse(APIBaseURL)
	authEndpoint, _ := url.Parse(AuthBaseURL)

	return &Client{
		keyID:        login,
		secret:       secret,
		APIEndpoint:  apiEndpoint,
		AuthEndpoint: authEndpoint,
		HTTPClient:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetZones(ctx context.Context, parentID string) ([]Zone, error) {
	endpoint := c.APIEndpoint.JoinPath("zones")

	query := endpoint.Query()
	query.Set("parentId", parentID)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zones APIResponse[Zone]
	err = c.do(req, &zones)
	if err != nil {
		return nil, err
	}

	return zones.Items, nil
}

func (c *Client) GetRecords(ctx context.Context, zoneID string) ([]Record, error) {
	endpoint := c.APIEndpoint.JoinPath("zones", zoneID, "records")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records APIResponse[Record]
	err = c.do(req, &records)
	if err != nil {
		return nil, err
	}

	return records.Items, nil
}

func (c *Client) CreateRecord(ctx context.Context, zoneID string, record Record) (*Record, error) {
	endpoint := c.APIEndpoint.JoinPath("zones", zoneID, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var result Record
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) DeleteRecord(ctx context.Context, zoneID, name, recordType string) error {
	endpoint := c.APIEndpoint.JoinPath("zones", zoneID, "records", name, recordType)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	tok := getToken(req.Context())
	if tok != nil {
		req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	} else {
		return fmt.Errorf("not logged in")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if result == nil {
		return nil
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
