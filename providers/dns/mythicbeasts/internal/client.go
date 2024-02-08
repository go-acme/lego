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
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// Default API endpoints.
const (
	APIBaseURL  = "https://api.mythic-beasts.com/dns/v2"
	AuthBaseURL = "https://auth.mythic-beasts.com/login"
)

// Client the Mythic Beasts API client.
type Client struct {
	username string
	password string

	APIEndpoint  *url.URL
	AuthEndpoint *url.URL
	HTTPClient   *http.Client

	token   *Token
	muToken sync.Mutex
}

// NewClient Creates a new Client.
func NewClient(username string, password string) *Client {
	apiEndpoint, _ := url.Parse(APIBaseURL)
	authEndpoint, _ := url.Parse(AuthBaseURL)

	return &Client{
		username:     username,
		password:     password,
		APIEndpoint:  apiEndpoint,
		AuthEndpoint: authEndpoint,
		HTTPClient:   &http.Client{Timeout: 5 * time.Second},
	}
}

// CreateTXTRecord creates a TXT record.
// https://www.mythic-beasts.com/support/api/dnsv2#ep-get-zoneszonerecords
func (c *Client) CreateTXTRecord(ctx context.Context, zone, leaf, value string, ttl int) error {
	resp, err := c.createTXTRecord(ctx, zone, leaf, "TXT", value, ttl)
	if err != nil {
		return err
	}

	if resp.Added != 1 {
		return fmt.Errorf("did not add TXT record for some reason: %s", resp.Message)
	}

	// Success
	return nil
}

// RemoveTXTRecord removes a TXT records.
// https://www.mythic-beasts.com/support/api/dnsv2#ep-delete-zoneszonerecords
func (c *Client) RemoveTXTRecord(ctx context.Context, zone, leaf, value string) error {
	resp, err := c.removeTXTRecord(ctx, zone, leaf, "TXT", value)
	if err != nil {
		return err
	}

	if resp.Removed != 1 {
		return fmt.Errorf("did not remove TXT record for some reason: %s", resp.Message)
	}

	// Success
	return nil
}

// https://www.mythic-beasts.com/support/api/dnsv2#ep-post-zoneszonerecords
func (c *Client) createTXTRecord(ctx context.Context, zone, leaf, recordType, value string, ttl int) (*createTXTResponse, error) {
	endpoint := c.APIEndpoint.JoinPath("zones", zone, "records", leaf, recordType)

	createReq := createTXTRequest{
		Records: []createTXTRecord{{
			Host: leaf,
			TTL:  ttl,
			Type: "TXT",
			Data: value,
		}},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, createReq)
	if err != nil {
		return nil, err
	}

	resp := &createTXTResponse{}
	err = c.do(req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// https://www.mythic-beasts.com/support/api/dnsv2#ep-delete-zoneszonerecords
func (c *Client) removeTXTRecord(ctx context.Context, zone, leaf, recordType, value string) (*deleteTXTResponse, error) {
	endpoint := c.APIEndpoint.JoinPath("zones", zone, "records", leaf, recordType)

	query := endpoint.Query()
	query.Add("data", value)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp := &deleteTXTResponse{}

	err = c.do(req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) do(req *http.Request, result any) error {
	tok := getToken(req.Context())
	if tok != nil {
		req.Header.Set("Authorization", "Bearer "+tok.Token)
	} else {
		return errors.New("not logged in")
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
