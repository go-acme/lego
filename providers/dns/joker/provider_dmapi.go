package joker

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/joker/internal/dmapi"
)

// dmapiProvider implements the challenge.Provider interface.
type dmapiProvider struct {
	config *Config
	client *dmapi.Client
}

// newDmapiProvider returns a DNSProvider instance configured for Joker.
// Credentials must be passed in the environment variable: JOKER_USERNAME, JOKER_PASSWORD or JOKER_API_KEY.
func newDmapiProvider() (*dmapiProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		var errU error
		values, errU = env.Get(EnvUsername, EnvPassword)
		if errU != nil {
			//nolint:errorlint // false-positive
			return nil, fmt.Errorf("joker: %v or %v", errU, err)
		}
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return newDmapiProviderConfig(config)
}

// newDmapiProviderConfig return a DNSProvider instance configured for Joker.
func newDmapiProviderConfig(config *Config) (*dmapiProvider, error) {
	if config == nil {
		return nil, errors.New("joker: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		if config.Username == "" || config.Password == "" {
			return nil, errors.New("joker: credentials missing")
		}
	}

	client := dmapi.NewClient(dmapi.AuthInfo{
		APIKey:   config.APIKey,
		Username: config.Username,
		Password: config.Password,
	})

	client.Debug = config.Debug

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &dmapiProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *dmapiProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *dmapiProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	relative := getRelative(fqdn, zone)

	if d.config.Debug {
		log.Infof("[%s] joker: adding TXT record %q to zone %q with value %q", domain, relative, zone, value)
	}

	response, err := d.client.Login()
	if err != nil {
		return formatResponseError(response, err)
	}

	response, err = d.client.GetZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone := dmapi.AddTxtEntryToZone(response.Body, relative, value, d.config.TTL)

	response, err = d.client.PutZone(zone, dnsZone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *dmapiProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	relative := getRelative(fqdn, zone)

	if d.config.Debug {
		log.Infof("[%s] joker: removing entry %q from zone %q", domain, relative, zone)
	}

	response, err := d.client.Login()
	if err != nil {
		return formatResponseError(response, err)
	}

	defer func() {
		// Try to logout in case of errors
		_, _ = d.client.Logout()
	}()

	response, err = d.client.GetZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone, modified := dmapi.RemoveTxtEntryFromZone(response.Body, relative)
	if modified {
		response, err = d.client.PutZone(zone, dnsZone)
		if err != nil || response.StatusCode != 0 {
			return formatResponseError(response, err)
		}
	}

	response, err = d.client.Logout()
	if err != nil {
		return formatResponseError(response, err)
	}
	return nil
}

func getRelative(fqdn, zone string) string {
	return dns01.UnFqdn(strings.TrimSuffix(fqdn, dns01.ToFqdn(zone)))
}

// formatResponseError formats error with optional details from DMAPI response.
func formatResponseError(response *dmapi.Response, err error) error {
	if response != nil {
		return fmt.Errorf("joker: DMAPI error: %w Response: %v", err, response.Headers)
	}
	return fmt.Errorf("joker: DMAPI error: %w", err)
}
