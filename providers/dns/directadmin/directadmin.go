package directadmin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/directadmin/internal"
)

// Environment variables names.
const (
	envNamespace = "DIRECTADMIN_"

	EnvAPIURL   = envNamespace + "API_URL"
	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Username           string
	Password           string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(EnvAPIURL, internal.DefaultBaseURL),
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
}

// NewDNSProvider returns a DNSProvider instance configured for DirectAdmin.
// Credentials must be passed in the environment variables:
// DIRECTADMIN_API_URL, DIRECTADMIN_USERNAME, DIRECTADMIN_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("directadmin: %w", err)
	}

	config := NewDefaultConfig()
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

	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.client.SetRecord(context.Background(), info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("directadmin: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.client.DeleteRecord(context.Background(), info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("directadmin: %w", err)
	}

	return nil
}
