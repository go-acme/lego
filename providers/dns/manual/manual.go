package manual

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
)

const (
	dnsTemplate = `%s %d IN TXT %q`
)

// Environment variables names.
const (
	envNamespace = "MANUAL_"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.Provider = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider is an implementation of the ChallengeProvider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("manual: the configuration of the DNS provider is nil")
	}

	return &DNSProvider{config: config}, nil
}

// Present prints instructions for manually creating the TXT record.
func (*DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manual: could not find zone: %w", err)
	}

	fmt.Printf("lego: Please create the following TXT record in your %s zone:\n", authZone)
	fmt.Printf(dnsTemplate+"\n", info.EffectiveFQDN, dns01.DefaultTTL, info.Value)
	fmt.Printf("lego: Press 'Enter' when you are done\n")

	_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("manual: %w", err)
	}

	return nil
}

// CleanUp prints instructions for manually removing the TXT record.
func (*DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manual: could not find zone: %w", err)
	}

	fmt.Printf("lego: You can now remove this TXT record from your %s zone:\n", authZone)
	fmt.Printf(dnsTemplate+"\n", info.EffectiveFQDN, dns01.DefaultTTL, "...")

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return dns01.DefaultPropagationTimeout
}
