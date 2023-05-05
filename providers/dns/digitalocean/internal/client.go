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

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/oauth2"
)

// DefaultBaseURL default API endpoint.
const DefaultBaseURL = "https://api.digitalocean.com"

// Client the Digital Ocean API client.
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client.
func NewClient(hc *http.Client) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Second}
	}

	return &Client{BaseURL: baseURL, httpClient: hc}
}

func (c *Client) AddTxtRecord(ctx context.Context, zone string, record Record) (*TxtRecordResponse, error) {
	endpoint := c.BaseURL.JoinPath("v2", "domains", dns01.UnFqdn(zone), "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	respData := &TxtRecordResponse{}
	err = c.do(req, respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

func (c *Client) RemoveTxtRecord(ctx context.Context, zone string, recordID int) error {
	endpoint := c.BaseURL.JoinPath("v2", "domains", dns01.UnFqdn(zone), "records", strconv.Itoa(recordID))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
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

	// NOTE: Even though the body is empty, DigitalOcean API docs still show setting this Content-Type...
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var errInfo APIError
	err := json.Unmarshal(raw, &errInfo)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code %d] %w", resp.StatusCode, errInfo)
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
