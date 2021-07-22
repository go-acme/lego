package gcore

import (
	"context"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/gcore/internal"
)

const (
	// ProviderCode a value for cli dns flag.
	ProviderCode = "gcore"

	envNamespace          = "GCORE_"
	envAPIUrl             = envNamespace + "_API_URL"
	envPermanentToken     = envNamespace + "PERMANENT_API_TOKEN"
	envTTL                = envNamespace + "TTL"
	envPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	envPollingInterval    = envNamespace + "POLLING_INTERVAL"
	envHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"

	defaultPropagationTimeout = 360 * time.Second
	defaultPollingInterval    = 20 * time.Second
)

type (
	// TXTRecordManager contract for API client.
	TXTRecordManager interface {
		AddTXTRecord(ctx context.Context, fqdn, value string, ttl int) error
		RemoveTXTRecord(ctx context.Context, fqdn, value string) error
	}
	// Config for DNSProvider.
	Config struct {
		PropagationTimeout time.Duration
		PollingInterval    time.Duration
		TTL                int
		HTTPTimeout        time.Duration
	}
	// DNSProviderOpt for constructor of DNSProvider.
	DNSProviderOpt func(*DNSProvider)
	// DNSProvider an implementation of challenge.Provider contract.
	DNSProvider struct {
		Config
		Client TXTRecordManager
	}
)

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() Config {
	return Config{
		TTL:                env.GetOrDefaultInt(envTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(envPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(envPollingInterval, defaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(envHTTPTimeout, defaultPropagationTimeout),
	}
}

// NewDNSProvider returns an instance of DNSProvider configured for G-Core Labs DNS API.
func NewDNSProvider(opts ...DNSProviderOpt) (*DNSProvider, error) {
	values, err := env.Get(envPermanentToken)
	if err != nil {
		return nil, fmt.Errorf("%s: required env vars: %w", ProviderCode, err)
	}
	cfg := NewDefaultConfig()
	p := &DNSProvider{
		Config: cfg,
		Client: internal.NewClient(
			values[envPermanentToken],
			func(client *internal.Client) {
				client.HTTPClient.Timeout = cfg.HTTPTimeout
			},
			func(client *internal.Client) {
				url := env.GetOrDefaultString(envAPIUrl, "")
				if url != "" {
					client.BaseURL = url
				}
			},
		),
	}
	for _, op := range opts {
		op(p)
	}
	return p, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	ctx, cancel := context.WithTimeout(context.Background(), d.Config.PropagationTimeout)
	defer cancel()
	err := d.Client.AddTXTRecord(ctx, fqdn, value, d.Config.TTL)
	if err != nil {
		return fmt.Errorf("add txt record: %w", err)
	}
	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	ctx, cancel := context.WithTimeout(context.Background(), d.Config.PropagationTimeout)
	defer cancel()
	err := d.Client.RemoveTXTRecord(ctx, fqdn, value)
	if err != nil {
		return fmt.Errorf("remove txt record: %w", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.Config.PropagationTimeout, d.Config.PollingInterval
}
