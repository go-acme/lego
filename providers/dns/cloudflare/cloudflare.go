// Package cloudflare implements a DNS provider for solving the DNS-01 challenge using cloudflare DNS.
package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	minTTL = 120
)

// Config is used to configure the creation of the DNSProvider.
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

// NewDefaultConfig returns a default configuration for the DNSProvider.
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

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *metaClient
	config *Config

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
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
// Instead, set up an API token with both Zone:Read and DNS:Edit permission, and pass the CLOUDFLARE_DNS_API_TOKEN environment variable.
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
			//nolint:errorlint
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
		return nil, fmt.Errorf("cloudflare: %w", err)
	}

	return &DNSProvider{
		client:    client,
		config:    config,
		recordIDs: make(map[string]string),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("cloudflare: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	zoneID, err := d.client.ZoneIDByName(authZone)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find zone %s: %w", authZone, err)
	}

	dnsRecord := cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    dns01.UnFqdn(info.EffectiveFQDN),
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	response, err := d.client.CreateDNSRecord(context.Background(), zoneID, dnsRecord)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to create TXT record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = response.ID
	d.recordIDsMu.Unlock()

	log.Infof("cloudflare: new record for %s, ID %s", domain, response.ID)

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("cloudflare: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	zoneID, err := d.client.ZoneIDByName(authZone)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find zone %s: %w", authZone, err)
	}

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("cloudflare: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	err = d.client.DeleteDNSRecord(context.Background(), zoneID, recordID)
	if err != nil {
		log.Printf("cloudflare: failed to delete TXT record: %w", err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}
