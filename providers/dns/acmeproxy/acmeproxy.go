// Package acmeproxy implements a DNS provider for solving the DNS-01 challenge using acme-proxy.
package acmeproxy

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	URL                string
	Provider           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("ACMEPROXY_PROPAGATION_TIMEOUT", 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("ACMEPROXY_POLLING_INTERVAL", 10*time.Second),
	}
}

// DNSProvider describes a provider for acme-proxy
type DNSProvider struct {
	client *http.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for acme-proxy.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("ACMEPROXY_URL", "ACMEPROXY_PROVIDER")
	if err != nil {
		return nil, fmt.Errorf("acmeproxy: %v", err)
	}

	config := NewDefaultConfig()
	config.URL = values["ACMEPROXY_URL"]
	config.Provider = values["ACMEPROXY_PROVIDER"]
	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for acme-proxy.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("acmeproxy: the configuration of the DNS provider is nil")
	}

	client := &http.Client{}

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {

	var jsonStr = []byte(`{"provider":"` + d.config.Provider + `","action":"present","domain": "` + domain + `", "keyauth": "` + keyAuth + `"}`)
	req, err := http.NewRequest("POST", d.config.URL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return fmt.Errorf("acmeproxy: error for %s in Cleanup: %v", domain, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("acmeproxy: error for %s in Present: %v", domain, err)
	}
	defer resp.Body.Close()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {

	var jsonStr = []byte(`{"provider":"` + d.config.Provider + `","action":"cleanup","domain": "` + domain + `", "keyauth": "` + keyAuth + `"}`)
	req, err := http.NewRequest("POST", d.config.URL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return fmt.Errorf("acmeproxy: error for %s in Cleanup: %v", domain, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("acmeproxy: error for %s in Cleanup: %v", domain, err)
	}
	defer resp.Body.Close()

	return nil
}
