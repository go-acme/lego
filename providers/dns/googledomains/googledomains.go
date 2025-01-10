// Package googledomains implements a DNS provider for solving the DNS-01 challenge using Google Domains DNS API.
package googledomains

import (
	"context"
	"errors"
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

	EnvAccessToken        = envNamespace + "ACCESS_TOKEN"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessToken        string
	PollingInterval    time.Duration
	PropagationTimeout time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// NewDNSProvider returns the Google Domains DNS provider with a default configuration.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessToken)
	if err != nil {
		return nil, fmt.Errorf("googledomains: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessToken = values[EnvAccessToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig returns the Google Domains DNS provider with the provided config.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("googledomains: the configuration of the DNS provider is nil")
	}

	if config.AccessToken == "" {
		return nil, errors.New("googledomains: access token is missing")
	}

	service, err := acmedns.NewService(context.Background(), option.WithHTTPClient(config.HTTPClient))
	if err != nil {
		return nil, fmt.Errorf("googledomains: error creating acme dns service: %w", err)
	}

	return &DNSProvider{
		config:  config,
		acmedns: service,
	}, nil
}

type DNSProvider struct {
	config  *Config
	acmedns *acmedns.Service
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("googledomains: could not find zone for domain %q: %w", domain, err)
	}

	rotateReq := acmedns.RotateChallengesRequest{
		AccessToken:        d.config.AccessToken,
		RecordsToAdd:       []*acmedns.AcmeTxtRecord{getAcmeTxtRecord(domain, keyAuth)},
		KeepExpiredRecords: false,
	}

	call := d.acmedns.AcmeChallengeSets.RotateChallenges(zone, &rotateReq)
	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("googledomains: error adding challenge for domain %s: %w", domain, err)
	}
	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("googledomains: could not find zone for domain %q: %w", domain, err)
	}

	rotateReq := acmedns.RotateChallengesRequest{
		AccessToken:        d.config.AccessToken,
		RecordsToRemove:    []*acmedns.AcmeTxtRecord{getAcmeTxtRecord(domain, keyAuth)},
		KeepExpiredRecords: false,
	}

	call := d.acmedns.AcmeChallengeSets.RotateChallenges(zone, &rotateReq)
	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("googledomains: error cleaning up challenge for domain %s: %w", domain, err)
	}
	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func getAcmeTxtRecord(domain, keyAuth string) *acmedns.AcmeTxtRecord {
	challengeInfo := dns01.GetChallengeInfo(domain, keyAuth)

	return &acmedns.AcmeTxtRecord{
		Fqdn:   challengeInfo.EffectiveFQDN,
		Digest: challengeInfo.Value,
	}
}
