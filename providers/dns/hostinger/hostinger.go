// Package hostinger implements a DNS provider for solving the DNS-01 challenge using Hostinger.
package hostinger

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hostinger/internal"
)

// Environment variables names.
const (
	envNamespace = "HOSTINGER_"

	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Hostinger.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("hostinger: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Hostinger.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hostinger: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("hostinger: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hostinger: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("hostinger: %w", err)
	}

	ctx := context.Background()

	recordSets, err := d.client.GetDNSRecords(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("hostinger: get DNS records: %w", err)
	}

	var newRecordSet []internal.RecordSet

	var added bool

	for _, recordSet := range recordSets {
		if recordSet.Name == subDomain && recordSet.Type == "TXT" {
			recordSet.Records = append(recordSet.Records, internal.Record{Content: info.Value})
			added = true
		}

		newRecordSet = append(newRecordSet, recordSet)
	}

	if !added {
		newRecordSet = append(newRecordSet, internal.RecordSet{
			Name: subDomain,
			Type: "TXT",
			TTL:  d.config.TTL,
			Records: []internal.Record{
				{Content: info.Value},
			},
		})
	}

	request := internal.ZoneRequest{
		Overwrite: false,
		Zone:      newRecordSet,
	}

	err = d.client.UpdateDNSRecords(ctx, dns01.UnFqdn(authZone), request)
	if err != nil {
		return fmt.Errorf("hostinger: update DNS records (add): %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("hostinger: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("hostinger: %w", err)
	}

	ctx := context.Background()

	recordSets, err := d.client.GetDNSRecords(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("hostinger: get DNS records: %w", err)
	}

	var changed bool

	var newRecordSet []internal.RecordSet

	for _, recordSet := range recordSets {
		if recordSet.Name == subDomain && recordSet.Type == "TXT" {
			var rs []internal.Record

			for _, record := range recordSet.Records {
				if record.Content == info.Value {
					changed = true
				} else {
					rs = append(rs, record)
				}
			}

			recordSet.Records = rs
		}

		newRecordSet = append(newRecordSet, recordSet)
	}

	if !changed {
		return nil
	}

	request := internal.ZoneRequest{
		Overwrite: false,
		Zone:      newRecordSet,
	}

	err = d.client.UpdateDNSRecords(ctx, dns01.UnFqdn(authZone), request)
	if err != nil {
		return fmt.Errorf("hostinger: update DNS records (delete): %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
