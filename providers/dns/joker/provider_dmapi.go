package joker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/joker/internal/dmapi"
)

var _ challenge.ProviderTimeout = (*dmapiProvider)(nil)

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

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &dmapiProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *dmapiProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *dmapiProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("joker: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	if d.config.Debug {
		log.Infof("[%s] joker: adding TXT record %q to zone %q with value %q", domain, subDomain, zone, info.Value)
	}

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return err
	}

	response, err := d.client.GetZone(ctx, zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone := dmapi.AddTxtEntryToZone(response.Body, subDomain, info.Value, d.config.TTL)

	response, err = d.client.PutZone(ctx, zone, dnsZone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *dmapiProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("joker: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	if d.config.Debug {
		log.Infof("[%s] joker: removing entry %q from zone %q", domain, subDomain, zone)
	}

	ctx, err := d.client.CreateAuthenticatedContext(context.Background())
	if err != nil {
		return err
	}

	defer func() {
		// Try to log out in case of errors
		_, _ = d.client.Logout(ctx)
	}()

	response, err := d.client.GetZone(ctx, zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone, modified := dmapi.RemoveTxtEntryFromZone(response.Body, subDomain)
	if modified {
		response, err = d.client.PutZone(ctx, zone, dnsZone)
		if err != nil || response.StatusCode != 0 {
			return formatResponseError(response, err)
		}
	}

	response, err = d.client.Logout(ctx)
	if err != nil {
		return formatResponseError(response, err)
	}

	return nil
}

// formatResponseError formats error with optional details from DMAPI response.
func formatResponseError(response *dmapi.Response, err error) error {
	if response != nil {
		return fmt.Errorf("joker: DMAPI error: %w Response: %v", err, response.Headers)
	}

	return fmt.Errorf("joker: DMAPI error: %w", err)
}
