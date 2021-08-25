package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
)

// Client the Infomaniak client.
type Client struct {
	apiEndpoint string
	apiToken    string
	HTTPClient  *http.Client
}

// New Creates a new Infomaniak client.
func New(apiEndpoint, apiToken string) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		apiToken:    apiToken,
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) CreateDNSRecord(domain *DNSDomain, record Record) (string, error) {
	rawJSON, err := json.Marshal(record)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("/1/domain/%d/dns/record", domain.ID)

	req, err := c.newRequest(http.MethodPost, uri, bytes.NewBuffer(rawJSON))
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	var recordID string
	if err = json.Unmarshal(resp.Data, &recordID); err != nil {
		return "", fmt.Errorf("expected record, got: %s", string(resp.Data))
	}

	return recordID, err
}

func (c *Client) DeleteDNSRecord(domainID uint64, recordID string) error {
	uri := fmt.Sprintf("/1/domain/%d/dns/record/%s", domainID, recordID)

	req, err := c.newRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)

	return err
}

// GetDomainByName gets a Domain object from its name.
func (c *Client) GetDomainByName(name string) (*DNSDomain, error) {
	name = dns01.UnFqdn(name)

	// Try to find the most specific domain
	// starts with the FQDN, then remove each left label until we have a match
	for {
		i := strings.Index(name, ".")
		if i == -1 {
			break
		}

		domain, err := c.getDomainByName(name)
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

func (c *Client) getDomainByName(name string) (*DNSDomain, error) {
	base, err := url.Parse("/1/product")
	if err != nil {
		return nil, err
	}

	query := base.Query()
	query.Add("service_name", "domain")
	query.Add("customer_name", name)
	base.RawQuery = query.Encode()

	req, err := c.newRequest(http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var domains []DNSDomain
	if err = json.Unmarshal(resp.Data, &domains); err != nil {
		return nil, fmt.Errorf("failed to marshal domains: %s", string(resp.Data))
	}

	for _, domain := range domains {
		if domain.CustomerName == name {
			return &domain, nil
		}
	}

	return nil, nil
}

func (c *Client) do(req *http.Request) (*APIResponse, error) {
	rawResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform API request: %w", err)
	}

	defer func() { _ = rawResp.Body.Close() }()

	content, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body, status code: %d", rawResp.StatusCode)
	}

	var resp APIResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the response body: %s", string(content))
	}

	if resp.Result != "success" {
		return nil, fmt.Errorf("%d: unexpected API result (%s): %w", rawResp.StatusCode, resp.Result, resp.ErrResponse)
	}

	return &resp, nil
}

func (c *Client) newRequest(method, uri string, body io.Reader) (*http.Request, error) {
	baseURL, err := url.Parse(c.apiEndpoint)
	if err != nil {
		return nil, err
	}

	endpoint, err := baseURL.Parse(path.Join(baseURL.Path, uri))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
