// Package callback implements a DNS provider for solving the DNS-01 challenge using callback functions.
package callback

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "CALLBACK_"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

const (
	defaultPropagationTimeout = 3 * time.Minute
	defaultPollInterval       = 15 * time.Second
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	presentCallback    func(fqdn, recordBody string) error
	cleanupCallback    func(fqdn, recordBody string) error
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		presentCallback:    func(fqdn, recordBody string) error { return errors.New("not implemented") },
		cleanupCallback:    func(fqdn, recordBody string) error { return errors.New("not implemented") },
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	return d.config.presentCallback(dns01.GetRecord(domain, keyAuth))
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.config.cleanupCallback(dns01.GetRecord(domain, keyAuth))
}

// NewDNSProvider returns a callback DNSProvider instance.
func NewDNSProvider(presentCallback, cleanupCallback func(fqdn, recordBody string) error) (*DNSProvider, error) {
	if presentCallback == nil {
		return nil, errors.New("callback: got nil presentCallback")
	}
	if cleanupCallback == nil {
		return nil, errors.New("callback: got nil cleanupCallback")
	}

	config := NewDefaultConfig()
	config.presentCallback = presentCallback
	config.cleanupCallback = cleanupCallback

	return &DNSProvider{config: config}, nil
}
