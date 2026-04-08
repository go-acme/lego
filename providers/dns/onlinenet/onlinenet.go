// Package onlinenet implements a DNS provider for solving the DNS-01 challenge using Online.net.
package onlinenet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/onlinenet/internal"
)

// Environment variables names.
const (
	envNamespace = "ONLINENET_"

	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	previousVersionUUIDs map[string]string
	versionUUIDs         map[string]string
	versionUUIDMu        sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Online.net.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("onlinenet: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Online.net.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("onlinenet: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("onlinenet: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:       config,
		client:       client,
		versionUUIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("onlinenet: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("onlinenet: %w", err)
	}

	currentVersion, err := d.client.GetZoneVersion(ctx, dns01.UnFqdn(authZone), "active")
	if err != nil {
		return fmt.Errorf("onlinenet: get zone version: %w", err)
	}

	currentZone, err := d.client.GetActiveZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("onlinenet: get active zone: %w", err)
	}

	zoneVersion, err := d.client.CreateZoneVersion(ctx, dns01.UnFqdn(authZone), "lego")
	if err != nil {
		return fmt.Errorf("onlinenet: create zone version: %w", err)
	}

	record := internal.RecordRequest{
		Name: subDomain,
		Type: "TXT",
		TTL:  d.config.TTL,
		Data: info.Value,
	}

	_, err = d.client.CreateResourceRecord(ctx, dns01.UnFqdn(authZone), zoneVersion.UUIDRef, record)
	if err != nil {
		return fmt.Errorf("onlinenet: create resource record: %w", err)
	}

	for _, resourceRecord := range currentZone {
		request := internal.RecordRequest{
			Name: resourceRecord.Name,
			Type: resourceRecord.Type,
			TTL:  resourceRecord.TTL,
			Data: resourceRecord.Data,
		}

		_, err = d.client.CreateResourceRecord(ctx, dns01.UnFqdn(authZone), zoneVersion.UUIDRef, request)
		if err != nil {
			return fmt.Errorf("onlinenet: create resource record: %w", err)
		}
	}

	err = d.client.EnableZoneVersion(ctx, dns01.UnFqdn(authZone), zoneVersion.UUIDRef)
	if err != nil {
		return fmt.Errorf("onlinenet: enable zone version: %w", err)
	}

	d.versionUUIDMu.Lock()
	d.previousVersionUUIDs[token] = currentVersion.UUIDRef
	d.versionUUIDs[token] = zoneVersion.UUIDRef
	d.versionUUIDMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("onlinenet: could not find zone for domain %q: %w", domain, err)
	}

	d.versionUUIDMu.Lock()
	previousVersionUUID, prevOK := d.previousVersionUUIDs[token]
	d.versionUUIDMu.Unlock()

	if !prevOK {
		return fmt.Errorf("onlinenet: unknown previous zone version UUID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.EnableZoneVersion(ctx, dns01.UnFqdn(authZone), previousVersionUUID)
	if err != nil {
		return fmt.Errorf("onlinenet: enable previous zone version: %w", err)
	}

	d.versionUUIDMu.Lock()
	versionUUID, ok := d.versionUUIDs[token]
	d.versionUUIDMu.Unlock()

	if !ok {
		return fmt.Errorf("onlinenet: unknown zone version UUID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.DeleteZoneVersion(ctx, dns01.UnFqdn(authZone), versionUUID)
	if err != nil {
		return fmt.Errorf("onlinenet: delete zone version: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}
