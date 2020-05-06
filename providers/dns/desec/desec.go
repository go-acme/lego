// Package desec implements a DNS provider for solving the DNS-01 challenge using deSEC DNS.
package desec

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/desec/internal"
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
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
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

	client := internal.NewClient(config.Token)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	quotedValue := fmt.Sprintf(`"%s"`, value)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("desec: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}

	recordName := getRecordName(fqdn, authZone)

	rrSet, err := d.client.GetTxtRRSet(dns01.UnFqdn(authZone), recordName)

	var nf *internal.NotFound
	if err != nil && !errors.As(err, &nf) {
		return fmt.Errorf("desec: failed to get records: %w", err)
	}

	if err != nil {
		var nf *internal.NotFound
		if !errors.As(err, &nf) {
			return fmt.Errorf("desec: failed to get records: %w", err)
		}

		// Not found case -> create
		_, err = d.client.AddTxtRRSet(internal.RRSet{
			Domain:  dns01.UnFqdn(authZone),
			SubName: recordName,
			Type:    "TXT",
			Records: []string{quotedValue},
			TTL:     d.config.TTL,
		})
		if err != nil {
			return fmt.Errorf("desec: failed to create records: %w", err)
		}

		return nil
	}

	// update
	records := append(rrSet.Records, quotedValue)

	_, err = d.client.UpdateTxtRRSet(dns01.UnFqdn(authZone), recordName, records)
	if err != nil {
		return fmt.Errorf("desec: failed to update records: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("desec: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}

	recordName := getRecordName(fqdn, authZone)

	rrSet, err := d.client.GetTxtRRSet(dns01.UnFqdn(authZone), recordName)
	if err != nil {
		return fmt.Errorf("desec: failed to create records: %w", err)
	}

	records := make([]string, 0)
	for _, record := range rrSet.Records {
		if record != fmt.Sprintf(`"%s"`, value) {
			records = append(records, record)
		}
	}

	_, err = d.client.UpdateTxtRRSet(dns01.UnFqdn(authZone), recordName, records)
	if err != nil {
		return fmt.Errorf("desec: failed to update records: %w", err)
	}

	return nil
}

func getRecordName(fqdn, authZone string) string {
	return fqdn[0 : len(fqdn)-len(authZone)-1]
}
