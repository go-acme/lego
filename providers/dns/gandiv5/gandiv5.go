// Package gandiv5 implements a DNS provider for solving the DNS-01 challenge using Gandi LiveDNS api.
package gandiv5

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/gandiv5/internal"
)

// Environment variables names.
const (
	envNamespace = "GANDIV5_"

	EnvAPIKey              = envNamespace + "API_KEY"
	EnvPersonalAccessToken = envNamespace + "PERSONAL_ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// inProgressInfo contains information about an in-progress challenge.
type inProgressInfo struct {
	fieldName string
	authZone  string
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL             string
	APIKey              string // Deprecated use PersonalAccessToken
	PersonalAccessToken string
	PropagationTimeout  time.Duration
	PollingInterval     time.Duration
	TTL                 int
	HTTPClient          *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	inProgressFQDNs map[string]inProgressInfo
	inProgressMu    sync.Mutex

	// findZoneByFqdn determines the DNS zone of a FQDN.
	// It is overridden during tests.
	// only for testing purpose.
	findZoneByFqdn func(fqdn string) (string, error)
}

// NewDNSProvider returns a DNSProvider instance configured for Gandi.
// Credentials must be passed in the environment variable: GANDIV5_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	// TODO(ldez): rewrite this when APIKey will be removed.
	config := NewDefaultConfig()
	config.APIKey = env.GetOrFile(EnvAPIKey)
	config.PersonalAccessToken = env.GetOrFile(EnvPersonalAccessToken)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gandi.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gandiv5: the configuration of the DNS provider is nil")
	}

	if config.APIKey != "" {
		log.Print("gandiv5: API Key is deprecated, use Personal Access Token instead")
	}

	if config.APIKey == "" && config.PersonalAccessToken == "" {
		return nil, errors.New("gandiv5: credentials information are missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("gandiv5: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.APIKey, config.PersonalAccessToken)

	if config.BaseURL != "" {
		baseURL, err := url.Parse(config.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("gandiv5: %w", err)
		}
		client.BaseURL = baseURL
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:          config,
		client:          client,
		inProgressFQDNs: make(map[string]inProgressInfo),
		findZoneByFqdn:  dns01.FindZoneByFqdn,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// find authZone
	authZone, err := d.findZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gandiv5: could not find zone for domain %q: %w", domain, err)
	}

	// determine name of TXT record
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("gandiv5: %w", err)
	}

	// acquire lock and check there is not a challenge already in
	// progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	// add TXT record into authZone
	err = d.client.AddTXTRecord(context.Background(), dns01.UnFqdn(authZone), subDomain, info.Value, d.config.TTL)
	if err != nil {
		return err
	}

	// save data necessary for CleanUp
	d.inProgressFQDNs[info.EffectiveFQDN] = inProgressInfo{
		authZone:  authZone,
		fieldName: subDomain,
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// acquire lock and retrieve authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.inProgressFQDNs[info.EffectiveFQDN]; !ok {
		// if there is no cleanup information then just return
		return nil
	}

	fieldName := d.inProgressFQDNs[info.EffectiveFQDN].fieldName
	authZone := d.inProgressFQDNs[info.EffectiveFQDN].authZone
	delete(d.inProgressFQDNs, info.EffectiveFQDN)

	// delete TXT record from authZone
	err := d.client.DeleteTXTRecord(context.Background(), dns01.UnFqdn(authZone), fieldName)
	if err != nil {
		return fmt.Errorf("gandiv5: %w", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
