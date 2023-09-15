// Package multi implements a DNS provider for solving the DNS-01 challenge using a bunch of sub DNS providers and farms out the calls to in parallel
package multi

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SubProviders []string
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	defaultProviders := []string{"route53", "gcloud"}
	return &Config{
		SubProviders: defaultProviders,
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	subProviders []challenge.Provider
}

// DNSProviderSequential implements challenge.Provider and sequential interfaces.
type DNSProviderSequential struct {
	DNSProvider
}

type sequential interface {
	Sequential() time.Duration
}

// NewDNSProvider returns a DNSProvider instance configured to use
// route53 and gcloud providers in sequence. See the documentation for
// those providers for a list of appliciable environment variables.
func NewDNSProvider() (challenge.Provider, error) {
	return NewDNSProviderConfig(NewDefaultConfig())
}

// NewDNSProviderConfig returns an DNSProvider or
// DNSProviderSequential instance that implements challenge.Provider.
// If one of the providers passed in via config implements a
// Sequential() function, a DNSProviderSequential is
// returned. Otherwise a DNSProvider is returned.
func NewDNSProviderConfig(config *Config) (challenge.Provider, error) {
	if config == nil {
		return nil, errors.New("multi: the configuration of the DNS provider is nil")
	}

	return NewDNSProviderByNames(config.SubProviders...)
}

// NewDNSProviderByNames returns an DNSProvider or
// DNSProviderSequential instance that implements challenge.Provider.
// If one of the providers passed in via config implements a
// Sequential() function, a DNSProviderSequential is
// returned. Otherwise a DNSProvider is returned.
func NewDNSProviderByNames(names ...string) (challenge.Provider, error) {
	sequentialType := false
	subProviders := make([]challenge.Provider, len(names))

	for i, name := range names {
		subProvider, err := dns.NewDNSChallengeProviderByName(name)
		if err != nil {
			return nil, fmt.Errorf("multi: error creating subprovider '%s': %w", name, err)
		}
		subProviders[i] = subProvider
		if _, ok := subProvider.(sequential); ok {
			sequentialType = true
		}
	}

	provider := DNSProvider{subProviders: subProviders}
	if sequentialType {
		return &DNSProviderSequential{DNSProvider: provider}, nil
	}
	return &provider, nil
}

// NewDNSProviderFromOthers returns an DNSProvider or
// DNSProviderSequential instance that implements challenge.Provider,
// and passes through all interface calls to the providers passed in.
// If one of the providers passed in via config implements a
// Sequential() function, a DNSProviderSequential is
// returned. Otherwise a DNSProvider is returned.
func NewDNSProviderFromOthers(providers ...challenge.Provider) challenge.Provider {
	provider := DNSProvider{subProviders: providers}

	for _, subProvider := range providers {
		if _, ok := subProvider.(sequential); ok {
			return &DNSProviderSequential{DNSProvider: provider}
		}
	}

	return &provider
}

// Timeout returns the timeout and interval to use when checking for
// DNS propagation.  Return the longest timeout and shortest interval
// of all the subproviders. This is probably the right behavior, since
// we have to wait for the longest timeout any way. In the worst case
// we will poll providers more often than actually necessary.
func (d *DNSProvider) Timeout() (maxTimeout, minInterval time.Duration) {
	maxTimeout = dns01.DefaultPropagationTimeout
	minInterval = dns01.DefaultPollingInterval

	for _, provider := range d.subProviders {
		if p, ok := provider.(challenge.ProviderTimeout); ok {
			timeout, interval := p.Timeout()
			if timeout > maxTimeout {
				maxTimeout = timeout
			}
			if interval < minInterval {
				minInterval = interval
			}
		}
	}

	return maxTimeout, minInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProviderSequential) Sequential() time.Duration {
	maxSequenceInterval := dns01.DefaultPropagationTimeout

	for _, provider := range d.subProviders {
		if p, ok := provider.(sequential); ok {
			sequenceInterval := p.Sequential()
			if sequenceInterval > maxSequenceInterval {
				maxSequenceInterval = sequenceInterval
			}
		}
	}

	return maxSequenceInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	for _, provider := range d.subProviders {
		err := provider.Present(domain, token, keyAuth)
		if err != nil {
			return fmt.Errorf("multi: subprovider failed to present: %w", err)
		}
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	for _, provider := range d.subProviders {
		err := provider.CleanUp(domain, token, keyAuth)
		if err != nil {
			return fmt.Errorf("multi: failed to clean up: %w", err)
		}
	}
	return nil
}
