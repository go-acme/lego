package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const defaultBaseURL = "https://api.ns1.hosttech.eu/api"

// Client a Hosttech client.
type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL

	apiKey string
}

// NewClient creates a new Client.
func NewClient(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
	}
}

// GetZones Get a list of all zones.
// https://api.ns1.hosttech.eu/api/documentation/#/Zones/get_api_user_v1_zones
func (c Client) GetZones(query string, limit, offset int) ([]Zone, error) {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "user", "v1", "zones"))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	values := endpoint.Query()
	values.Set("query", query)

	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		values.Set("offset", strconv.Itoa(offset))
	}

	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var r []Zone
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response data: %s: %w", string(raw), err)
	}

	return r, nil
}

// GetZone Get a single zone.
// https://api.ns1.hosttech.eu/api/documentation/#/Zones/get_api_user_v1_zones__zoneId_
func (c Client) GetZone(zoneID string) (*Zone, error) {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "user", "v1", "zones", zoneID))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var r Zone
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response data: %s: %w", string(raw), err)
	}

	return &r, nil
}

// GetRecords Returns a list of all records for the given zone.
// https://api.ns1.hosttech.eu/api/documentation/#/Records/get_api_user_v1_zones__zoneId__records
func (c Client) GetRecords(zoneID, recordType string) ([]Record, error) {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "user", "v1", "zones", zoneID, "records"))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	values := endpoint.Query()

	if recordType != "" {
		values.Set("type", recordType)
	}

	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var r []Record
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response data: %s: %w", string(raw), err)
	}

	return r, nil
}

// AddRecord Adds a new record to the zone and returns the newly created record.
// https://api.ns1.hosttech.eu/api/documentation/#/Records/post_api_user_v1_zones__zoneId__records
func (c Client) AddRecord(zoneID string, record Record) (*Record, error) {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "user", "v1", "zones", zoneID, "records"))
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal request data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var r Record
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response data: %s: %w", string(raw), err)
	}

	return &r, nil
}

// DeleteRecord Deletes a single record for the given id.
// https://api.ns1.hosttech.eu/api/documentation/#/Records/delete_api_user_v1_zones__zoneId__records__recordId_
func (c Client) DeleteRecord(zoneID, recordID string) error {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "user", "v1", "zones", zoneID, "records", recordID))
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	_, err = c.do(req)

	return err
}

func (c Client) do(req *http.Request) (json.RawMessage, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, errD := c.HTTPClient.Do(req)
	if errD != nil {
		return nil, fmt.Errorf("send request: %w", errD)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		all, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		var r apiResponse
		err = json.Unmarshal(all, &r)
		if err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}

		return r.Data, nil

	case http.StatusNoContent:
		return nil, nil

	default:
		data, _ := ioutil.ReadAll(resp.Body)

		e := APIError{StatusCode: resp.StatusCode}
		err := json.Unmarshal(data, &e)
		if err != nil {
			e.Message = string(data)
		}

		return nil, e
	}
}
