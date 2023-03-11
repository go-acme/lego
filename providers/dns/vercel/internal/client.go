package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://api.vercel.com"

// Client Vercel client.
type Client struct {
	teamID string

	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a Client.
func NewClient(hc *http.Client, teamID string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		teamID:     teamID,
		baseURL:    baseURL,
		httpClient: hc,
	}
}

// CreateRecord creates a DNS record.
// https://vercel.com/docs/rest-api#endpoints/dns/create-a-dns-record
func (c *Client) CreateRecord(ctx context.Context, zone string, record Record) (*CreateRecordResponse, error) {
	endpoint := c.baseURL.JoinPath("v2", "domains", dns01.UnFqdn(zone), "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	respData := &CreateRecordResponse{}
	err = c.do(req, respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

// DeleteRecord deletes a DNS record.
// https://vercel.com/docs/rest-api#endpoints/dns/delete-a-dns-record
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordID string) error {
	endpoint := c.baseURL.JoinPath("v2", "domains", dns01.UnFqdn(zone), "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	if c.teamID != "" {
		query := req.URL.Query()
		query.Add("teamId", c.teamID)
		req.URL.RawQuery = query.Encode()
	}

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

	var response APIErrorResponse
	err := json.Unmarshal(raw, &response)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, response.Error)
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
