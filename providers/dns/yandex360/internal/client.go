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
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api360.yandex.net/"

type Client struct {
	oauthToken string
	orgID      int64

	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(oauthToken string, orgID int64) (*Client, error) {
	if oauthToken == "" {
		return nil, errors.New("OAuth token is required")
	}

	if orgID == 0 {
		return nil, errors.New("orgID is required")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		oauthToken: oauthToken,
		orgID:      orgID,
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// AddRecord Adds a DNS record.
// POST https://api30.yandex.net/directory/v1/org/{orgId}/domains/{domain}/dns
// https://yandex.ru/dev/api360/doc/ref/DomainDNSService/DomainDNSService_Create.html
func (c Client) AddRecord(ctx context.Context, domain string, record Record) (*Record, error) {
	endpoint := c.baseURL.JoinPath("directory", "v1", "org", strconv.FormatInt(c.orgID, 10), "domains", domain, "dns")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return nil, err
	}

	var newRecord Record

	err = c.do(req, &newRecord)
	if err != nil {
		return nil, err
	}

	return &newRecord, nil
}

// DeleteRecord Deletes a DNS record.
// DELETE https://api360.yandex.net/directory/v1/org/{orgId}/domains/{domain}/dns/{recordId}
// https://yandex.ru/dev/api360/doc/ref/DomainDNSService/DomainDNSService_Delete.html
func (c Client) DeleteRecord(ctx context.Context, domain string, recordID int64) error {
	endpoint := c.baseURL.JoinPath("directory", "v1", "org", strconv.FormatInt(c.orgID, 10), "domains", domain, "dns", strconv.FormatInt(recordID, 10))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c Client) do(req *http.Request, result any) error {
	req.Header.Set("Authorization", "OAuth "+c.oauthToken)

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

	var apiErr APIError
	err := json.Unmarshal(raw, &apiErr)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return fmt.Errorf("[status code: %d] %w", resp.StatusCode, apiErr)
}
