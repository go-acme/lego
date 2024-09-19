// Package volcengine implements a DNS provider for solving the DNS-01 challenge using Volcano Engine.
package volcengine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/miekg/dns"
	"github.com/volcengine/volc-sdk-golang/base"
	volc "github.com/volcengine/volc-sdk-golang/service/dns"
)

// Environment variables names.
const (
	envNamespace = "VOLC_"

	EnvAccessKey = envNamespace + "ACCESSKEY"
	EnvSecretKey = envNamespace + "SECRETKEY"

	EnvRegion = envNamespace + "REGION"
	EnvHost   = envNamespace + "HOST"
	EnvScheme = envNamespace + "SCHEME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKey string
	SecretKey string

	Region string
	Host   string
	Scheme string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, volc.Timeout*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *volc.Client
	config *Config

	recordIDs   map[string]*string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Volcano Engine.
// Credentials must be passed in the environment variable: VOLC_ACCESSKEY, VOLC_SECRETKEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("volcengine: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.Scheme = env.GetOrDefaultString(EnvScheme, "https")
	config.Host = env.GetOrDefaultString(EnvHost, "open.volcengineapi.com")
	config.Region = env.GetOrDefaultString(EnvRegion, volc.DefaultRegion)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Volcano Engine.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("volcengine: the configuration of the DNS provider is nil")
	}

	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, errors.New("volcengine: missing credentials")
	}

	return &DNSProvider{
		config:    config,
		client:    newClient(config),
		recordIDs: make(map[string]*string),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneID, err := d.getZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("volcengine: get zone ID: %w", err)
	}

	crr := &volc.CreateRecordRequest{
		Host:  pointer(dns01.UnFqdn(info.EffectiveFQDN)),
		TTL:   pointer(int64(d.config.TTL)),
		Type:  pointer("TXT"),
		Value: pointer(info.Value),
		ZID:   zoneID,
	}

	record, err := d.client.CreateRecord(ctx, crr)
	if err != nil {
		return fmt.Errorf("volcengine: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = record.RecordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// gets the record's unique ID
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("volcengine: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	drr := &volc.DeleteRecordRequest{RecordID: recordID}

	err := d.client.DeleteRecord(context.Background(), drr)
	if err != nil {
		return fmt.Errorf("volcengine: delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) getZoneID(ctx context.Context, fqdn string) (*int64, error) {
	for _, index := range dns.Split(fqdn) {
		domain := fqdn[index:]

		lzr := &volc.ListZonesRequest{
			Key:        pointer(dns01.UnFqdn(domain)),
			SearchMode: pointer("exact"),
		}

		zones, err := d.client.ListZones(ctx, lzr)
		if err != nil {
			return nil, fmt.Errorf("list zones: %w", err)
		}

		total := deref(zones.Total)

		if total == 0 || len(zones.Zones) == 0 {
			continue
		}

		if total > 1 {
			return nil, fmt.Errorf("too many zone for %s", domain)
		}

		return zones.Zones[0].ZID, nil
	}

	return nil, fmt.Errorf("zone no found for fqdn: %s", fqdn)
}

// https://github.com/volcengine/volc-sdk-golang/tree/main/service/dns
// https://github.com/volcengine/volc-sdk-golang/blob/main/example/dns/demo_dns_test.go
func newClient(config *Config) *volc.Client {
	// https://github.com/volcengine/volc-sdk-golang/blob/fae992a31d02754e271c322095413d374ea4ea1b/service/dns/config.go#L20-L35
	serviceInfo := &base.ServiceInfo{
		Timeout: config.HTTPTimeout,
		Host:    config.Host,
		Header:  http.Header{"Accept": []string{"application/json"}},
		Scheme:  config.Scheme,
		Credentials: base.Credentials{
			Service:         volc.ServiceName,
			Region:          config.Region,
			AccessKeyID:     config.AccessKey,
			SecretAccessKey: config.SecretKey,
		},
	}

	// https://github.com/volcengine/volc-sdk-golang/blob/fae992a31d02754e271c322095413d374ea4ea1b/service/dns/caller.go#L17-L19
	client := base.NewClient(serviceInfo, nil)

	// https://github.com/volcengine/volc-sdk-golang/blob/fae992a31d02754e271c322095413d374ea4ea1b/service/dns/caller.go#L25-L34
	caller := &volc.VolcCaller{Volc: client}
	caller.Volc.SetAccessKey(serviceInfo.Credentials.AccessKeyID)
	caller.Volc.SetSecretKey(serviceInfo.Credentials.SecretAccessKey)
	caller.Volc.SetHost(serviceInfo.Host)
	caller.Volc.SetScheme(serviceInfo.Scheme)
	caller.Volc.SetTimeout(serviceInfo.Timeout)

	return volc.NewClient(caller)
}

func pointer[T any](v T) *T { return &v }

func deref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
