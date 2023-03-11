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
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"golang.org/x/oauth2"
)

// DefaultBaseURL Default API endpoint.
const DefaultBaseURL = "https://api.infomaniak.com"

// Client the Infomaniak client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// New Creates a new Infomaniak client.
func New(hc *http.Client, apiEndpoint string) (*Client, error) {
	baseURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}

	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Second}
	}

	return &Client{baseURL: baseURL, httpClient: hc}, nil
}

func (c *Client) CreateDNSRecord(ctx context.Context, domain *DNSDomain, record Record) (string, error) {
	endpoint := c.baseURL.JoinPath("1", "domain", strconv.FormatUint(domain.ID, 10), "dns", "record")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	result := APIResponse[string]{}
	err = c.do(req, &result)
	if err != nil {
		return "", err
	}

	return result.Data, err
}

func (c *Client) DeleteDNSRecord(ctx context.Context, domainID uint64, recordID string) error {
	endpoint := c.baseURL.JoinPath("1", "domain", strconv.FormatUint(domainID, 10), "dns", "record", recordID)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, &APIResponse[json.RawMessage]{})
}

// GetDomainByName gets a Domain object from its name.
func (c *Client) GetDomainByName(ctx context.Context, name string) (*DNSDomain, error) {
	name = dns01.UnFqdn(name)

	// Try to find the most specific domain
	// starts with the FQDN, then remove each left label until we have a match
	for {
		i := strings.Index(name, ".")
		if i == -1 {
			break
		}

		domain, err := c.getDomainByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if domain != nil {
			return domain, nil
		}

		log.Infof("domain %q not found, trying with %q", name, name[i+1:])

		name = name[i+1:]
	}

	return nil, fmt.Errorf("domain not found %s", name)
}

func (c *Client) getDomainByName(ctx context.Context, name string) (*DNSDomain, error) {
	endpoint := c.baseURL.JoinPath("1", "product")

	query := endpoint.Query()
	query.Add("service_name", "domain")
	query.Add("customer_name", name)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := APIResponse[[]DNSDomain]{}
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	for _, domain := range result.Data {
		if domain.CustomerName == name {
			return &domain, nil
		}
	}

	return nil, nil
}

func (c *Client) do(req *http.Request, result Response) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if err := json.Unmarshal(raw, result); err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if result.GetResult() != "success" {
		return fmt.Errorf("%d: unexpected API result (%s): %w", resp.StatusCode, result.GetResult(), result.GetError())
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
