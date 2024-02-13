// Package abion implements a DNS provider for solving the DNS-01 challenge using Abion.
package abion

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/abion/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "ABION_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
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
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Abion.
// Credentials must be passed in the environment variable: ABION_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("abion: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Abion.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("abion: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("abion: credentials missing")
	}

	client := internal.NewClient(config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("abion: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("abion: %w", err)
	}

	zones, err := d.client.GetZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("abion: get zone %w", err)
	}

	var data []internal.Record

	if sub, ok := zones.Data.Attributes.Records[subDomain]; ok {
		if records, exist := sub["TXT"]; exist {
			data = append(data, records...)
		}
	}

	data = append(data, internal.Record{
		TTL:      d.config.TTL,
		Data:     info.Value,
		Comments: "lego",
	})

	patch := internal.ZoneRequest{
		Data: internal.Zone{
			Type: "zone",
			ID:   dns01.UnFqdn(authZone),
			Attributes: internal.Attributes{
				Records: map[string]map[string][]internal.Record{
					subDomain: {"TXT": data},
				},
			},
		},
	}

	_, err = d.client.UpdateZone(ctx, dns01.UnFqdn(authZone), patch)
	if err != nil {
		return fmt.Errorf("abion: update zone %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("abion: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("abion: %w", err)
	}

	zones, err := d.client.GetZone(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("abion: get zone %w", err)
	}

	var data []internal.Record

	if sub, ok := zones.Data.Attributes.Records[subDomain]; ok {
		if records, exist := sub["TXT"]; exist {
			for _, record := range records {
				if record.Data != info.Value {
					data = append(data, record)
				}
			}
		}
	}

	payload := map[string][]internal.Record{}
	if len(data) == 0 {
		payload["TXT"] = nil
	} else {
		payload["TXT"] = data
	}

	patch := internal.ZoneRequest{
		Data: internal.Zone{
			Type: "zone",
			ID:   dns01.UnFqdn(authZone),
			Attributes: internal.Attributes{
				Records: map[string]map[string][]internal.Record{
					subDomain: payload,
				},
			},
		},
	}

	_, err = d.client.UpdateZone(ctx, dns01.UnFqdn(authZone), patch)
	if err != nil {
		return fmt.Errorf("abion: update zone %w", err)
	}

	return nil
}
