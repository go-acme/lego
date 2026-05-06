// Package ucloud implements a DNS provider for solving the DNS-01 challenge using UCloud.
package ucloud

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/useragent"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/ucloud/internal"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// Environment variables names.
const (
	envNamespace = "UCLOUD_"

	EnvPublicKey  = envNamespace + "PUBLIC_KEY"
	EnvPrivateKey = envNamespace + "PRIVATE_KEY"

	EnvRegion    = envNamespace + "REGION"
	EnvProjectID = envNamespace + "PROJECT_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Region    string
	ProjectID string

	PublicKey  string
	PrivateKey string

	// only for test
	baseURL string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config

	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for UCloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPublicKey, EnvPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("ucloud: %w", err)
	}

	config := NewDefaultConfig()
	config.PublicKey = values[EnvPublicKey]
	config.PrivateKey = values[EnvPrivateKey]

	config.Region = env.GetOrFile(EnvRegion)
	config.ProjectID = env.GetOrFile(EnvProjectID)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for UCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ucloud: the configuration of the DNS provider is nil")
	}

	if config.PublicKey == "" || config.PrivateKey == "" {
		return nil, errors.New("ucloud: credentials missing")
	}

	credential := auth.NewCredential()
	credential.PublicKey = config.PublicKey
	credential.PrivateKey = config.PrivateKey

	cfg := ucloud.NewConfig()
	cfg.UserAgent = useragent.Get()

	if config.baseURL != "" {
		cfg.BaseUrl = config.baseURL
	}

	if config.Region != "" {
		cfg.Region = config.Region
	}

	if config.ProjectID != "" {
		cfg.ProjectId = config.ProjectID
	}

	return &DNSProvider{
		config: config,
		client: internal.NewClient(&cfg, &credential),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ucloud: could not find zone for domain %q: %w", domain, err)
	}

	addRequest := d.client.NewDomainDNSAddRequest()
	addRequest.Dn = ucloud.String(dns01.UnFqdn(authZone))
	addRequest.RecordName = ucloud.String(dns01.UnFqdn(info.EffectiveFQDN))
	addRequest.DnsType = ucloud.String("TXT")
	addRequest.Content = ucloud.String(info.Value)
	addRequest.TTL = ucloud.String(strconv.Itoa(d.config.TTL))
	addRequest.WithTimeout(d.config.HTTPTimeout)

	_, err = d.client.DomainDNSAdd(addRequest)
	if err != nil {
		return fmt.Errorf("ucloud: domain DNS add: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ucloud: could not find zone for domain %q: %w", domain, err)
	}

	queryRequest := d.client.NewDomainDNSQueryRequest()
	queryRequest.Dn = ucloud.String(dns01.UnFqdn(authZone))
	queryRequest.WithTimeout(d.config.HTTPTimeout)

	dom, err := d.client.DomainDNSQuery(queryRequest)
	if err != nil {
		return fmt.Errorf("ucloud: domain DNS query: %w", err)
	}

	for _, record := range dom.Data {
		if record.Type != "TXT" || record.Name != dns01.UnFqdn(info.EffectiveFQDN) || record.Content != info.Value {
			continue
		}

		deleteRequest := d.client.NewDeleteDNSRecordRequest()
		deleteRequest.Dn = ucloud.String(dns01.UnFqdn(authZone))
		deleteRequest.RecordName = ucloud.String(dns01.UnFqdn(info.EffectiveFQDN))
		deleteRequest.DnsType = ucloud.String(record.Type)
		deleteRequest.Content = ucloud.String(record.Content)
		deleteRequest.WithTimeout(d.config.HTTPTimeout)

		_, err = d.client.DeleteDNSRecord(deleteRequest)
		if err != nil {
			return fmt.Errorf("ucloud: delete DNS record: %w", err)
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
