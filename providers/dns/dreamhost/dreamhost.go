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
	ApiKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	client := acme.HTTPClient
	client.Timeout = env.GetOrDefaultSecond("DREAMHOST_HTTP_TIMEOUT", 30*time.Second)

	return &Config{
		BaseURL:            defaultBaseURL,
		PropagationTimeout: env.GetOrDefaultSecond("DREAMHOST_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DREAMHOST_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient:         &client,
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
	config.ApiKey = values["DREAMHOST_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DreamHost.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dreamhost: the configuration of the DNS provider is nil")
	}

	if config.ApiKey == "" {
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

	err := d.updateTxtRecord("add", d.config.ApiKey, record, value)
	if err != nil {
		return fmt.Errorf("dreamhost: %v", err)
	}
	return nil
}

// CleanUp clears DreamHost TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	record := acme.UnFqdn(fqdn)
	return d.updateTxtRecord("remove", d.config.ApiKey, record, value)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// addTxtRecord adds the TXT record to the domain
func (d *DNSProvider) updateTxtRecord(action, token, domain, txt string) error {
	var cmd string
	switch action {
	case "add":
		cmd = "dns-add_record"
	case "remove":
		cmd = "dns-remove_record"
	default:
		return fmt.Errorf("dreamhost: updateTxtRecord called with invalid action: %s", action)
	}

	comment := url.QueryEscape("Managed By lego")
	u := fmt.Sprintf("%s/?key=%s&cmd=%s&format=json&record=%s&type=TXT&value=%s&comment=%s", d.config.BaseURL, token, cmd, domain, txt, comment)

	resp, err := acme.HTTPClient.Get(u)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with HTTP status code %d", resp.StatusCode)
	}

	var response responseStruct
	err = json.NewDecoder(resp.Body).Decode(&response)
	// return fmt.Errorf(fmt.Sprint(response.Result))
	if err != nil {
		return fmt.Errorf("unable to decode API server response")
	} else if response.Result == "error" {
		return fmt.Errorf("dreamhost: add TXT record failed: %s", response.Data)
	}

	log.Infof("dreamhost: %s", response.Data)
	return nil
}
