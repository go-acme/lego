// Package infomaniak implements a DNS provider for solving the DNS-01 challenge using Infomaniak DNS.
package infomaniak

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Infomaniak API reference: https://api.infomaniak.com/doc
// Create a Token:			 https://manager.infomaniak.com/v3/infomaniak-api

// Environment variables names.
const (
	envNamespace = "INFOMANIAK_"

	EnvEndpoint    = envNamespace + "ENDPOINT"
	EnvAccessToken = envNamespace + "ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Record a DNS record.
type Record struct {
	ID     string `json:"id,omitempty"`
	Source string `json:"source,omitempty"`
	Type   string `json:"type,omitempty"`
	TTL    int    `json:"ttl,omitempty"`
	Target string `json:"target,omitempty"`
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIEndpoint        string
	AccessToken        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

type Client struct {
	apiEndpoint string
	apiToken    string
	Client      *http.Client
}

type APIErrorResponse struct {
	Code        string             `json:"code"`
	Description string             `json:"description,omitempty"`
	Context     map[string]string  `json:"context,omitempty"`
	Errors      []APIErrorResponse `json:"errors,omitempty"`
}

type APIResponse struct {
	Result      string           `json:"result"`
	Data        *json.RawMessage `json:"data,omitempty"`
	ErrResponse APIErrorResponse `json:"error,omitempty"`
}

type DNSDomain struct {
	ID           uint64 `json:"id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
}

func (ik *Client) request(method, path string, body io.Reader) (*APIResponse, error) {
	if path[0] != '/' {
		path = "/" + path
	}
	requestURL := ik.apiEndpoint + path

	client := &http.Client{}

	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ik.apiToken)
	req.Header.Set("Content-Type", "application/json")

	rawResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := json.NewDecoder(rawResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("%s %s response parsing error: %v", method, path, err)
	}

	if resp.Result != "success" {
		return nil, fmt.Errorf("%s %s failed: %v", method, path, resp.ErrResponse)
	}

	return &resp, nil
}

func (ik *Client) get(path string, params url.Values) (*APIResponse, error) {
	base, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if params != nil {
		base.RawQuery = params.Encode()
	}
	return ik.request("GET", base.String(), nil)
}

func (ik *Client) post(path string, body io.Reader) (*APIResponse, error) {
	return ik.request("POST", path, body)
}

func (ik *Client) delete(path string) (*APIResponse, error) {
	return ik.request("DELETE", path, nil)
}

// GetDomainByName gather a Domain object from its name.
func (ik *Client) GetDomainByName(name string) (*DNSDomain, error) {
	// remove trailing . if present
	if strings.HasSuffix(name, ".") {
		name = name[:len(name)-1]
	}

	// Try to find the most specific domain
	// starts with the FQDN, then remove each left label until we have a match
	for {
		i := strings.Index(name, ".")
		if i == -1 {
			break
		}
		params := url.Values{}
		params.Add("service_name", "domain")
		params.Add("customer_name", name)

		resp, err := ik.get("/1/product", params)
		if err != nil {
			return nil, err
		}

		var domains []DNSDomain

		if err = json.Unmarshal(*resp.Data, &domains); err != nil {
			return nil, fmt.Errorf("expected array of Domain, got: %v", string(*resp.Data))
		}

		for _, domain := range domains {
			if domain.CustomerName == name {
				return &domain, nil
			}
		}

		log.Infof("domain `%s` not found, trying with `%s`", name, name[i+1:])
		name = name[i+1:]
	}

	return nil, fmt.Errorf("domain not found %s", name)
}

func (ik *Client) CreateDNSRecord(domain *DNSDomain, source, target, recordType string, ttl int) (*string, error) {
	var recordID *string
	record := Record{Source: source, Target: target, Type: recordType, TTL: ttl}
	rawJSON, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	resp, err := ik.post(fmt.Sprintf("/1/domain/%d/dns/record", domain.ID), bytes.NewBuffer(rawJSON))
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(*resp.Data, &recordID); err != nil {
		return nil, fmt.Errorf("expected record, got: %v", string(*resp.Data))
	}
	return recordID, err
}

func (ik *Client) DeleteDNSRecord(domainID uint64, recordID string) error {
	_, err := ik.delete(fmt.Sprintf("/1/domain/%d/dns/record/%s", domainID, recordID))

	return err
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		APIEndpoint:        env.GetOrDefaultString(EnvEndpoint, "https://api.infomaniak.com"),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL*60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config      *Config
	client      *Client
	recordIDs   map[string]string
	domainIDs   map[string]uint64
	recordIDsMu sync.Mutex
	domainIDsMu sync.Mutex
}

func NewInfomaniakAPI(apiEndpoint string, apiToken string) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		apiToken:    apiToken,
	}
}

// NewDNSProvider returns a DNSProvider instance configured for Infomaniak
// Credentials must be passed in the environment variables:
// INFOMANIAK_ACCESS_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessToken)
	if err != nil {
		return nil, fmt.Errorf("infomaniak: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessToken = values[EnvAccessToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Infomaniak.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("infomaniak: the configuration of the DNS provider is nil")
	}

	if config.APIEndpoint == "" || config.AccessToken == "" {
		return nil, errors.New("infomaniak: credentials missing")
	}

	client := NewInfomaniakAPI(
		config.APIEndpoint,
		config.AccessToken,
	)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
		domainIDs: make(map[string]uint64),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := getZone(fqdn)
	if err != nil {
		return err
	}

	infomaniakDomain, err := d.client.GetDomainByName(domain)
	if err != nil {
		return fmt.Errorf("infomaniak: could not get domain %q: %w", domain, err)
	}

	d.domainIDsMu.Lock()
	d.domainIDs[token] = infomaniakDomain.ID
	d.domainIDsMu.Unlock()

	recordID, err := d.client.CreateDNSRecord(
		infomaniakDomain,
		extractRecordName(fqdn, authZone),
		value,
		"TXT",
		d.config.TTL,
	)
	if err != nil {
		return fmt.Errorf("infomaniak: error when calling api to create DNS record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = *recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("infomaniak: unknown record ID for '%s'", fqdn)
	}

	d.domainIDsMu.Lock()
	domainID, ok := d.domainIDs[token]
	d.domainIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("infomaniak: unknown domain ID for '%s'", fqdn)
	}

	err := d.client.DeleteDNSRecord(domainID, recordID)
	if err != nil {
		return fmt.Errorf("infomaniak: could not delete record %q: %w", domain, err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	// Delete domain ID from map
	d.domainIDsMu.Lock()
	delete(d.domainIDs, token)
	d.domainIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func getZone(fqdn string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	return dns01.UnFqdn(authZone), nil
}

func extractRecordName(fqdn, zone string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+zone); idx != -1 {
		return name[:idx]
	}
	return name
}
