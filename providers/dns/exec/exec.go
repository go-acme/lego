// Package exec implements a DNS provider which runs a program for adding/removing the DNS record.
package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "EXEC_"

	EnvPath = envNamespace + "PATH"
	EnvMode = envNamespace + "MODE"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config Provider configuration.
type Config struct {
	Program            string
	Mode               string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider which runs the program in the
// environment variable EXEC_PATH for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPath)
	if err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}

	config := NewDefaultConfig()
	config.Program = values[EnvPath]
	config.Mode = os.Getenv(EnvMode)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig returns a new DNS provider which runs the given configuration
// for adding and removing the DNS record.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration is nil")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	var args []string
	if d.config.Mode == "RAW" {
		args = []string{"present", "--", domain, token, keyAuth}
	} else {
		fqdn, value := dns01.GetRecord(domain, keyAuth)
		args = []string{"present", fqdn, value}
	}

	cmd := exec.Command(d.config.Program, args...)

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}

	return err
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	var args []string
	if d.config.Mode == "RAW" {
		args = []string{"cleanup", "--", domain, token, keyAuth}
	} else {
		fqdn, value := dns01.GetRecord(domain, keyAuth)
		args = []string{"cleanup", fqdn, value}
	}

	cmd := exec.Command(d.config.Program, args...)

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}

	return err
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}
