// Package gigahostno implements a DNS provider for solving the DNS-01 challenge using Gigahost.no.
package gigahostno

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/gigahostno/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "GIGAHOSTNO_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvSecret   = envNamespace + "SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string
	Secret   string

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

	identifier *internal.Identifier
	client     *internal.Client

	tokenMu sync.Mutex
	token   *internal.Token
}

// NewDNSProvider returns a DNSProvider instance configured for Gigahost.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("gigahostno: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.Secret = env.GetOrFile(EnvSecret)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gigahost.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gigahostno: the configuration of the DNS provider is nil")
	}

	identifier, err := internal.NewIdentifier(config.Username, config.Password, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("gigahostno: %w", err)
	}

	if config.HTTPClient != nil {
		identifier.HTTPClient = config.HTTPClient
	}

	identifier.HTTPClient = clientdebug.Wrap(identifier.HTTPClient)

	client := internal.NewClient()

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:     config,
		identifier: identifier,
		client:     client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.authenticate(ctx)
	if err != nil {
		return fmt.Errorf("gigahostno: %w", err)
	}

	ctx = internal.WithContext(ctx, d.token.Token)

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
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.authenticate(ctx)
	if err != nil {
		return fmt.Errorf("gigahostno: %w", err)
	}

	ctx = internal.WithContext(ctx, d.token.Token)

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

func (d *DNSProvider) authenticate(ctx context.Context) error {
	d.tokenMu.Lock()
	defer d.tokenMu.Unlock()

	if !d.token.IsExpired() {
		return nil
	}

	tok, err := d.identifier.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	d.token = tok

	return nil
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
