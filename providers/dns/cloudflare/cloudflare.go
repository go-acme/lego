// Package cloudflare implements a DNS provider for solving the DNS-01 challenge using cloudflare DNS.
package cloudflare

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/platform/config/env"
)

const (
	minTTL = 120
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AuthEmail string
	AuthKey   string

	AuthToken string
	ZoneToken string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("CLOUDFLARE_TTL", minTTL),
		PropagationTimeout: env.GetOrDefaultSecond("CLOUDFLARE_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("CLOUDFLARE_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("CLOUDFLARE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the challenge.Provider interface
type DNSProvider struct {
	client *metaClient
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Cloudflare.
// Credentials must be passed in as environment variables:
//
// Either provide CLOUDFLARE_EMAIL and CLOUDFLARE_API_KEY,
// or a CLOUDFLARE_DNS_API_TOKEN.
//
// For a more paranoid setup, provide CLOUDFLARE_DNS_API_TOKEN and CLOUDFLARE_ZONE_API_TOKEN.
//
// The email and API key should be avoided, if possible.
// Instead setup a API token with both Zone:Read and DNS:Edit permission, and pass the CLOUDFLARE_DNS_API_TOKEN environment variable.
// You can split the Zone:Read and DNS:Edit permissions across multiple API tokens:
// in this case pass both CLOUDFLARE_ZONE_API_TOKEN and CLOUDFLARE_DNS_API_TOKEN accordingly.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.GetWithFallback(
		[]string{"CLOUDFLARE_EMAIL", "CF_API_EMAIL"},
		[]string{"CLOUDFLARE_API_KEY", "CF_API_KEY"},
	)
	if err != nil {
		var errT error
		values, errT = env.GetWithFallback(
			[]string{"CLOUDFLARE_DNS_API_TOKEN", "CF_DNS_API_TOKEN"},
			[]string{"CLOUDFLARE_ZONE_API_TOKEN", "CF_ZONE_API_TOKEN", "CLOUDFLARE_DNS_API_TOKEN", "CF_DNS_API_TOKEN"},
		)
		if errT != nil {
			return nil, fmt.Errorf("cloudflare: %v or %v", err, errT)
		}
	}

	config := NewDefaultConfig()
	config.AuthEmail = values["CLOUDFLARE_EMAIL"]
	config.AuthKey = values["CLOUDFLARE_API_KEY"]
	config.AuthToken = values["CLOUDFLARE_DNS_API_TOKEN"]
	config.ZoneToken = values["CLOUDFLARE_ZONE_API_TOKEN"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Cloudflare.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("cloudflare: the configuration of the DNS provider is nil")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("cloudflare: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client, err := newClient(config)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: %v", err)
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("cloudflare: %v", err)
	}

	zoneID, err := d.client.ZoneIDByName(authZone)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find zone %s: %v", authZone, err)
	}

	dnsRecord := cloudflare.DNSRecord{
		Type:    "TXT",
		Name:    dns01.UnFqdn(fqdn),
		Content: value,
		TTL:     d.config.TTL,
	}

	response, err := d.client.CreateDNSRecord(zoneID, dnsRecord)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to create TXT record: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("cloudflare: failed to create TXT record: %+v %+v", response.Errors, response.Messages)
	}

	log.Infof("cloudflare: new record for %s, ID %s", domain, response.Result.ID)

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("cloudflare: %v", err)
	}

	zoneID, err := d.client.ZoneIDByName(authZone)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find zone %s: %v", authZone, err)
	}

	dnsRecord := cloudflare.DNSRecord{
		Type: "TXT",
		Name: dns01.UnFqdn(fqdn),
	}

	records, err := d.client.DNSRecords(zoneID, dnsRecord)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find TXT records: %v", err)
	}

	for _, record := range records {
		err = d.client.DeleteDNSRecord(zoneID, record.ID)
		if err != nil {
			log.Printf("cloudflare: failed to delete TXT record: %v", err)
		}
	}

	return nil
}
