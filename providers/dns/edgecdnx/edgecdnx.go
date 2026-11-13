// Package edgecdnx implements a DNS provider for solving the DNS-01 challenge using EdgeCDN-X.
package edgecdnx

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/edgecdnx/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "EDGECDNX_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvProjectID = envNamespace + "PROJECT_ID"
	EnvAPIURL    = envNamespace + "API_URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey    string
	ProjectID string
	APIURL    string

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

	zoneIDs   map[string]string
	zoneIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for EdgeCDN-X.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvProjectID)
	if err != nil {
		return nil, fmt.Errorf("edgecdnx: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.ProjectID = values[EnvProjectID]
	config.APIURL = env.GetOrFile(EnvAPIURL)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EdgeCDN-X.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("edgecdnx: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIURL, config.APIKey, config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("edgecdnx: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:  config,
		client:  client,
		zoneIDs: make(map[string]string),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zoneID, err := d.findZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("edgecdnx: %w", err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[token] = zoneID
	d.zoneIDsMu.Unlock()

	endpoints, err := d.client.ListDNSEndpoints(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("edgecdnx: list DNS endpoints: %w", err)
	}

	endpoint, err := findDNSEndpoint(endpoints, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		// Create
		record := internal.RecordSpec{
			DNSName:       dns01.UnFqdn(info.EffectiveFQDN),
			Type:          "TXT",
			TTL:           d.config.TTL,
			RoutingPolicy: internal.RoutingPolicySimple,
			Targets:       []string{info.Value},
		}

		err = d.client.CreateDNSEndpoint(ctx, zoneID, record)
		if err != nil {
			return fmt.Errorf("edgecdnx: create DNS endpoint: %w", err)
		}

		return nil
	}

	// Update

	targets := slices.Clone(endpoint.Spec.Targets)
	targets = append(targets, info.Value)

	record := internal.RecordSpec{
		Targets: targets,
	}

	err = d.client.UpdateDNSEndpoint(ctx, zoneID, endpoint.Metadata.Name, record)
	if err != nil {
		return fmt.Errorf("edgecdnx: update DNS endpoint: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.zoneIDsMu.Lock()
	zoneID, ok := d.zoneIDs[token]
	d.zoneIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("edgecdnx: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	d.zoneIDsMu.Lock()
	delete(d.zoneIDs, token)
	d.zoneIDsMu.Unlock()

	endpoints, err := d.client.ListDNSEndpoints(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("edgecdnx: list DNS endpoints: %w", err)
	}

	endpoint, err := findDNSEndpoint(endpoints, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("edgecdnx: %w", err)
	}

	targets := slices.Clone(endpoint.Spec.Targets)
	targets = slices.DeleteFunc(targets, func(target string) bool { return target == info.Value })

	// Delete
	if len(targets) == 0 {
		err = d.client.DeleteDNSEndpoint(ctx, zoneID, endpoint.Metadata.Name, "")
		if err != nil {
			return fmt.Errorf("edgecdnx: delete DNS endpoint: %w", err)
		}

		return nil
	}

	// Update

	record := internal.RecordSpec{
		Targets: targets,
	}

	err = d.client.UpdateDNSEndpoint(ctx, zoneID, endpoint.Metadata.Name, record)
	if err != nil {
		return fmt.Errorf("edgecdnx: update DNS endpoint: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZoneID(ctx context.Context, fqdn string) (string, error) {
	zones, err := d.client.ListZones(ctx)
	if err != nil {
		return "", fmt.Errorf("list zones: %w", err)
	}

	return findZoneID(zones, fqdn)
}

func findZoneID(zones []internal.Zone, fqdn string) (string, error) {
	slices.SortFunc(zones, func(a, b internal.Zone) int {
		bLabels := dns.SplitDomainName(b.Spec.Zone)
		slices.Reverse(bLabels)

		aLabels := dns.SplitDomainName(a.Spec.Zone)
		slices.Reverse(aLabels)

		return cmp.Compare(strings.Join(bLabels, "."), strings.Join(aLabels, "."))
	})

	for s := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, zone := range zones {
			if zone.Spec.Zone == "" || zone.Metadata.Name == "" {
				continue
			}

			if dns01.UnFqdn(zone.Spec.Zone) == s {
				return zone.Metadata.Name, nil
			}
		}
	}

	return "", errors.New("zone not found")
}

func findDNSEndpoint(records []internal.Record, fqdn string) (internal.Record, error) {
	for _, record := range records {
		if record.Spec.Type != "TXT" || !strings.EqualFold(record.Spec.DNSName, fqdn) {
			continue
		}

		if record.Spec.RoutingPolicy != internal.RoutingPolicySimple {
			continue
		}

		return record, nil
	}

	return internal.Record{}, errors.New("DNS endpoint not found")
}
