// Package gravity implements a DNS provider for solving the DNS-01 challenge using Gravity.
package gravity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/gravity/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/google/uuid"
)

// Environment variables names.
const (
	envNamespace = "GRAVITY_"

	EnvUsername  = envNamespace + "USERNAME"
	EnvPassword  = envNamespace + "PASSWORD"
	EnvServerURL = envNamespace + "SERVER_URL"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username  string
	Password  string
	ServerURL string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, 1*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	records   map[string]internal.Record
	recordsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Gravity.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword, EnvServerURL)
	if err != nil {
		return nil, fmt.Errorf("gravity: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.ServerURL = values[EnvServerURL]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gravity.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gravity: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.ServerURL, config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("gravity: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:  config,
		client:  client,
		records: make(map[string]internal.Record),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	_, err := d.client.Login(ctx)
	if err != nil {
		return fmt.Errorf("gravity: login: %w", err)
	}

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gravity: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("gravity: %w", err)
	}

	id := uuid.New()

	record := internal.Record{
		Data:     info.Value,
		Hostname: subDomain,
		Type:     "TXT",
		UID:      id.String(),
	}

	err = d.client.CreateDNSRecord(ctx, zone, record)
	if err != nil {
		return fmt.Errorf("gravity: create DNS record: %w", err)
	}

	d.recordsMu.Lock()

	record.Fqdn = zone
	d.records[token] = record
	d.recordsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordsMu.Lock()
	record, ok := d.records[token]
	d.recordsMu.Unlock()

	if !ok {
		return fmt.Errorf("gravity: unknown record for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteDNSRecord(context.Background(), record.Fqdn, record)
	if err != nil {
		return fmt.Errorf("gravity: delete record: %w", err)
	}

	d.recordsMu.Lock()
	delete(d.records, token)
	d.recordsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential implements the [dns01.sequential] interface.
// It changes the behavior of the provider to resolve DNS challenges sequentially.
// Returns the interval between each iteration.
//
// Gravity supports adding multiple records for the same domain, but the DNS server doesn't work as expected:
// if you call the DNS server, it will answer only the latest record instead of all of them.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

func (d *DNSProvider) findZone(ctx context.Context, effectiveFQDN string) (string, error) {
	var zone string

	for fqdn := range dns01.DomainsSeq(effectiveFQDN) {
		zones, err := d.client.GetDNSZones(ctx, fqdn)
		if err != nil {
			return "", fmt.Errorf("get DNS zones: %w", err)
		}

		if len(zones) != 0 {
			zone = zones[0].Name
			break
		}
	}

	if zone == "" {
		return "", fmt.Errorf("could not find zone for %q", effectiveFQDN)
	}

	return zone, nil
}
