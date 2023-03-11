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

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// defaultBaseURL is the GleSYS API endpoint used by Present and CleanUp.
const defaultBaseURL = "https://api.glesys.com/"

type Client struct {
	apiUser string
	apiKey  string

	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(apiUser string, apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiUser:    apiUser,
		apiKey:     apiKey,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddTXTRecord adds a dns record to a domain.
// https://github.com/GleSYS/API/wiki/API-Documentation#domainaddrecord
func (c *Client) AddTXTRecord(ctx context.Context, domain, name, value string, ttl int) (int, error) {
	endpoint := c.baseURL.JoinPath("domain", "addrecord")

	request := addRecordRequest{
		DomainName: domain,
		Host:       name,
		Type:       "TXT",
		Data:       value,
		TTL:        ttl,
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, request)
	if err != nil {
		return 0, err
	}

	response, err := c.do(req)
	if err != nil {
		return 0, err
	}

	if response != nil && response.Response.Status.Code == http.StatusOK {
		return response.Response.Record.RecordID, nil
	}

	return 0, err
}

// DeleteTXTRecord removes a dns record from a domain.
// https://github.com/GleSYS/API/wiki/API-Documentation#domaindeleterecord
func (c *Client) DeleteTXTRecord(ctx context.Context, recordID int) error {
	endpoint := c.baseURL.JoinPath("domain", "deleterecord")

	request := deleteRecordRequest{RecordID: recordID}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, request)
	if err != nil {
		return err
	}

	_, err = c.do(req)

	return err
}

func (c *Client) do(req *http.Request) (*apiResponse, error) {
	req.SetBasicAuth(c.apiUser, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var response apiResponse
	err = json.Unmarshal(raw, &response)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return &response, nil
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
