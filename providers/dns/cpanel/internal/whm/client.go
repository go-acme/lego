package whm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/shared"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const statusFailed = 0

type Client struct {
	username string
	token    string

	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(baseURL string, username string, token string) (*Client, error) {
	apiEndpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		username:   username,
		token:      token,
		baseURL:    apiEndpoint.JoinPath("json-api"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// FetchZoneInformation fetches zone information.
// https://api.docs.cpanel.net/openapi/whm/operation/parse_dns_zone/
func (c Client) FetchZoneInformation(ctx context.Context, domain string) ([]shared.ZoneRecord, error) {
	endpoint := c.baseURL.JoinPath("parse_dns_zone")

	query := endpoint.Query()
	query.Set("zone", domain)
	endpoint.RawQuery = query.Encode()

	var result APIResponse[ZoneData]

	err := c.doRequest(ctx, endpoint, &result)
	if err != nil {
		return nil, err
	}

	if result.Metadata.Result == statusFailed {
		return nil, toError(result.Metadata)
	}

	return result.Data.Payload, nil
}

// AddRecord adds a new record.
//
//	add='{"dname":"example", "ttl":14400, "record_type":"TXT", "data":["string1", "string2"]}'
func (c Client) AddRecord(ctx context.Context, serial uint32, domain string, record shared.Record) (*shared.ZoneSerial, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON data: %w", err)
	}

	return c.updateZone(ctx, serial, domain, "add", string(data))
}

// EditRecord edits an existing record.
//
//	edit='{"line_index": 9, "dname":"example", "ttl":14400, "record_type":"TXT", "data":["string1", "string2"]}'
func (c Client) EditRecord(ctx context.Context, serial uint32, domain string, record shared.Record) (*shared.ZoneSerial, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON data: %w", err)
	}

	return c.updateZone(ctx, serial, domain, "edit", string(data))
}

// DeleteRecord deletes an existing record.
//
//	remove=22
func (c Client) DeleteRecord(ctx context.Context, serial uint32, domain string, lineIndex int) (*shared.ZoneSerial, error) {
	return c.updateZone(ctx, serial, domain, "remove", strconv.Itoa(lineIndex))
}

// https://api.docs.cpanel.net/openapi/whm/operation/mass_edit_dns_zone/
func (c Client) updateZone(ctx context.Context, serial uint32, domain, action, data string) (*shared.ZoneSerial, error) {
	endpoint := c.baseURL.JoinPath("mass_edit_dns_zone")

	query := endpoint.Query()
	query.Set("serial", strconv.FormatUint(uint64(serial), 10))
	query.Set(action, data)
	query.Set("zone", domain)
	endpoint.RawQuery = query.Encode()

	var result APIResponse[shared.ZoneSerial]

	err := c.doRequest(ctx, endpoint, &result)
	if err != nil {
		return nil, err
	}

	if result.Metadata.Result == statusFailed {
		return nil, toError(result.Metadata)
	}

	return &result.Data, nil
}

func (c Client) doRequest(ctx context.Context, endpoint *url.URL, result any) error {
	query := endpoint.Query()
	query.Set("api.version", "1")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	// https://api.docs.cpanel.net/whm/tokens/
	req.Header.Set("Authorization", fmt.Sprintf("whm %s:%s", c.username, c.token))
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
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
