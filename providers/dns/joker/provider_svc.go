package joker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/joker/internal/svc"
)

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

	return &svcProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *svcProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *svcProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("joker: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	return d.client.SendRequest(context.Background(), dns01.UnFqdn(zone), subDomain, info.Value)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *svcProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("joker: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("joker: %w", err)
	}

	return d.client.SendRequest(context.Background(), dns01.UnFqdn(zone), subDomain, "")
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *svcProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}
