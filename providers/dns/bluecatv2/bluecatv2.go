// Package bluecatv2 implements a DNS provider for solving the DNS-01 challenge using Bluecat v2.
package bluecatv2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/bluecatv2/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "BLUECATV2_"

	EnvServerURL  = envNamespace + "SERVER_URL"
	EnvUsername   = envNamespace + "USERNAME"
	EnvPassword   = envNamespace + "PASSWORD"
	EnvConfigName = envNamespace + "CONFIG_NAME"
	EnvViewName   = envNamespace + "VIEW_NAME"
	EnvSkipDeploy = envNamespace + "SKIP_DEPLOY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ServerURL  string
	Username   string
	Password   string
	ConfigName string
	ViewName   string
	SkipDeploy bool

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		SkipDeploy: env.GetOrDefaultBool(EnvSkipDeploy, false),

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

	zoneIDs     map[string]int64
	recordIDs   map[string]int64
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Bluecat v2.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvServerURL, EnvUsername, EnvPassword, EnvConfigName, EnvViewName)
	if err != nil {
		return nil, fmt.Errorf("bluecatv2: %w", err)
	}

	config := NewDefaultConfig()
	config.ServerURL = values[EnvServerURL]
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.ConfigName = values[EnvConfigName]
	config.ViewName = values[EnvViewName]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Bluecat v2.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("bluecatv2: the configuration of the DNS provider is nil")
	}

	if config.ServerURL == "" {
		return nil, errors.New("bluecatv2: missing server URL")
	}

	if config.ConfigName == "" {
		return nil, errors.New("bluecatv2: missing configuration name")
	}

	if config.ViewName == "" {
		return nil, errors.New("bluecatv2: missing view name")
	}

	client, err := internal.NewClient(config.ServerURL, config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("bluecatv2: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int64),
		zoneIDs:   make(map[string]int64),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return fmt.Errorf("bluecatv2: %w", err)
	}

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("bluecatv2: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.AbsoluteName)
	if err != nil {
		return fmt.Errorf("bluecatv2: %w", err)
	}

	record := internal.RecordTXT{
		CommonResource: internal.CommonResource{
			Type: "TXTRecord",
			Name: subDomain,
		},
		Text:       info.Value,
		TTL:        d.config.TTL,
		RecordType: "TXT",
	}

	newRecord, err := d.client.CreateZoneResourceRecord(ctx, zone.ID, record)
	if err != nil {
		return fmt.Errorf("bluecatv2: create resource record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.zoneIDs[token] = zone.ID
	d.recordIDs[token] = newRecord.ID
	d.recordIDsMu.Unlock()

	if d.config.SkipDeploy {
		return nil
	}

	_, err = d.client.CreateZoneDeployment(ctx, zone.ID)
	if err != nil {
		return fmt.Errorf("bluecat: deploy zone: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, recordOK := d.recordIDs[token]
	zoneID, zoneOK := d.zoneIDs[token]
	d.recordIDsMu.Unlock()

	if !recordOK {
		return fmt.Errorf("bluecatv2: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	if !zoneOK {
		return fmt.Errorf("bluecatv2: unknown zone ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return fmt.Errorf("bluecatv2: %w", err)
	}

	err = d.client.DeleteResourceRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("bluecatv2: delete resource record: %w", err)
	}

	if d.config.SkipDeploy {
		return nil
	}

	_, err = d.client.CreateZoneDeployment(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("bluecat: deploy zone: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (*internal.ZoneResource, error) {
	for name := range dns01.UnFqdnDomainsSeq(fqdn) {
		opts := &internal.CollectionOptions{
			Fields: "id,absoluteName,configuration.id,configuration.name,view.id,view.name",
			Filter: internal.And(
				internal.Eq("absoluteName", name),
				internal.Eq("configuration.name", d.config.ConfigName),
				internal.Eq("view.name", d.config.ViewName),
			).String(),
		}

		zones, err := d.client.RetrieveZones(ctx, opts)
		if err != nil {
			// TODO(ldez) maybe add a log in v5.
			continue
		}

		for _, zone := range zones {
			if zone.AbsoluteName == name {
				return &zone, nil
			}
		}
	}

	return nil, fmt.Errorf("no zone found for fqdn: %s", fqdn)
}
