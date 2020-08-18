// Package dnsmadeeasy implements a DNS provider for solving the DNS-01 challenge using DNS Made Easy.
package dnsmadeeasy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/dnsmadeeasy/internal"
)

// Environment variables names.
const (
	envNamespace = "DNSMADEEASY_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"
	EnvSandbox   = envNamespace + "SANDBOX"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	APIKey             string
	APISecret          string
	Sandbox            bool
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 10*time.Second),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for DNSMadeEasy DNS.
// Credentials must be passed in the environment variables:
// DNSMADEEASY_API_KEY and DNSMADEEASY_API_SECRET.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsmadeeasy: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.Sandbox = env.GetOrDefaultBool(conf, EnvSandbox, false)
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNS Made Easy.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsmadeeasy: the configuration of the DNS provider is nil")
	}

	var baseURL string
	if config.Sandbox {
		baseURL = "https://api.sandbox.dnsmadeeasy.com/V2.0"
	} else {
		if len(config.BaseURL) > 0 {
			baseURL = config.BaseURL
		} else {
			baseURL = "https://api.dnsmadeeasy.com/V2.0"
		}
	}

	client, err := internal.NewClient(config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsmadeeasy: %w", err)
	}

	client.HTTPClient = config.HTTPClient
	client.BaseURL = baseURL

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domainName, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domainName, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to find zone for %s: %w", fqdn, err)
	}

	// fetch the domain details
	domain, err := d.client.GetDomain(authZone)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to get domain for zone %s: %w", authZone, err)
	}

	// create the TXT record
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	record := &internal.Record{Type: "TXT", Name: name, Value: value, TTL: d.config.TTL}

	err = d.client.CreateRecord(domain, record)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to create record for %s: %w", name, err)
	}
	return nil
}

// CleanUp removes the TXT records matching the specified parameters.
func (d *DNSProvider) CleanUp(domainName, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domainName, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to find zone for %s: %w", fqdn, err)
	}

	// fetch the domain details
	domain, err := d.client.GetDomain(authZone)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to get domain for zone %s: %w", authZone, err)
	}

	// find matching records
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	records, err := d.client.GetRecords(domain, name, "TXT")
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to get records for domain %s: %w", domain.Name, err)
	}

	// delete records
	var lastError error
	for _, record := range *records {
		err = d.client.DeleteRecord(record)
		if err != nil {
			lastError = fmt.Errorf("dnsmadeeasy: unable to delete record [id=%d, name=%s]: %w", record.ID, record.Name, err)
		}
	}

	return lastError
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
