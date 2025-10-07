// Package variomedia implements a DNS provider for solving the DNS-01 challenge using Variomedia DNS.
package variomedia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/go-acme/lego/v4/providers/dns/variomedia/internal"
)

// Environment variables names.
const (
	envNamespace = "VARIOMEDIA_"

	EnvAPIToken = envNamespace + "API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
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
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
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

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("variomedia: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Variomedia.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.APIToken == "" {
		return nil, errors.New("variomedia: missing credentials")
	}

	client := internal.NewClient(config.APIToken)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]string),
	}, nil
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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("variomedia: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("variomedia: %w", err)
	}

	ctx := context.Background()

	record := internal.DNSRecord{
		RecordType: "TXT",
		Name:       subDomain,
		Domain:     dns01.UnFqdn(authZone),
		Data:       info.Value,
		TTL:        d.config.TTL,
	}

	cdrr, err := d.client.CreateDNSRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("variomedia: %w", err)
	}

	err = d.waitJob(ctx, domain, cdrr.Data.ID)
	if err != nil {
		return fmt.Errorf("variomedia: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = strings.TrimPrefix(cdrr.Data.Links.DNSRecord, "https://api.variomedia.de/dns-records/")
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("variomedia: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	ddrr, err := d.client.DeleteDNSRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("variomedia: %w", err)
	}

	err = d.waitJob(ctx, domain, ddrr.Data.ID)
	if err != nil {
		return fmt.Errorf("variomedia: %w", err)
	}

	return nil
}

func (d *DNSProvider) waitJob(ctx context.Context, domain, id string) error {
	return wait.Retry(ctx,
		func() error {
			result, err := d.client.GetJob(ctx, id)
			if err != nil {
				return fmt.Errorf("apply change on %s: %w", domain, err)
			}

			log.Infof("variomedia: [%s] %s: %s %s", domain, result.Data.ID, result.Data.Attributes.JobType, result.Data.Attributes.Status)

			if result.Data.Attributes.Status != "done" {
				return fmt.Errorf("apply change on %s: status: %s", domain, result.Data.Attributes.Status)
			}

			return nil
		},
		backoff.WithBackOff(backoff.NewConstantBackOff(d.config.PollingInterval)),
		backoff.WithMaxElapsedTime(d.config.PropagationTimeout),
	)
}
