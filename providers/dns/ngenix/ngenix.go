// Package ngenix implements a DNS provider for solving the DNS-01 challenge using Ngenix.
package ngenix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/ngenix/internal"
)

// Environment variables names.
const (
	envNamespace = "NGENIX_"

	EnvUsername   = envNamespace + "USERNAME"
	EnvToken      = envNamespace + "TOKEN"
	EnvCustomerID = envNamespace + "CUSTOMER_ID"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username   string
	Token      string
	CustomerID string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 600*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 20*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Ngenix.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvToken, EnvCustomerID)
	if err != nil {
		return nil, fmt.Errorf("ngenix: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Token = values[EnvToken]
	config.CustomerID = values[EnvCustomerID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Ngenix.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ngenix: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Username, config.Token, config.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("ngenix: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zoneView, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ngenix: %w", err)
	}

	zone, err := d.client.GetDNSZone(ctx, zoneView.ID)
	if err != nil {
		return fmt.Errorf("ngenix: get DNS zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zoneView.Name)
	if err != nil {
		return fmt.Errorf("ngenix: %w", err)
	}

	records := append(zone.Records, internal.DNSRecord{
		Name: subDomain,
		Type: "TXT",
		Data: info.Value,
	})

	zoneUpdate := internal.DNSZoneUpdate{
		Records: records,
		Comment: zone.Comment,
		DNSSec:  zone.DNSSec,
	}

	_, err = d.client.UpdateDNSZone(ctx, zone.ID, zoneUpdate)
	if err != nil {
		return fmt.Errorf("ngenix: update DNS zone (add): %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zoneView, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ngenix: %w", err)
	}

	zone, err := d.client.GetDNSZone(ctx, zoneView.ID)
	if err != nil {
		return fmt.Errorf("ngenix: get DNS zone: %w", err)
	}

	var records []internal.DNSRecord

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zoneView.Name)
	if err != nil {
		return fmt.Errorf("ngenix: %w", err)
	}

	for _, record := range zone.Records {
		if record.Type == "TXT" && record.Name == subDomain && record.Data == info.Value {
			continue
		}

		records = append(records, record)
	}

	zoneUpdate := internal.DNSZoneUpdate{
		Records: records,
		Comment: zone.Comment,
		DNSSec:  zone.DNSSec,
	}

	_, err = d.client.UpdateDNSZone(ctx, zone.ID, zoneUpdate)
	if err != nil {
		return fmt.Errorf("ngenix: update DNS zone (remove): %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (*internal.DNSZoneCollectionView, error) {
	zones, err := d.client.ListDNSZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list DNS zones: %w", err)
	}

	for dom := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, zone := range zones {
			if zone.Name == dom {
				return &zone, nil
			}
		}
	}

	return nil, errors.New("could not find zone")
}
