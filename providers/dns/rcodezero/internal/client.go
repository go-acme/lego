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
	"path"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/miekg/dns"
)

// Client for the RcodeZero API.
type Client struct {
	apiToken   string
	Host       *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(host *url.URL, apiToken string) *Client {
	return &Client{
		apiToken:   apiToken,
		Host:       host,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) UpdateRecords(ctx context.Context, authZone string, sets []UpdateRRSet) error {
	endpoint := c.joinPath("/", "zones", strings.TrimSuffix(dns.Fqdn(authZone), "."), "/", "rrsets")

	req, err := newJSONRequest(ctx, http.MethodPatch, endpoint, sets)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) joinPath(elem ...string) *url.URL {
	p := path.Join(elem...)

	return c.Host.JoinPath(p)
}

func (c *Client) do(req *http.Request) (json.RawMessage, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnprocessableEntity && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	var msg json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// empty body
			return nil, nil
		}
		// other error
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return msg, nil
	}
	// check for error message
	if len(msg) > 0 && msg[0] == '{' {
		var apiResp apiResponse
		err = json.Unmarshal(msg, &apiResp)
		if err != nil {
			return nil, errutils.NewUnmarshalError(req, resp.StatusCode, msg, err)
		}
		return nil, apiResp
	}

	return msg, nil
}

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(endpoint.String(), "/"), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
