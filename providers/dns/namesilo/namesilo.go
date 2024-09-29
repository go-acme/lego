// Package namesilo implements a DNS provider for solving the DNS-01 challenge using namesilo DNS.
package namesilo

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nrdcg/namesilo"
)

const (
	defaultTTL = 3600
	maxTTL     = 2592000
)

// Environment variables names.
const (
	envNamespace = "NAMESILO_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *namesilo.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for namesilo.
// API_KEY must be passed in the environment variables: NAMESILO_API_KEY.
//
// See: https://www.namesilo.com/api_reference.php
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("namesilo: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Namesilo.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namesilo: the configuration of the DNS provider is nil")
	}

	if config.TTL < defaultTTL || config.TTL > maxTTL {
		return nil, fmt.Errorf("namesilo: TTL should be in [%d, %d]", defaultTTL, maxTTL)
	}

	transport, err := namesilo.NewTokenTransport(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("namesilo: %w", err)
	}

	return &DNSProvider{client: namesilo.NewClient(transport.Client()), config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("namesilo: could not find zone for domain %q: %w", domain, err)
	}

	zoneName := dns01.UnFqdn(zone)

	subdomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zoneName)
	if err != nil {
		return fmt.Errorf("namesilo: %w", err)
	}

	_, err = d.client.DnsAddRecord(&namesilo.DnsAddRecordParams{
		Domain: zoneName,
		Type:   "TXT",
		Host:   subdomain,
		Value:  info.Value,
		TTL:    d.config.TTL,
	})
	if err != nil {
		return fmt.Errorf("namesilo: failed to add record %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("namesilo: could not find zone for domain %q: %w", domain, err)
	}

	zoneName := dns01.UnFqdn(zone)

	resp, err := d.client.DnsListRecords(&namesilo.DnsListRecordsParams{Domain: zoneName})
	if err != nil {
		return fmt.Errorf("namesilo: %w", err)
	}

	subdomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zoneName)
	if err != nil {
		return fmt.Errorf("namesilo: %w", err)
	}

	for _, r := range resp.Reply.ResourceRecord {
		if r.Type == "TXT" && r.Value == info.Value && (r.Host == subdomain || r.Host == dns01.UnFqdn(info.EffectiveFQDN)) {
			_, err := d.client.DnsDeleteRecord(&namesilo.DnsDeleteRecordParams{Domain: zoneName, ID: r.RecordID})
			if err != nil {
				return fmt.Errorf("namesilo: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("namesilo: no TXT record to delete for %s (%s)", info.EffectiveFQDN, info.Value)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
