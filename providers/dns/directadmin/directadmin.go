package directadmin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/directadmin/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "DIRECTADMIN_"

	EnvAPIURL   = envNamespace + "API_URL"
	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvZoneName = envNamespace + "ZONE_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL  string
	Username string
	Password string

	ZoneName string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		ZoneName:           env.GetOrFile(EnvZoneName),
		TTL:                env.GetOrDefaultInt(EnvTTL, 30),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 60*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for DirectAdmin.
// Credentials must be passed in the environment variables:
// DIRECTADMIN_API_URL, DIRECTADMIN_USERNAME, DIRECTADMIN_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIURL, EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("directadmin: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvAPIURL]
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DirectAdmin.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.BaseURL == "" {
		return nil, errors.New("directadmin: missing API URL")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("directadmin: some credentials information are missing")
	}

	client, err := internal.NewClient(config.BaseURL, config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("directadmin: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := d.getZoneName(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("directadmin: [domain: %q] %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("directadmin: %w", err)
	}

	record := internal.Record{
		Name:  subDomain,
		Type:  "TXT",
		Value: info.Value,
		TTL:   d.config.TTL,
	}

	err = d.client.SetRecord(context.Background(), dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("directadmin: set record for zone %s and subdomain %s: %w", authZone, subDomain, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := d.getZoneName(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("directadmin: [domain: %q] %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("directadmin: %w", err)
	}

	record := internal.Record{
		Name:  subDomain,
		Type:  "TXT",
		Value: info.Value,
	}

	err = d.client.DeleteRecord(context.Background(), dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("directadmin: delete record for zone %s and subdomain %s: %w", authZone, subDomain, err)
	}

	return nil
}

func (d *DNSProvider) getZoneName(fqdn string) (string, error) {
	if d.config.ZoneName != "" {
		return d.config.ZoneName, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone for %s: %w", fqdn, err)
	}

	if authZone == "" {
		return "", errors.New("empty zone name")
	}

	return authZone, nil
}
