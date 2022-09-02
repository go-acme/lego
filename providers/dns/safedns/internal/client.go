package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
)

const defaultBaseURL = "https://api.ukfast.io/safedns/v1"

// Client the UKFast SafeDNS client.
type Client struct {
	authToken string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new Client.
func NewClient(authToken string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)
	return &Client{
		authToken:  authToken,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddRecord adds a DNS record.
func (c *Client) AddRecord(zone string, record Record) (*AddRecordResponse, error) {
	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "zones", dns01.UnFqdn(zone), "records"))
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

	respData := &AddRecordResponse{}
	err = json.Unmarshal(content, respData)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return respData, nil
}

// RemoveRecord removes a DNS record.
func (c *Client) RemoveRecord(zone string, recordID int) error {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "zones", dns01.UnFqdn(zone), "records", strconv.Itoa(recordID)))
	if err != nil {
		return err
	}

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

func (c *Client) newRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.authToken)

	return req, nil
}

func readError(req *http.Request, resp *http.Response) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(toUnreadableBodyMessage(req, content))
	}

	var errInfo APIError
	err = json.Unmarshal(content, &errInfo)
	if err != nil {
		return fmt.Errorf("unmarshaling error: %w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return errInfo
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s received a response with an invalid format: %q", req.URL, string(rawBody))
}
