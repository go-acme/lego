// Package easydns implements a DNS provider for solving the DNS-01 challenge using EasyDNS API.
package easydns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/easydns/internal"
)

// Environment variables names.
const (
	envNamespace = "EASYDNS_"

	EnvEndpoint = envNamespace + "ENDPOINT"
	EnvToken    = envNamespace + "TOKEN"
	EnvKey      = envNamespace + "KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Endpoint           *url.URL
	Token              string
	Key                string
	TTL                int
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
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

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	endpoint, err := url.Parse(env.GetOrDefaultString(EnvEndpoint, internal.DefaultBaseURL))
	if err != nil {
		return nil, fmt.Errorf("easydns: %w", err)
	}
	config.Endpoint = endpoint

	values, err := env.Get(EnvToken, EnvKey)
	if err != nil {
		return nil, fmt.Errorf("easydns: %w", err)
	}

	config.Token = values[EnvToken]
	config.Key = values[EnvKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EasyDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("easydns: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("easydns: the API token is missing")
	}

	if config.Key == "" {
		return nil, errors.New("easydns: the API key is missing")
	}

	client := internal.NewClient(config.Token, config.Key)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	if config.Endpoint != nil {
		client.BaseURL = config.Endpoint
	}

	return &DNSProvider{config: config, client: client, recordIDs: map[string]string{}}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := d.findZone(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("easydns: %w", err)
	}

	if authZone == "" {
		return fmt.Errorf("easydns: could not find zone for domain %q", domain)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("easydns: %w", err)
	}

	record := internal.ZoneRecord{
		Domain:   authZone,
		Host:     subDomain,
		Type:     "TXT",
		Rdata:    info.Value,
		TTL:      strconv.Itoa(d.config.TTL),
		Priority: "0",
	}

	recordID, err := d.client.AddRecord(ctx, dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("easydns: error adding zone record: %w", err)
	}

	key := getMapKey(info.EffectiveFQDN, info.Value)

	d.recordIDsMu.Lock()
	d.recordIDs[key] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	key := getMapKey(info.EffectiveFQDN, info.Value)

	d.recordIDsMu.Lock()
	recordID, exists := d.recordIDs[key]
	d.recordIDsMu.Unlock()

	if !exists {
		return nil
	}

	authZone, err := d.findZone(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("easydns: %w", err)
	}

	if authZone == "" {
		return fmt.Errorf("easydns: could not find zone for domain %q", domain)
	}

	err = d.client.DeleteRecord(ctx, dns01.UnFqdn(authZone), recordID)

	d.recordIDsMu.Lock()
	defer delete(d.recordIDs, key)
	d.recordIDsMu.Unlock()

	if err != nil {
		return fmt.Errorf("easydns: %w", err)
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

func getMapKey(fqdn, value string) string {
	return fqdn + "|" + value
}

func (d *DNSProvider) findZone(ctx context.Context, domain string) (string, error) {
	var errAll error

	for {
		i := strings.Index(domain, ".")
		if i == -1 {
			break
		}

		_, err := d.client.ListZones(ctx, domain)
		if err == nil {
			return domain, nil
		}

		errAll = errors.Join(errAll, err)

		domain = domain[i+1:]
	}

	return "", errAll
}
