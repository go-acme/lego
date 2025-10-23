// Package hetznerhcloud implements a DNS provider for solving the DNS-01 challenge using Hetzner Cloud DNS API.
package hetznerhcloud

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
	"github.com/go-acme/lego/v4/providers/dns/hetznerhcloud/internal"
)

const (
	envNamespace = "HCLOUD_"

	EnvToken              = envNamespace + "TOKEN"
	EnvBaseURL            = envNamespace + "BASE_URL"
	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	Token              string
	BaseURL            string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(EnvBaseURL, internal.DefaultBaseURL),
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient:         &http.Client{Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second)},
	}
}

type DNSProvider struct {
	config    *Config
	client    *internal.Client
	recordIDs map[string]string
	mu        sync.Mutex
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("hetznerhcloud: %w", err)
	}
	config := NewDefaultConfig()
	config.Token = values[EnvToken]
	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil || config.Token == "" {
		return nil, errors.New("hetznerhcloud: missing credentials")
	}
	client := internal.NewClient(config.Token)
	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}
	if config.BaseURL != "" {
		if err := client.SetBaseURL(config.BaseURL); err != nil {
			return nil, fmt.Errorf("hetznerhcloud: invalid base URL: %w", err)
		}
	}
	return &DNSProvider{config: config, client: client, recordIDs: make(map[string]string)}, nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	zone, _ := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	zoneName := dns01.UnFqdn(zone)

	ctx := context.Background()
	zoneID, err := d.client.GetZone(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("hetznerhcloud: %w", err)
	}

	recordName := d.getRecordName(info.EffectiveFQDN, zoneName)
	recordID, err := d.client.CreateRecord(ctx, zoneID, recordName, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("hetznerhcloud: %w", err)
	}

	d.mu.Lock()
	d.recordIDs[strings.ToLower(info.EffectiveFQDN)] = recordID
	d.mu.Unlock()
	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.mu.Lock()
	recordID, ok := d.recordIDs[strings.ToLower(info.EffectiveFQDN)]
	delete(d.recordIDs, strings.ToLower(info.EffectiveFQDN))
	d.mu.Unlock()

	if !ok {
		return nil
	}

	zone, _ := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	zoneName := dns01.UnFqdn(zone)

	ctx := context.Background()
	zoneID, err := d.client.GetZone(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("hetznerhcloud: %w", err)
	}

	return d.client.DeleteRecord(ctx, zoneID, recordID)
}

func (d *DNSProvider) getRecordName(fqdn, zoneName string) string {
	fqdn = dns01.UnFqdn(fqdn)
	suffix := "." + zoneName
	if strings.EqualFold(fqdn, zoneName) || !strings.HasSuffix(strings.ToLower(fqdn), strings.ToLower(suffix)) {
		return "_acme-challenge"
	}
	return fqdn[:len(fqdn)-len(suffix)]
}
