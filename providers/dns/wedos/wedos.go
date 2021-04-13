package wedos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/wedos/libdns"
)

// Environment variables names.
const (
	envNamespace = "WEDOS_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "WAPI_PASSWORD"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	wedos  *Provider
	ctx    context.Context
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Username = env.GetOrFile(EnvUsername)
	config.Password = env.GetOrFile(EnvPassword)
	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("wedos: the configuration of the DNS provider is nil")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("wedos: some credentials information are missing")
	}

	api := Provider{
		Username:     config.Username,
		WapiPassword: config.Password,
		HTTPClient:   config.HTTPClient,
	}

	return &DNSProvider{config: config, wedos: &api, ctx: context.TODO()}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	return d.execute(domain, keyAuth, true)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.execute(domain, keyAuth, false)
}

func (d *DNSProvider) execute(domain, keyAuth string, doInsert bool) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("wedos: could not determine zone for domain %q: %w", domain, err)
	}

	subdomain := fqdn[0 : len(fqdn)-len(authZone)-1]

	rec := libdns.Record{
		ID:    "",
		Type:  "TXT",
		Name:  subdomain,
		Value: value,
		TTL:   5 * time.Minute, // wedos minimum
	}
	rec, err = d.wedos.FillRecordID(d.ctx, authZone, rec)
	if err != nil {
		return fmt.Errorf("wedos: could not list records for domain %q: %w", domain, err)
	}
	if doInsert {
		_, err = d.wedos.SetRecords(d.ctx, authZone, []libdns.Record{rec})
		if err != nil {
			return fmt.Errorf("wedos: could not set TXT record for domain %q: %w", domain, err)
		}
	} else if rec.ID != "" {
		_, err = d.wedos.DeleteRecords(d.ctx, authZone, []libdns.Record{rec})
		if err != nil {
			return fmt.Errorf("wedos: could not remove TXT record for domain %q: %w", domain, err)
		}
	}

	err = d.wedos.Commit(d.ctx, authZone)
	if err != nil {
		return fmt.Errorf("wedos: could not commit TXT record for domain %q: %w", domain, err)
	}

	return nil
}
