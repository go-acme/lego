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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

// Client the Bluecat v2 API client.
type Client struct {
	username string
	password string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(serverURL, username, password string) (*Client, error) {
	if serverURL == "" {
		return nil, errors.New("server URL missing")
	}

	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		username:   username,
		password:   password,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// RetrieveZones retrieves all zones.
func (c *Client) RetrieveZones(ctx context.Context, opts *CollectionOptions) ([]ZoneResource, error) {
	endpoint := c.baseURL.JoinPath("api", "v2", "zones")

	collection, err := retrieveCollection[ZoneResource](ctx, c, endpoint, opts)
	if err != nil {
		return nil, err
	}

	return collection.Data, nil
}

// RetrieveZoneDeployments retrieves all deployments for a zone.
func (c *Client) RetrieveZoneDeployments(ctx context.Context, zoneID int64, opts *CollectionOptions) ([]QuickDeployment, error) {
	endpoint := c.baseURL.JoinPath("api", "v2", "zones", strconv.FormatInt(zoneID, 10), "deployments")

	collection, err := retrieveCollection[QuickDeployment](ctx, c, endpoint, opts)
	if err != nil {
		return nil, err
	}

	return collection.Data, nil
}

// CreateZoneDeployment creates a new deployment for a zone.
func (c *Client) CreateZoneDeployment(ctx context.Context, zoneID int64) (*QuickDeployment, error) {
	endpoint := c.baseURL.JoinPath("api", "v2", "zones", strconv.FormatInt(zoneID, 10), "deployments")

	payload := CommonResource{
		Type: "QuickDeployment",
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	result := new(QuickDeployment)

	err = c.doAuthenticated(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateZoneResourceRecord creates a new TXT record in a zone.
func (c *Client) CreateZoneResourceRecord(ctx context.Context, zoneID int64, record RecordTXT) (*RecordTXT, error) {
	endpoint := c.baseURL.JoinPath("api", "v2", "zones", strconv.FormatInt(zoneID, 10), "resourceRecords")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	result := new(RecordTXT)

	err = c.doAuthenticated(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteResourceRecord deletes a resource record.
func (c *Client) DeleteResourceRecord(ctx context.Context, recordID int64) error {
	endpoint := c.baseURL.JoinPath("api", "v2", "resourceRecords", strconv.FormatInt(recordID, 10))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.doAuthenticated(ctx, req, nil)
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

func retrieveCollection[T any](ctx context.Context, client *Client, endpoint *url.URL, opts *CollectionOptions) (*Collection[T], error) {
	if opts != nil {
		values, err := querystring.Values(opts)
		if err != nil {
			return nil, err
		}

		endpoint.RawQuery = values.Encode()
	}

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := &Collection[T]{}

	err = client.doAuthenticated(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
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
