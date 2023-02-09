package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
)

const defaultBaseURL = "https://api.vercel.com"

// Client Vercel client.
type Client struct {
	authToken  string
	teamID     string
	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a Client.
func NewClient(authToken string, teamID string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		authToken:  authToken,
		teamID:     teamID,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// CreateRecord creates a DNS record.
// https://vercel.com/docs/rest-api#endpoints/dns/create-a-dns-record
func (c *Client) CreateRecord(zone string, record Record) (*CreateRecordResponse, error) {
	endpoint := c.baseURL.JoinPath("v2", "domains", dns01.UnFqdn(zone), "records")

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, readError(req, resp)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	// Everything looks good; but we'll need the ID later to delete the record
	respData := &CreateRecordResponse{}
	err = json.Unmarshal(content, respData)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return respData, nil
}

// DeleteRecord deletes a DNS record.
// https://vercel.com/docs/rest-api#endpoints/dns/delete-a-dns-record
func (c *Client) DeleteRecord(zone string, recordID string) error {
	endpoint := c.baseURL.JoinPath("v2", "domains", dns01.UnFqdn(zone), "records", recordID)

	req, err := c.newRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return readError(req, resp)
	}

	return nil
}

func (c *Client) newRequest(method, reqURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	if c.teamID != "" {
		query := req.URL.Query()
		query.Add("teamId", c.teamID)
		req.URL.RawQuery = query.Encode()
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	return req, nil
}

func readError(req *http.Request, resp *http.Response) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(toUnreadableBodyMessage(req, content))
	}

	var errInfo APIErrorResponse
	err = json.Unmarshal(content, &errInfo)
	if err != nil {
		return fmt.Errorf("API Error unmarshaling error: %w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return fmt.Errorf("HTTP %d: %w", resp.StatusCode, errInfo.Error)
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
