// Package volcengine implements a DNS provider for solving the DNS-01 challenge using Volcano Engine.
package volcengine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
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

// https://www.volcengine.com/docs/6758/170354
const defaultTTL = 600

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

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
		Scheme: env.GetOrDefaultString(EnvScheme, "https"),
		Host:   env.GetOrDefaultString(EnvHost, "open.volcengineapi.com"),
		Region: env.GetOrDefaultString(EnvRegion, volc.DefaultRegion),

		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 240*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
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

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("volcengine: get zone ID: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, deref(zone.ZoneName))
	if err != nil {
		return fmt.Errorf("volcengine: %w", err)
	}

	crr := &volc.CreateRecordRequest{
		Host:  pointer(subDomain),
		TTL:   pointer(int64(d.config.TTL)),
		Type:  pointer("TXT"),
		Value: pointer(info.Value),
		ZID:   zone.ZID,
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

func (d *DNSProvider) getZone(ctx context.Context, fqdn string) (volc.TopZoneResponse, error) {
	for _, index := range dns.Split(fqdn) {
		domain := fqdn[index:]

		lzr := &volc.ListZonesRequest{
			Key:        pointer(dns01.UnFqdn(domain)),
			SearchMode: pointer("exact"),
		}

		zones, err := d.client.ListZones(ctx, lzr)
		if err != nil {
			return volc.TopZoneResponse{}, fmt.Errorf("list zones: %w", err)
		}

		total := deref(zones.Total)

		if total == 0 || len(zones.Zones) == 0 {
			continue
		}

		if total > 1 {
			return volc.TopZoneResponse{}, fmt.Errorf("too many zone for %s", domain)
		}

		return zones.Zones[0], nil
	}

	return volc.TopZoneResponse{}, fmt.Errorf("zone no found for fqdn: %s", fqdn)
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
