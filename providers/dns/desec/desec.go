// Package desec implements a DNS provider for solving the DNS-01 challenge using deSEC DNS.
package desec

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nrdcg/desec"
)

// Environment variables names.
const (
	envNamespace = "DESEC_"

	EnvToken = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// https://github.com/desec-io/desec-stack/issues/216
// https://desec.readthedocs.io/_/downloads/en/latest/pdf/
const defaultTTL int = 3600

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
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
	client *desec.Client
}

// NewDNSProvider returns a DNSProvider instance configured for deSEC.
// Credentials must be passed in the environment variable: DESEC_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("desec: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for deSEC.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("desec: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("desec: incomplete credentials, missing token")
	}

	opts := desec.NewDefaultClientOptions()
	if config.HTTPClient != nil {
		opts.HTTPClient = config.HTTPClient
	}
	opts.Logger = log.Default()

	client := desec.New(config.Token, opts)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	quotedValue := fmt.Sprintf(`%q`, value)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("desec: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}

	recordName, err := dns01.ExtractSubDomain(fqdn, authZone)
	if err != nil {
		return fmt.Errorf("desec: %w", err)
	}

	domainName := dns01.UnFqdn(authZone)

	rrSet, err := d.client.Records.Get(ctx, domainName, recordName, "TXT")
	if err != nil {
		var nf *desec.NotFoundError
		if !errors.As(err, &nf) {
			return fmt.Errorf("desec: failed to get records: domainName=%s, recordName=%s: %w", domainName, recordName, err)
		}

		// Not found case -> create
		_, err = d.client.Records.Create(ctx, desec.RRSet{
			Domain:  domainName,
			SubName: recordName,
			Type:    "TXT",
			Records: []string{quotedValue},
			TTL:     d.config.TTL,
		})
		if err != nil {
			return fmt.Errorf("desec: failed to create records: domainName=%s, recordName=%s: %w", domainName, recordName, err)
		}

		return nil
	}

	// update
	records := append(rrSet.Records, quotedValue)

	_, err = d.client.Records.Update(ctx, domainName, recordName, "TXT", desec.RRSet{Records: records})
	if err != nil {
		return fmt.Errorf("desec: failed to update records: domainName=%s, recordName=%s: %w", domainName, recordName, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("desec: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}

	recordName, err := dns01.ExtractSubDomain(fqdn, authZone)
	if err != nil {
		return fmt.Errorf("desec: %w", err)
	}

	domainName := dns01.UnFqdn(authZone)

	rrSet, err := d.client.Records.Get(ctx, domainName, recordName, "TXT")
	if err != nil {
		return fmt.Errorf("desec: failed to get records: domainName=%s, recordName=%s: %w", domainName, recordName, err)
	}

	records := make([]string, 0)
	for _, record := range rrSet.Records {
		if record != fmt.Sprintf(`%q`, value) {
			records = append(records, record)
		}
	}

	_, err = d.client.Records.Update(ctx, domainName, recordName, "TXT", desec.RRSet{Records: records})
	if err != nil {
		return fmt.Errorf("desec: failed to update records: domainName=%s, recordName=%s: %w", domainName, recordName, err)
	}

	return nil
}
