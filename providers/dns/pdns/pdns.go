// Package pdns implements a DNS provider for solving the DNS-01 challenge using PowerDNS nameserver.
package pdns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/pdns/internal"
)

// Environment variables names.
const (
	envNamespace = "PDNS_"

	EnvAPIKey = envNamespace + "API_KEY"
	EnvAPIURL = envNamespace + "API_URL"

	EnvTTL                = envNamespace + "TTL"
	EnvAPIVersion         = envNamespace + "API_VERSION"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvServerName         = envNamespace + "SERVER_NAME"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	Host               *url.URL
	ServerName         string
	APIVersion         int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		ServerName:         env.GetOrDefaultString(EnvServerName, "localhost"),
		APIVersion:         env.GetOrDefaultInt(EnvAPIVersion, 0),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for pdns.
// Credentials must be passed in the environment variable:
// PDNS_API_URL and PDNS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPIURL)
	if err != nil {
		return nil, fmt.Errorf("pdns: %w", err)
	}

	hostURL, err := url.Parse(values[EnvAPIURL])
	if err != nil {
		return nil, fmt.Errorf("pdns: %w", err)
	}

	config := NewDefaultConfig()
	config.Host = hostURL
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for pdns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("pdns: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("pdns: API key missing")
	}

	if config.Host == nil || config.Host.Host == "" {
		return nil, errors.New("pdns: API URL missing")
	}

	client := internal.NewClient(config.Host, config.ServerName, config.APIVersion, config.APIKey)

	if config.APIVersion <= 0 {
		err := client.SetAPIVersion(context.Background())
		if err != nil {
			log.Warnf("pdns: failed to get API version %v", err)
		}
	}

	return &DNSProvider{config: config, client: client}, nil
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
		return fmt.Errorf("pdns: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zone, err := d.client.GetHostedZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	name := info.EffectiveFQDN
	if d.client.APIVersion() == 0 {
		// pre-v1 API wants non-fqdn
		name = dns01.UnFqdn(info.EffectiveFQDN)
	}

	// Look for existing records.
	existingRRSet := findTxtRecord(zone, info.EffectiveFQDN)

	// merge the existing and new records
	var records []internal.Record
	if existingRRSet != nil {
		records = existingRRSet.Records
	}

	rec := internal.Record{
		Content:  "\"" + info.Value + "\"",
		Disabled: false,

		// pre-v1 API
		Type: "TXT",
		Name: name,
		TTL:  d.config.TTL,
	}

	rrSets := internal.RRSets{
		RRSets: []internal.RRSet{
			{
				Name:       name,
				ChangeType: "REPLACE",
				Type:       "TXT",
				Kind:       "Master",
				TTL:        d.config.TTL,
				Records:    append(records, rec),
			},
		},
	}

	err = d.client.UpdateRecords(ctx, zone, rrSets)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	return d.client.Notify(ctx, zone)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("pdns: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	zone, err := d.client.GetHostedZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	set := findTxtRecord(zone, info.EffectiveFQDN)

	if set == nil {
		return fmt.Errorf("pdns: no existing record found for %s", info.EffectiveFQDN)
	}

	rrSets := internal.RRSets{
		RRSets: []internal.RRSet{
			{
				Name:       set.Name,
				Type:       set.Type,
				ChangeType: "DELETE",
			},
		},
	}

	err = d.client.UpdateRecords(ctx, zone, rrSets)
	if err != nil {
		return fmt.Errorf("pdns: %w", err)
	}

	return d.client.Notify(ctx, zone)
}

func findTxtRecord(zone *internal.HostedZone, fqdn string) *internal.RRSet {
	for _, set := range zone.RRSets {
		if set.Type == "TXT" && (set.Name == dns01.UnFqdn(fqdn) || set.Name == fqdn) {
			return &set
		}
	}

	return nil
}
