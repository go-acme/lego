package googledomains

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"google.golang.org/api/acmedns/v1"
	"google.golang.org/api/option"
)

// Environment variables names.
const (
	envNamespace = "GOOGLE_DOMAINS_"

	EnvAccessToken = envNamespace + "ACCESS_TOKEN"
	EnvZoneName    = envNamespace + "ZONE_NAME"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// static check on interface implementation
var _ challenge.Provider = &DNSProvider{}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessToken        string
	PollingInterval    time.Duration
	PropagationTimeout time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		AccessToken:        env.GetOrDefaultString(EnvAccessToken, ""),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
	}
}

// NewDNSProvider returns the Google Domains DNS provider with a default configuration.
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(NewDefaultConfig())
}

// NewDNSProviderConfig returns the Google Domains DNS provider with the provided config.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("google domains: the configuration of the DNS provider is nil")
	}

	if config.AccessToken == "" {
		return nil, fmt.Errorf("google domains: access token is missing")
	}

	httpClient := &http.Client{
		Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}

	service, err := acmedns.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("google domains: error creating acme dns service: %w", err)
	}

	return &DNSProvider{
		config: config,
		acmedns: service,
	}, nil
}

type DNSProvider struct {
	config     *Config
	acmedns *acmedns.Service
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	zone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return fmt.Errorf("error finding zone for domain %s: %w", domain, err)
	}

	rotateReq := acmedns.RotateChallengesRequest{
		AccessToken:        d.config.AccessToken,
		RecordsToAdd:       []*acmedns.AcmeTxtRecord{getAcmeTxtRecord(domain, token, keyAuth)},
		KeepExpiredRecords: false,
	}

	call := d.acmedns.AcmeChallengeSets.RotateChallenges(zone, &rotateReq)
	_, err = call.Do()
	if err != nil {
		return err
	}
	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	zone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return fmt.Errorf("error finding zone for domain %s: %w", domain, err)
	}

	rotateReq := acmedns.RotateChallengesRequest{
		AccessToken:        d.config.AccessToken,
		RecordsToRemove:    []*acmedns.AcmeTxtRecord{getAcmeTxtRecord(domain, token, keyAuth)},
		KeepExpiredRecords: false,
	}

	call := d.acmedns.AcmeChallengeSets.RotateChallenges(zone, &rotateReq)
	_, err = call.Do()
	if err != nil {
		return err
	}
	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func getAcmeTxtRecord(domain, token, keyAuth string) *acmedns.AcmeTxtRecord {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	return &acmedns.AcmeTxtRecord{
		Fqdn:   fqdn,
		Digest: value,
	}
}
