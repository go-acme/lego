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
	"strings"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://clouddns.manageengine.com/v1"

// Client the ManageEngine CloudDNS API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		baseURL:    baseURL,
		httpClient: hc,
	}
}

// GetAllZones gets all zones.
// https://pitstop.manageengine.com/portal/en/kb/articles/manageengine-clouddns-rest-api-documentation#GET_All
func (c *Client) GetAllZones(ctx context.Context) ([]Zone, error) {
	endpoint := c.baseURL.JoinPath("dns", "domain")

	req, err := newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var results []Zone

	err = c.do(req, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetAllZoneRecords gets all "zone records" for a zone.
// https://pitstop.manageengine.com/portal/en/kb/articles/manageengine-clouddns-rest-api-documentation#GET_All_9
func (c *Client) GetAllZoneRecords(ctx context.Context, zoneID int) ([]ZoneRecord, error) {
	endpoint := c.baseURL.JoinPath("dns", "domain", strconv.Itoa(zoneID), "records", "SPF_TXT")

	req, err := newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var results []ZoneRecord

	err = c.do(req, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteZoneRecord deletes a "zone record".
// https://pitstop.manageengine.com/portal/en/kb/articles/manageengine-clouddns-rest-api-documentation#DEL_Delete_10
func (c *Client) DeleteZoneRecord(ctx context.Context, zoneID, domainID int) error {
	endpoint := c.baseURL.JoinPath("dns", "domain", strconv.Itoa(zoneID), "records", "SPF_TXT", strconv.Itoa(domainID))

	req, err := newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	var results APIResponse

	return c.do(req, &results)
}

// CreateZoneRecord creates a "zone record".
// https://pitstop.manageengine.com/portal/en/kb/articles/manageengine-clouddns-rest-api-documentation#POST_Create_10
func (c *Client) CreateZoneRecord(ctx context.Context, zoneID int, record ZoneRecord) error {
	endpoint := c.baseURL.JoinPath("dns", "domain", strconv.Itoa(zoneID), "records", "SPF_TXT", "/")

	req, err := newRequest(ctx, http.MethodPost, endpoint, []ZoneRecord{record})
	if err != nil {
		return err
	}

	var results APIResponse

	return c.do(req, &results)
}

// UpdateZoneRecord update an existing "zone record".
// https://pitstop.manageengine.com/portal/en/kb/articles/manageengine-clouddns-rest-api-documentation#PUT_Update_10
func (c *Client) UpdateZoneRecord(ctx context.Context, record ZoneRecord) error {
	if record.SpfTxtDomainID == 0 {
		return errors.New("SpfTxtDomainID is empty")
	}
	if record.ZoneID == 0 {
		return errors.New("ZoneID is empty")
	}

	endpoint := c.baseURL.JoinPath("dns", "domain", strconv.Itoa(record.ZoneID), "records", "SPF_TXT", strconv.Itoa(record.SpfTxtDomainID), "/")

	req, err := newRequest(ctx, http.MethodPut, endpoint, []ZoneRecord{record})
	if err != nil {
		return err
	}

	var results APIResponse

	return c.do(req, &results)
}

func (c *Client) do(req *http.Request, result any) error {
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

func newRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	var body io.Reader = http.NoBody

	if payload != nil {
		buf := new(bytes.Buffer)

		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}

		values := url.Values{}
		values.Set("config", buf.String())
		body = strings.NewReader(values.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, &errAPI)
}
