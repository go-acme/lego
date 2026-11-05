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
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const defaultBaseURL = "https://domain.chief.app/api/v1/"

const chiefTeamHeader = "X-Chief-Team"

// Client the Domain Chief API client.
type Client struct {
	apiToken string
	teamID   string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiToken, teamID string) (*Client, error) {
	if apiToken == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiToken:   apiToken,
		teamID:     teamID,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// CreateRecord create a single DNS record for the domain.
// https://docs.chief.tools/api/domainchief/dns#create-dns-record
func (c *Client) CreateRecord(ctx context.Context, domain string, record Record) (*Record, error) {
	endpoint := c.BaseURL.JoinPath("domains", domain, "dns", "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var result APIEnvelope[*Record]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// DeleteRecord deletes a single DNS record for the domain.
// https://docs.chief.tools/api/domainchief/dns#delete-dns-record
func (c *Client) DeleteRecord(ctx context.Context, domain, recordID string) error {
	endpoint := c.BaseURL.JoinPath("domains", domain, "dns", "records", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Accept-Language", "en")

	req.Header.Set("Authorization", "bearer "+c.apiToken)

	if c.teamID != "" {
		req.Header.Set(chiefTeamHeader, c.teamID)
	}

	resp, err := c.HTTPClient.Do(req)
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

	errAPI := new(APIError)

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}
