// Package otc implements a DNS provider for solving the DNS-01 challenge using Open Telekom Cloud Managed DNS.
package otc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/otc/internal"
)

const defaultIdentityEndpoint = "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens"

// minTTL 300 is otc minimum value for TTL.
const minTTL = 300

// Environment variables names.
const (
	envNamespace = "OTC_"

	EnvDomainName       = envNamespace + "DOMAIN_NAME"
	EnvUserName         = envNamespace + "USER_NAME"
	EnvPassword         = envNamespace + "PASSWORD"
	EnvProjectName      = envNamespace + "PROJECT_NAME"
	EnvIdentityEndpoint = envNamespace + "IDENTITY_ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	IdentityEndpoint   string
	DomainName         string
	ProjectName        string
	UserName           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		IdentityEndpoint:   env.GetOrDefaultString(EnvIdentityEndpoint, defaultIdentityEndpoint),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,

				// Workaround for keep alive bug in otc api
				DisableKeepAlives: true,
			},
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for OTC DNS.
// Credentials must be passed in the environment variables: OTC_USER_NAME,
// OTC_DOMAIN_NAME, OTC_PASSWORD OTC_PROJECT_NAME and OTC_IDENTITY_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvDomainName, EnvUserName, EnvPassword, EnvProjectName)
	if err != nil {
		return nil, fmt.Errorf("otc: %w", err)
	}

	config := NewDefaultConfig()
	config.DomainName = values[EnvDomainName]
	config.UserName = values[EnvUserName]
	config.Password = values[EnvPassword]
	config.ProjectName = values[EnvProjectName]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OTC DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("otc: the configuration of the DNS provider is nil")
	}

	if config.DomainName == "" || config.UserName == "" || config.Password == "" || config.ProjectName == "" {
		return nil, errors.New("otc: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("otc: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.UserName, config.Password, config.DomainName, config.ProjectName)

	if config.IdentityEndpoint != "" {
		client.IdentityEndpoint = config.IdentityEndpoint
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("otc: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	ctx := context.Background()

	err = d.client.Login(ctx)
	if err != nil {
		return fmt.Errorf("otc: %w", err)
	}

	zoneID, err := d.client.GetZoneID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("otc: unable to get zone: %w", err)
	}

	record := internal.RecordSets{
		Name:        info.EffectiveFQDN,
		Description: "Added TXT record for ACME dns-01 challenge using lego client",
		Type:        "TXT",
		TTL:         d.config.TTL,
		Records:     []string{fmt.Sprintf("%q", info.Value)},
	}

	err = d.client.CreateRecordSet(ctx, zoneID, record)
	if err != nil {
		return fmt.Errorf("otc: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("otc: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	ctx := context.Background()

	err = d.client.Login(ctx)
	if err != nil {
		return fmt.Errorf("otc: %w", err)
	}

	zoneID, err := d.client.GetZoneID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("otc: %w", err)
	}

	recordID, err := d.client.GetRecordSetID(ctx, zoneID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("otc: unable to get record %s for zone %s: %w", info.EffectiveFQDN, domain, err)
	}

	err = d.client.DeleteRecordSet(ctx, zoneID, recordID)
	if err != nil {
		return fmt.Errorf("otc: %w", err)
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
