// Package webglobe implements a DNS provider for solving the DNS-01 challenge using Webglobe.
package webglobe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/webglobe/internal"
)

// Environment variables names.
const (
	envNamespace = "WEBGLOBE_"

	EnvLogin = envNamespace + "LOGIN"
	EnvToken = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Login string
	Token string

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

	zoneIDs     map[string]int64
	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Webglobe.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvLogin, EnvToken)
	if err != nil {
		return nil, fmt.Errorf("webglobe: %w", err)
	}

	config := NewDefaultConfig()
	config.Login = values[EnvLogin]
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Webglobe.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("webglobe: the configuration of the DNS provider is nil")
	}

	identifier, err := internal.NewIdentifier(config.Login, config.Token)
	if err != nil {
		return nil, fmt.Errorf("webglobe: %w", err)
	}

	if config.HTTPClient != nil {
		identifier.HTTPClient = config.HTTPClient
	}

	identifier.HTTPClient = clientdebug.Wrap(identifier.HTTPClient)

	client, err := internal.NewClient()
	if err != nil {
		return nil, fmt.Errorf("webglobe: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:     config,
		identifier: identifier,
		client:     client,
		recordIDs:  make(map[string]int64),
		zoneIDs:    make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	tok, err := d.identifier.Login(ctx)
	if err != nil {
		return fmt.Errorf("webglobe: login: %w", err)
	}

	ctxAuth := internal.WithContext(ctx, tok.Token)

	zone, err := d.findZone(ctxAuth, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("webglobe: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Domain)
	if err != nil {
		return fmt.Errorf("webglobe: %w", err)
	}

	record := internal.Record{
		Name: subDomain,
		Type: "TXT",
		Data: info.Value,
		TTL:  d.config.TTL,
	}

	newRecord, err := d.client.CreateRecord(ctxAuth, zone.DomainID, record)
	if err != nil {
		return fmt.Errorf("webglobe: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zone.DomainID
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, recordOK := d.recordIDs[token]
	zoneID, zoneOK := d.zoneIDs[token]
	d.recordIDsMu.Unlock()

	if !recordOK {
		return fmt.Errorf("webglobe: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !zoneOK {
		return fmt.Errorf("webglobe: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	tok, err := d.identifier.Login(ctx)
	if err != nil {
		return fmt.Errorf("webglobe: login: %w", err)
	}

	ctxAuth := internal.WithContext(ctx, tok.Token)

	err = d.client.DeleteRecord(ctxAuth, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("webglobe: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.zoneIDs, token)
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (*internal.Domain, error) {
	for dom := range dns01.UnFqdnDomainsSeq(fqdn) {
		details, err := d.client.GetDomainDetails(ctx, dom)
		if err != nil {
			log.Warn("get domain details", log.ErrorAttr(err), slog.String("domain", dom))

			continue
		}

		return details, nil
	}

	return nil, fmt.Errorf("could not find a zone for %s", fqdn)
}
