// Package emhostAdds lego support for http://dreamhost.com DNS updates
// See https://help.dreamhost.com/hc/en-us/articles/217560167-API_overview and https://help.dreamhost.com/hc/en-us/articles/217555707-DNS-API-commands for the API spec.
package dreamhost

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		PropagationTimeout: env.GetOrDefaultSecond("DREAMHOST_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DREAMHOST_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DREAMHOST_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider adds and removes the record for the DNS challenge
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider using
// environment variable DREAMHOST_TOKEN for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DREAMHOST_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("dreamhost: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["DREAMHOST_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DreamHost.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dreamhost: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("dreamhost: credentials missing")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	record := acme.UnFqdn(fqdn)

	err := d.updateTxtRecord(cmdAddRecord, record, value)
	if err != nil {
		return fmt.Errorf("dreamhost: %v", err)
	}
	return nil
}

// CleanUp clears DreamHost TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	record := acme.UnFqdn(fqdn)
	return d.updateTxtRecord(cmdRemoveRecord, record, value)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// updateTxtRecord will either add or remove a TXT record.
// action is either cmdAddRecord or cmdRemoveRecord
func (d *DNSProvider) updateTxtRecord(action, domain, txt string) error {

	u, err := url.Parse(d.config.BaseURL)
	if err != nil {
		return err
	}

	query := u.Query()
	query.Set("key", d.config.APIKey)
	query.Set("cmd", action)
	query.Set("format", "json")
	query.Set("record", domain)
	query.Set("type", "TXT")
	query.Set("value", txt)
	query.Set("comment", url.QueryEscape("Managed By lego"))

	u.RawQuery = query.Encode()

	resp, err := d.config.HTTPClient.Get(u.String())

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with HTTP status code %d", resp.StatusCode)
	}

	var response apiResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("unable to decode API server response")
	}

	if response.Result == "error" {
		return fmt.Errorf("dreamhost: add TXT record failed: %s", response.Data)
	}

	log.Infof("dreamhost: %s", response.Data)
	return nil
}
