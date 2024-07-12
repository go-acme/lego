// Package glesys implements a DNS provider for solving the DNS-01 challenge using GleSYS api.
package glesys

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/glesys/internal"
)

const minTTL = 60

// Environment variables names.
const (
	envNamespace = "GLESYS_"

	EnvAPIUser = envNamespace + "API_USER"
	EnvAPIKey  = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIUser            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	activeRecords map[string]int
	inProgressMu  sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for GleSYS.
// Credentials must be passed in the environment variables:
// GLESYS_API_USER and GLESYS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUser, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("glesys: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values[EnvAPIUser]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for GleSYS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("glesys: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, errors.New("glesys: incomplete credentials provided")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("glesys: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.APIUser, config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:        config,
		client:        client,
		activeRecords: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// find authZone
	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("glesys: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("glesys: %w", err)
	}

	// acquire lock and check there is not a challenge already in progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	// add TXT record into authZone
	recordID, err := d.client.AddTXTRecord(context.Background(), dns01.UnFqdn(authZone), subDomain, info.Value, d.config.TTL)
	if err != nil {
		return err
	}

	// save data necessary for CleanUp
	d.activeRecords[info.EffectiveFQDN] = recordID
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// acquire lock and retrieve authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.activeRecords[info.EffectiveFQDN]; !ok {
		// if there is no cleanup information then just return
		return nil
	}

	recordID := d.activeRecords[info.EffectiveFQDN]
	delete(d.activeRecords, info.EffectiveFQDN)

	// delete TXT record from authZone
	return d.client.DeleteTXTRecord(context.Background(), recordID)
}

// Timeout returns the values (20*time.Minute, 20*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with GleSYS.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
