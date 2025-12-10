package shellrent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/shellrent/internal"
)

// Environment variables names.
const (
	envNamespace = "SHELLRENT_"

	EnvUsername = envNamespace + "USERNAME"
	EnvToken    = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultTTL = 3600

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type reqKey struct {
	domainID int
	recordID int
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username           string
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
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

	recordIDs   map[string]reqKey
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Shellrent.
// Credentials must be passed in the environment variable: SHELLRENT_USERNAME, SHELLRENT_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvToken)
	if err != nil {
		return nil, fmt.Errorf("shellrent: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Shellrent.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("shellrent: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("shellrent: missing credentials: username")
	}

	if config.Token == "" {
		return nil, errors.New("shellrent: missing credentials: token")
	}

	client := internal.NewClient(config.Username, config.Token)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]reqKey),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.findZone(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("shellrent: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.DomainName)
	if err != nil {
		return fmt.Errorf("shellrent: %w", err)
	}

	record := internal.Record{
		Type:        "TXT",
		Host:        subDomain,
		TTL:         internal.TTLRounder(d.config.TTL),
		Destination: info.Value,
	}

	recordID, err := d.client.CreateRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("shellrent: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = reqKey{domainID: zone.ID, recordID: recordID}
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// gets the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	key, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("shellrent: unknown request key for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, key.domainID, key.recordID)
	if err != nil {
		return fmt.Errorf("shellrent: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) findZone(ctx context.Context, domain string) (*internal.DomainDetails, error) {
	services, err := d.client.ListServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}

	for _, service := range services {
		details, err := d.client.GetServiceDetails(ctx, service)
		if err != nil {
			return nil, fmt.Errorf("get service details: %w", err)
		}

		domainDetails, err := d.client.GetDomainDetails(ctx, details.DomainID)
		if err != nil {
			return nil, fmt.Errorf("get domain details: %w", err)
		}

		domain := domain

		for {
			i := strings.Index(domain, ".")
			if i == -1 {
				break
			}

			if strings.EqualFold(domainDetails.DomainName, domain) {
				return domainDetails, nil
			}

			domain = domain[i+1:]
		}
	}

	return nil, errors.New("zone not found")
}
