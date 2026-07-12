// Package openprovider implements a DNS provider for solving the DNS-01 challenge using Openprovider.
package openprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/openprovider/internal"
)

// Environment variables names.
const (
	envNamespace = "OPENPROVIDER_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

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

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 600*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	zoneIDs   map[string]int
	zoneIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Openprovider.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("openprovider: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Openprovider.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("openprovider: the configuration of the DNS provider is nil")
	}

	if config.TTL < 600 {
		return nil, errors.New("openprovider: TTL must be >= 600")
	}

	client, err := internal.NewClient(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("openprovider: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:  config,
		client:  client,
		zoneIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	ctxAuth, err := d.client.CreateAuthenticatedContext(ctx)
	if err != nil {
		return fmt.Errorf("openprovider: login: %w", err)
	}

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctxAuth, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("openprovider: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	zone, err := d.findZone(ctxAuth, authZone)
	if err != nil {
		return fmt.Errorf("openprovider: find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("openprovider: %w", err)
	}

	action := internal.ZoneAction{
		ID:   zone.ID,
		Name: authZone,
		Records: internal.RecordAction{
			Add: []internal.Record{{
				Name:  subDomain,
				TTL:   d.config.TTL,
				Type:  "TXT",
				Value: info.Value,
			}},
		},
	}

	err = d.client.UpdateZone(ctxAuth, authZone, action)
	if err != nil {
		return fmt.Errorf("openprovider: update zone: %w", err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[token] = zone.ID
	d.zoneIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	ctxAuth, err := d.client.CreateAuthenticatedContext(ctx)
	if err != nil {
		return fmt.Errorf("openprovider: login: %w", err)
	}

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctxAuth, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("openprovider: could not find zone for domain %q: %w", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	d.zoneIDsMu.Lock()
	zoneID, ok := d.zoneIDs[token]
	d.zoneIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("openprovider: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("openprovider: %w", err)
	}

	action := internal.ZoneAction{
		ID:   zoneID,
		Name: authZone,
		Records: internal.RecordAction{
			Remove: []internal.Record{{
				Name:  subDomain,
				TTL:   d.config.TTL,
				Type:  "TXT",
				Value: info.Value,
			}},
		},
	}

	err = d.client.UpdateZone(ctxAuth, authZone, action)
	if err != nil {
		return fmt.Errorf("openprovider: update zone %w", err)
	}

	d.zoneIDsMu.Lock()
	delete(d.zoneIDs, token)
	d.zoneIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, authZone string) (*internal.Zone, error) {
	zr := &internal.ZonesRequest{
		Limit:       500,
		Offset:      0,
		NamePattern: authZone,
	}

	for {
		zones, err := d.client.ListZones(ctx, zr)
		if err != nil {
			return nil, err
		}

		for _, zone := range zones {
			if zone.Name == authZone {
				return &zone, nil
			}
		}

		if len(zones) < zr.Limit {
			break
		}

		zr.Offset += zr.Limit
	}

	return nil, fmt.Errorf("zone %q not found", authZone)
}
