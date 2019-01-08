// Package zone implements a DNS provider for solving the DNS-01 challenge through zone.ee.
package zone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/platform/config/env"
)

const defaultEndpoint = "https://api.zone.eu/v2/dns"

type message struct {
	FQDN  string `json:"name"`
	Value string `json:"destination"`
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Endpoint           *url.URL
	Username           string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("ZONE_PROPAGATION_TIMEOUT", 5*dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("ZONE_POLLING_INTERVAL", 5*dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("ZONE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider describes a provider for acme-proxy
type DNSProvider struct {
	config    *Config
	recordURL string
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	var endpoint *url.URL
	if s, exists := os.LookupEnv("ZONE_ENDPOINT"); exists {
		var err error
		endpoint, err = url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("zone: %v", err)
		}
	} else {
		endpoint, _ = url.Parse(defaultEndpoint)
	}

	config := NewDefaultConfig()
	config.Username = os.Getenv("ZONE_USERNAME")
	config.APIKey = os.Getenv("ZONE_APIKEY")
	config.Endpoint = endpoint
	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider .
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("zone: the configuration of the DNS provider is nil")
	}

	if config.Endpoint == nil {
		return nil, errors.New("zone: the endpoint is missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	msg := &message{
		FQDN:  fqdn[:len(fqdn)-1],
		Value: value,
	}

	endpoint := strings.TrimSuffix(d.config.Endpoint.String(), "/")
	uri := fmt.Sprintf("%s/%s/txt", endpoint, domain)

	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(msg); err != nil {
		return fmt.Errorf("zone: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, reqBody)
	if err != nil {
		return fmt.Errorf("zone: %v", err)
	}

	if _, err = d.doRequest(req); err != nil {
		return fmt.Errorf("zone: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record previously created
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value := dns01.GetRecord(domain, keyAuth)

	endpoint := strings.TrimSuffix(d.config.Endpoint.String(), "/")
	uri := fmt.Sprintf("%s/%s/txt", endpoint, domain)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return fmt.Errorf("zone: %v", err)
	}

	m, err := d.doRequest(req)
	if err != nil {
		return fmt.Errorf("zone: %v", err)
	}

	var deleteURL string
	for _, e := range m {
		if e["destination"].(string) == value {
			deleteURL = e["resource_url"].(string)
		}
	}

	if deleteURL == "" {
		return fmt.Errorf("zone: txt record does not exist for %v", value)
	}

	u, err := url.Parse(deleteURL)
	if err != nil {
		return fmt.Errorf("zone: server returned an invalid url: %v (%v)", deleteURL, err)
	}

	uri = fmt.Sprintf("%s/%s", endpoint, strings.TrimLeft(u.Path, "/v2/dns"))

	req, err = http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("zone: %v", err)
	}

	if _, err = d.doRequest(req); err != nil {
		return fmt.Errorf("zone: %v", err)
	}

	return nil
}

func (d *DNSProvider) doRequest(req *http.Request) ([]map[string]interface{}, error) {
	req.Header.Set("Content-Type", "application/json")

	if len(d.config.Username) > 0 && len(d.config.APIKey) > 0 {
		req.SetBasicAuth(d.config.Username, d.config.APIKey)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%d: failed to read response body: %v", resp.StatusCode, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%d: request failed: %v", resp.StatusCode, string(body))
	}

	if req.Method == http.MethodDelete {
		return nil, nil
	}

	var m []map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("%d: failed to decode request: %v", resp.StatusCode, string(body))
	}

	return m, nil
}
