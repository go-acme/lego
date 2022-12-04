package wedos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/wedos/internal"
)

// Environment variables names.
const (
	envNamespace = "WEDOS_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "WAPI_PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const minTTL = 5 * 60 // 5 minutes

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("wedos: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

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

	if config.TTL < minTTL {
		return nil, fmt.Errorf("wedos: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(config.Username, config.Password)

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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("wedos: could not determine zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, authZone)
	if err != nil {
		return fmt.Errorf("wedos: %w", err)
	}

	record := internal.DNSRow{
		Name: subDomain,
		TTL:  json.Number(strconv.Itoa(d.config.TTL)),
		Type: "TXT",
		Data: value,
	}

	records, err := d.client.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("wedos: could not get records for domain %q: %w", domain, err)
	}

	for _, candidate := range records {
		if candidate.Type == "TXT" && candidate.Name == subDomain && candidate.Data == value {
			record.ID = candidate.ID
			break
		}
	}

	err = d.client.AddRecord(ctx, authZone, record)
	if err != nil {
		return fmt.Errorf("wedos: could not add TXT record for domain %q: %w", domain, err)
	}

	err = d.client.Commit(ctx, authZone)
	if err != nil {
		return fmt.Errorf("wedos: could not commit TXT record for domain %q: %w", domain, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("wedos: could not determine zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, authZone)
	if err != nil {
		return fmt.Errorf("wedos: %w", err)
	}

	records, err := d.client.GetRecords(ctx, authZone)
	if err != nil {
		return fmt.Errorf("wedos: could not get records for domain %q: %w", domain, err)
	}

	for _, candidate := range records {
		if candidate.Type != "TXT" || candidate.Name != subDomain || candidate.Data != value {
			continue
		}

		err = d.client.DeleteRecord(ctx, authZone, candidate.ID)
		if err != nil {
			return fmt.Errorf("wedos: could not remove TXT record for domain %q: %w", domain, err)
		}

		err = d.client.Commit(ctx, authZone)
		if err != nil {
			return fmt.Errorf("wedos: could not commit TXT record for domain %q: %w", domain, err)
		}

		return nil
	}

	return nil
}
