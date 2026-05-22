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
	"strings"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
)

const defaultBaseURL = "https://api.dynadot.com"

// Client the Dynadot RESTful v2 API client.
type Client struct {
	apiKey    string
	apiSecret string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey, apiSecret string) (*Client, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// SetDNS adds DNS records for the specified domain.
// Uses `add_dns_to_current_setting=true` so the request appends records without overwriting existing ones.
// https://www.dynadot.com/domain/api-document
func (c *Client) SetDNS(ctx context.Context, domain string, payload *SetDNSRequest) error {
	endpoint := c.BaseURL.JoinPath("restful", "v2", "domains", domain, "records")

	req, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// RemoveDNS removes DNS records for the specified domain.
// Records are only removed if all fields in the request exactly match existing records.
// https://www.dynadot.com/domain/api-document
func (c *Client) RemoveDNS(ctx context.Context, domain string, payload *RemoveDNSRequest) error {
	endpoint := c.BaseURL.JoinPath("restful", "v2", "domains", domain, "records")

	req, err := c.newJSONRequest(ctx, http.MethodDelete, endpoint, payload)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	// Dynadot mirrors HTTP-style status codes inside the response body and
	// always returns code: 200 on success.
	var envelope APIResponse
	if len(raw) != 0 {
		if jerr := json.Unmarshal(raw, &envelope); jerr != nil {
			return errutils.NewUnmarshalError(req, resp.StatusCode, raw, jerr)
		}
	}

	if resp.StatusCode/100 != 2 || (envelope.Code != 0 && envelope.Code != 200) {
		apiErr := &APIError{
			Code:    envelope.Code,
			Message: envelope.Message,
		}
		if envelope.Error != nil {
			apiErr.Description = envelope.Error.Description
		}

		// Fall back to the HTTP-level error if the body did not include a
		// Dynadot envelope.
		if apiErr.Code == 0 && apiErr.Message == "" {
			return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
		}

		return apiErr
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(raw, result)
}

func (c *Client) newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	var body string

	if payload != nil {
		jsonb, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}

		body = string(jsonb)
		buf = bytes.NewBuffer(jsonb)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Build the path used for signing: path + (?query if present).
	// Note: we always need a leading "/" because that is what Dynadot
	// signs against; url.URL.EscapedPath can drop it when the base URL
	// has no path component.
	pathAndQuery := endpoint.EscapedPath()
	if !strings.HasPrefix(pathAndQuery, "/") {
		pathAndQuery = "/" + pathAndQuery
	}

	if endpoint.RawQuery != "" {
		pathAndQuery += "?" + endpoint.RawQuery
	}

	req.Header.Set("X-Signature", generateSignature(c.apiKey, c.apiSecret, pathAndQuery, "", body))

	return req, nil
}
