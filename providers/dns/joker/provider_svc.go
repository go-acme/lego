package joker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/config/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/joker/internal/svc"
)

var _ challenge.ProviderTimeout = (*svcProvider)(nil)

// svcProvider implements the challenge.Provider interface.
type svcProvider struct {
	config *Config
	client *svc.Client
}

// newSvcProvider returns a DNSProvider instance configured for Joker.
// Credentials must be passed in the environment variable: JOKER_USERNAME, JOKER_PASSWORD.
func newSvcProvider() (*svcProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("joker: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return newSvcProviderConfig(config)
}

// newSvcProviderConfig return a DNSProvider instance configured for Joker.
func newSvcProviderConfig(config *Config) (*svcProvider, error) {
	if config == nil {
		return nil, errors.New("joker: the configuration of the DNS provider is nil")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("joker: credentials missing")
	}

	client := svc.NewClient(config.Username, config.Password)

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &svcProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *svcProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *svcProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("joker: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	err = d.client.SendRequest(ctx, dns01.UnFqdn(zone), subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("joker: send request: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *svcProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("joker: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	err = d.client.SendRequest(ctx, dns01.UnFqdn(zone), subDomain, "")
	if err != nil {
		return fmt.Errorf("joker: send request: %w", err)
	}

	return nil
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *svcProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}
