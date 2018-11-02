package selectel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Domain represents domain name in Selectel Domain API.
type Domain struct {
	ID int `json:"id,omitempty"`

	Name string `json:"name,omitempty"`
}

// Record represents DNS record in Selectel Domain API.
type Record struct {
	ID int `json:"id,omitempty"`

	Name string `json:"name,omitempty"`

	// Record type (SOA, NS, A/AAAA, CNAME, SRV, MX, TXT, SPF)
	Type string `json:"type,omitempty"`

	TTL int `json:"ttl,omitempty"`

	// Email of domain's admin (only for SOA records)
	Email string `json:"email,omitempty"`

	// Record content (not for SRV)
	Content string `json:"content,omitempty"`
}

// Client represents Selectel DNS client.
type Client struct {
	baseURL string

	token string

	userAgent string

	httpClient *http.Client

	timeout time.Duration
}

// ClientOpts represents options to init Selectel DNS client.
type ClientOpts struct {
	BaseURL string

	Token string

	UserAgent string

	HTTPClient *http.Client

	Timeout time.Duration
}

// NewSelectelDNS returns Selectel DNS client instance.
func NewSelectelDNS(opts ClientOpts) *Client {

	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{}
	}

	c := &Client{
		token:      opts.Token,
		baseURL:    opts.BaseURL,
		httpClient: opts.HTTPClient,
		userAgent:  opts.UserAgent,
		timeout:    opts.Timeout,
	}

	return c
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body with error: %s", err)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create new http request with error: %s", err)
	}

	req = req.WithContext(ctx)

	req.Header.Add("X-Token", c.token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, to interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed with error: %s", err)
	}

	if resp.StatusCode >= http.StatusBadRequest &&
		resp.StatusCode <= http.StatusNetworkAuthenticationRequired {

		if resp.Body != nil {
			body, er := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, er
			}
			defer resp.Body.Close()

			return resp, fmt.Errorf("request failed with status code %d, response body: %s",
				resp.StatusCode,
				string(body))
		}

		return resp,
			fmt.Errorf("request failed with status code %d adn empty body",
				resp.StatusCode)
	}

	if to != nil {
		if err = extractResult(resp, to); err != nil {
			return resp,
				fmt.Errorf("failed to extract request body into interface with error: %s",
					err)
		}
	}

	return resp, nil
}

// extractResult reads response body and unmarshal it to given interface
func extractResult(resp *http.Response, to interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, to)
	return err
}

// GetDomainByName gets Domain object by its name.
func (c *Client) GetDomainByName(domainName string) (*Domain, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req, err := c.newRequest(ctx,
		http.MethodGet,
		fmt.Sprintf("/%s", domainName), nil)
	if err != nil {
		return nil, err
	}

	domain := &Domain{}
	_, err = c.do(req, domain)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

// AddRecord adds Record for given domain.
func (c *Client) AddRecord(domainID int, body Record) (*Record, error) {

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req, err := c.newRequest(ctx,
		http.MethodPost,
		fmt.Sprintf("/%d/records/", domainID), body)
	if err != nil {
		return nil, err
	}

	record := &Record{}
	_, err = c.do(req, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// ListRecords returns list records for specific domain.
func (c *Client) ListRecords(domainID int) ([]*Record, error) {

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req, err := c.newRequest(ctx,
		http.MethodGet,
		fmt.Sprintf("/%d/records/", domainID), nil)
	if err != nil {
		return nil, err
	}

	records := make([]*Record, 0)
	_, err = c.do(req, &records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// DeleteRecord deletes specific record.
func (c *Client) DeleteRecord(domainID, recordID int) error {

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req, err := c.newRequest(ctx,
		http.MethodDelete,
		fmt.Sprintf("/%d/records/%d", domainID, recordID), nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}
