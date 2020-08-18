// Package joker implements a DNS provider for solving the DNS-01 challenge using joker.com DMAPI.
package joker

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "JOKER_"

	EnvAPIKey   = envNamespace + "API_KEY"
	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvDebug    = envNamespace + "DEBUG"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Debug              bool
	BaseURL            string
	APIKey             string
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
	AuthSid            string
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		Debug:              env.GetOrDefaultBool(conf, EnvDebug, false),
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 60*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Joker DMAPI.
// Credentials must be passed in the environment variable JOKER_API_KEY.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIKey)
	if err != nil {
		var errU error
		values, errU = env.Get(conf, EnvUsername, EnvPassword)
		if errU != nil {
			return nil, fmt.Errorf("joker: %v or %v", errU, err)
		}
	}

	config := NewDefaultConfig(conf)
	config.APIKey = values[EnvAPIKey]
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Joker DMAPI.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("joker: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		if config.Username == "" || config.Password == "" {
			return nil, errors.New("joker: credentials missing")
		}
	}

	if !strings.HasSuffix(config.BaseURL, "/") {
		config.BaseURL += "/"
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present installs a TXT record for the DNS challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	relative := getRelative(fqdn, zone)

	if d.config.Debug {
		log.Infof("[%s] joker: adding TXT record %q to zone %q with value %q", domain, relative, zone, value)
	}

	response, err := d.login()
	if err != nil {
		return formatResponseError(response, err)
	}

	response, err = d.getZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone := addTxtEntryToZone(response.Body, relative, value, d.config.TTL)

	response, err = d.putZone(zone, dnsZone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	return nil
}

// CleanUp removes a TXT record used for a previous DNS challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	relative := getRelative(fqdn, zone)

	if d.config.Debug {
		log.Infof("[%s] joker: removing entry %q from zone %q", domain, relative, zone)
	}

	response, err := d.login()
	if err != nil {
		return formatResponseError(response, err)
	}

	defer func() {
		// Try to logout in case of errors
		_, _ = d.logout()
	}()

	response, err = d.getZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone, modified := removeTxtEntryFromZone(response.Body, relative)
	if modified {
		response, err = d.putZone(zone, dnsZone)
		if err != nil || response.StatusCode != 0 {
			return formatResponseError(response, err)
		}
	}

	response, err = d.logout()
	if err != nil {
		return formatResponseError(response, err)
	}
	return nil
}

func getRelative(fqdn, zone string) string {
	return dns01.UnFqdn(strings.TrimSuffix(fqdn, dns01.ToFqdn(zone)))
}

// formatResponseError formats error with optional details from DMAPI response.
func formatResponseError(response *response, err error) error {
	if response != nil {
		return fmt.Errorf("joker: DMAPI error: %w Response: %v", err, response.Headers)
	}
	return fmt.Errorf("joker: DMAPI error: %w", err)
}
