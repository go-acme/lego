package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const (
	defaultBaseURL = "https://api.dns.constellix.com"
	defaultVersion = "v1"
)

// Client the Constellix client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for communicating with the API
	Domains    *DomainService
	TxtRecords *TxtRecordService
}

// NewClient Creates a Constellix client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	client := &Client{
		BaseURL:    defaultBaseURL,
		HTTPClient: httpClient,
	}

	client.common.client = client
	client.Domains = (*DomainService)(&client.common)
	client.TxtRecords = (*TxtRecordService)(&client.common)

	return client
}

type service struct {
	client *Client
}

// do sends an API request and returns the API response.
func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if err = json.Unmarshal(raw, result); err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func (c *Client) createEndpoint(fragment ...string) (string, error) {
	return url.JoinPath(c.BaseURL, fragment...)
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err == nil && raw != nil {
		errAPI := &APIError{StatusCode: resp.StatusCode}

		if json.Unmarshal(raw, errAPI) != nil {
			return fmt.Errorf("API error: status code: %d: %v", resp.StatusCode, string(raw))
		}

		switch resp.StatusCode {
		case http.StatusNotFound:
			return &NotFound{APIError: errAPI}
		case http.StatusBadRequest:
			return &BadRequest{APIError: errAPI}
		default:
			return errAPI
		}
	}

	return fmt.Errorf("API error, status code: %d", resp.StatusCode)
}
