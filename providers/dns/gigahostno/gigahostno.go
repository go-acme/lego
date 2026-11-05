// Package gigahostno implements a DNS provider for solving the DNS-01 challenge using Gigahost.no.
package gigahostno

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/gigahostno/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"golang.org/x/oauth2"
)

// Environment variables names.
const (
	envNamespace = "GIGAHOSTNO_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvSecret   = envNamespace + "SECRET"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string
	Secret   string

	APIkey string

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

// NewDNSProvider returns a DNSProvider instance configured for Gigahost.
func NewDNSProvider() (*DNSProvider, error) {
	apiKey := env.GetOrFile(EnvAPIKey)

	config := NewDefaultConfig()

	if apiKey != "" {
		config.APIkey = apiKey
	} else {
		values, err := env.Get(EnvUsername, EnvPassword)
		if err != nil {
			return nil, fmt.Errorf("gigahostno: %w", err)
		}

		config.Username = values[EnvUsername]
		config.Password = values[EnvPassword]
		config.Secret = env.GetOrFile(EnvSecret)
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gigahost.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gigahostno: the configuration of the DNS provider is nil")
	}

	authConfig, err := internal.NewAuthConfig(config.APIkey, config.Username, config.Password, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("gigahostno: %w", err)
	}

	return newDNSProviderConfig(config, authConfig)
}

func newDNSProviderConfig(config *Config, authConfig *internal.Config) (*DNSProvider, error) {
	client := internal.NewClient()

	ctx := context.Background()

	if config.HTTPClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, config.HTTPClient)
	}

	client.HTTPClient = clientdebug.Wrap(oauth2.NewClient(ctx, authConfig.TokenSource(ctx)))

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gigahostno: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Name)
	if err != nil {
		return fmt.Errorf("gigahostno: %w", err)
	}

	record := internal.Record{
		Name:  subDomain,
		Type:  "TXT",
		Value: info.Value,
		TTL:   d.config.TTL,
	}

	err = d.client.CreateNewRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("gigahostno: create new record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gigahostno: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Name)
	if err != nil {
		return fmt.Errorf("gigahostno: %w", err)
	}

	records, err := d.client.GetZoneRecords(ctx, zone.ID)
	if err != nil {
		return fmt.Errorf("gigahostno: get zone records: %w", err)
	}

	for _, record := range records {
		if record.Type == "TXT" && record.Name == subDomain && record.Value == info.Value {
			err := d.client.DeleteRecord(ctx, zone.ID, record.ID, record.Name, record.Type)
			if err != nil {
				return fmt.Errorf("gigahostno: delete record: %w", err)
			}

			break
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (*internal.Zone, error) {
	zones, err := d.client.GetZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("get zones: %w", err)
	}

	for d := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, zone := range zones {
			if zone.Name == d && zone.Active == "1" {
				return &zone, nil
			}
		}
	}

	return nil, fmt.Errorf("zone not found for %q", fqdn)
}
