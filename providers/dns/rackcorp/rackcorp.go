// Package rackcorp implements a DNS provider for solving the DNS-01 challenge using RackCorp.
package rackcorp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/rackcorp/internal"
)

// Environment variables names.
const (
	envNamespace = "RACKCORP_"

	EnvAPIUUID   = envNamespace + "API_UUID"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIUUID   string
	APISecret string

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

	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for RackCorp.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUUID, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("rackcorp: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUUID = values[EnvAPIUUID]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for RackCorp.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rackcorp: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIUUID, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("rackcorp: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	dom, err := d.findDomain(ctx, domain)
	if err != nil {
		return fmt.Errorf("rackcorp: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, dom.Name)
	if err != nil {
		return fmt.Errorf("rackcorp: %w", err)
	}

	record := internal.Record{
		Type:     "TXT",
		Lookup:   subDomain,
		DomainID: dom.ID,
		Data:     info.Value,
		TTL:      d.config.TTL,
	}

	records, err := d.client.CreateRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("rackcorp: create record: %w", err)
	}

	var recordID int64

	for _, r := range records {
		if r.Type == "TXT" && r.Data == info.Value {
			recordID = r.ID
			break
		}
	}

	if recordID == 0 {
		return fmt.Errorf("rackcorp: record not found (%s)", info.EffectiveFQDN)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("rackcorp: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("rackcorp: delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDomain(ctx context.Context, fqdn string) (*internal.Domain, error) {
	domains, err := d.client.GetDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("get domains: %w", err)
	}

	for dom := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, domain := range domains {
			if dom == domain.Name {
				return &domain, nil
			}
		}
	}

	return nil, fmt.Errorf("domain not found for %s", fqdn)
}
