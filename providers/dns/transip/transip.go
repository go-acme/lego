// Package transip implements a DNS provider for solving the DNS-01 challenge using TransIP.
package transip

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/transip/gotransip/v6"
	transipdomain "github.com/transip/gotransip/v6/domain"
)

// Environment variables names.
const (
	envNamespace = "TRANSIP_"

	EnvAccountName    = envNamespace + "ACCOUNT_NAME"
	EnvPrivateKeyPath = envNamespace + "PRIVATE_KEY_PATH"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccountName        string
	PrivateKeyPath     string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int64
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                int64(env.GetOrDefaultInt(EnvTTL, 10)),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config     *Config
	repository transipdomain.Repository
}

// NewDNSProvider returns a DNSProvider instance configured for TransIP.
// Credentials must be passed in the environment variables:
// TRANSIP_ACCOUNTNAME, TRANSIP_PRIVATEKEYPATH.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccountName, EnvPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("transip: %w", err)
	}

	config := NewDefaultConfig()
	config.AccountName = values[EnvAccountName]
	config.PrivateKeyPath = values[EnvPrivateKeyPath]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for TransIP.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("transip: the configuration of the DNS provider is nil")
	}

	client, err := gotransip.NewClient(gotransip.ClientConfiguration{
		AccountName:    config.AccountName,
		PrivateKeyPath: config.PrivateKeyPath,
	})
	if err != nil {
		return nil, fmt.Errorf("transip: %w", err)
	}

	repo := transipdomain.Repository{Client: client}

	return &DNSProvider{repository: repo, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("transip: could not find zone for domain %q: %w", domain, err)
	}

	// get the subDomain
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("transip: %w", err)
	}

	domainName := dns01.UnFqdn(authZone)

	entry := transipdomain.DNSEntry{
		Name:    subDomain,
		Expire:  int(d.config.TTL),
		Type:    "TXT",
		Content: info.Value,
	}

	err = d.repository.AddDNSEntry(domainName, entry)
	if err != nil {
		return fmt.Errorf("transip: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("transip: could not find zone for domain %q: %w", domain, err)
	}

	// get the subDomain
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("transip: %w", err)
	}

	domainName := dns01.UnFqdn(authZone)

	// get all DNS entries
	dnsEntries, err := d.repository.GetDNSEntries(domainName)
	if err != nil {
		return fmt.Errorf("transip: error for %s in CleanUp: %w", info.EffectiveFQDN, err)
	}

	// loop through the existing entries and remove the specific record
	for _, entry := range dnsEntries {
		if entry.Name == subDomain && entry.Content == info.Value {
			if err = d.repository.RemoveDNSEntry(domainName, entry); err != nil {
				return fmt.Errorf("transip: couldn't get Record ID in CleanUp: %w", err)
			}
			return nil
		}
	}

	return nil
}
