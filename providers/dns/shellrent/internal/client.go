package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// DefaultBaseURL the default API endpoint.
const defaultBaseURL = "https://manager.shellrent.com/api2"

const authorizationHeader = "Authorization"

// Client the Shellrent API client.
type Client struct {
	username string
	token    string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(username, token string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		token:      token,
		username:   username,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ListServices lists service IDs.
// https://api.shellrent.com/elenco-dei-servizi-acquistati
func (c Client) ListServices(ctx context.Context) ([]int, error) {
	endpoint := c.baseURL.JoinPath("purchase")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := Response[[]IntOrString]{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, result.Base
	}

	var ids []int

	for _, datum := range result.Data {
		ids = append(ids, datum.Value())
	}

	return ids, nil
}

// GetServiceDetails gets service details.
// https://api.shellrent.com/dettagli-servizio-acquistato
func (c Client) GetServiceDetails(ctx context.Context, serviceID int) (*ServiceDetails, error) {
	endpoint := c.baseURL.JoinPath("purchase", "details", strconv.Itoa(serviceID))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := Response[*ServiceDetails]{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, result.Base
	}

	return result.Data, nil
}

// GetDomainDetails gets domain details.
// https://api.shellrent.com/dettagli-dominio
func (c Client) GetDomainDetails(ctx context.Context, domainID int) (*DomainDetails, error) {
	endpoint := c.baseURL.JoinPath("domain", "details", strconv.Itoa(domainID))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := Response[*DomainDetails]{}

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, result.Base
	}
	return result.Data, nil
}

// CreateRecord created a record.
// https://api.shellrent.com/creazione-record-dns-di-un-dominio
func (c Client) CreateRecord(ctx context.Context, domainID int, record Record) (int, error) {
	endpoint := c.baseURL.JoinPath("dns_record", "store", strconv.Itoa(domainID))

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return 0, err
	}

	result := Response[*Record]{}

	err = c.do(req, &result)
	if err != nil {
		return 0, err
	}

	if result.Code != 0 {
		return 0, result.Base
	}
	return result.Data.ID.Value(), nil
}

// DeleteRecord deletes a record.
// https://api.shellrent.com/eliminazione-record-dns-di-un-dominio
func (c Client) DeleteRecord(ctx context.Context, domainID, recordID int) error {
	endpoint := c.baseURL.JoinPath("dns_record", "remove", strconv.Itoa(domainID), strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	result := Response[any]{}

	err = c.do(req, &result)
	if err != nil {
		return err
	}

	if result.Code != 0 {
		return result.Base
	}

	return nil
}

func (c Client) do(req *http.Request, result any) error {
	req.Header.Set(authorizationHeader, c.username+"."+c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
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

	var response Base
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return response
}

// TTLRounder rounds the given TTL in seconds to the next accepted value.
// Accepted TTL values are:
//   - 3600
//   - 14400
//   - 28800
//   - 57600
//   - 86400
func TTLRounder(ttl int) int {
	for _, validTTL := range []int{3600, 14400, 28800, 57600, 86400} {
		if ttl <= validTTL {
			return validTTL
		}
	}

	return 3600
}
