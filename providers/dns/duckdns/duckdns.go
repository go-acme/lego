// Package duckdns implements a DNS provider for solving the DNS-01 challenge using DuckDNS.
// See http://www.duckdns.org/spec.jsp for more info on updating TXT records.
package duckdns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("DUCKDNS_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DUCKDNS_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DUCKDNS_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider adds and removes the record for the DNS challenge
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider using
// environment variable DUCKDNS_TOKEN for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DUCKDNS_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("duckdns: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = values["DUCKDNS_TOKEN"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DuckDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("duckdns: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("duckdns: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, txtRecord, _ := dns01.GetRecord(domain, keyAuth)
	return d.updateTxtRecord(domain, d.config.Token, txtRecord, false)
}

// CleanUp clears DuckDNS TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.updateTxtRecord(domain, d.config.Token, "", true)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
