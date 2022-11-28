// Package namedotcom implements a DNS provider for solving the DNS-01 challenge using Name.com's DNS service.
package namedotcom

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/namedotcom/go/namecom"
)

// according to https://www.name.com/api-docs/DNS#CreateRecord
const minTTL = 300

// Environment variables names.
const (
	envNamespace = "NAMECOM_"

	EnvUsername = envNamespace + "USERNAME"
	EnvAPIToken = envNamespace + "API_TOKEN"
	EnvServer   = envNamespace + "SERVER"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username           string
	APIToken           string
	Server             string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 15*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *namecom.NameCom
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for namedotcom.
// Credentials must be passed in the environment variables:
// NAMECOM_USERNAME and NAMECOM_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("namedotcom: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.APIToken = values[EnvAPIToken]
	config.Server = env.GetOrFile(EnvServer)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for namedotcom.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namedotcom: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("namedotcom: username is required")
	}

	if config.APIToken == "" {
		return nil, errors.New("namedotcom: API token is required")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("namedotcom: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := namecom.New(config.Username, config.APIToken)
	client.Client = config.HTTPClient

	if config.Server != "" {
		client.Server = config.Server
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	domainDetails, err := d.client.GetDomain(&namecom.GetDomainRequest{DomainName: domain})
	if err != nil {
		return fmt.Errorf("namedotcom: API call failed: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, domainDetails.DomainName)
	if err != nil {
		return fmt.Errorf("namedotcom: %w", err)
	}

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	request := &namecom.Record{
		DomainName: domain,
		Host:       subDomain,
		Type:       "TXT",
		TTL:        uint32(d.config.TTL),
		Answer:     value,
	}

	_, err = d.client.CreateRecord(request)
	if err != nil {
		return fmt.Errorf("namedotcom: API call failed: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	records, err := d.getRecords(domain)
	if err != nil {
		return fmt.Errorf("namedotcom: %w", err)
	}

	for _, rec := range records {
		if rec.Fqdn == fqdn && rec.Type == "TXT" {
			// TODO(ldez) replace domain by FQDN to follow CNAME.
			request := &namecom.DeleteRecordRequest{
				DomainName: domain,
				ID:         rec.ID,
			}
			_, err := d.client.DeleteRecord(request)
			if err != nil {
				return fmt.Errorf("namedotcom: %w", err)
			}
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getRecords(domain string) ([]*namecom.Record, error) {
	request := &namecom.ListRecordsRequest{
		DomainName: domain,
		Page:       1,
	}

	var records []*namecom.Record
	for request.Page > 0 {
		response, err := d.client.ListRecords(request)
		if err != nil {
			return nil, err
		}

		records = append(records, response.Records...)
		request.Page = response.NextPage
	}

	return records, nil
}
