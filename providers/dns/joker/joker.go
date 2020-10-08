// Package joker implements a DNS provider for solving the DNS-01 challenge using joker.com.
package joker

import (
	"net/http"
	"os"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "JOKER_"

	EnvAPIKey   = envNamespace + "API_KEY"
	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvDebug    = envNamespace + "DEBUG"
	EnvMode     = envNamespace + "API_MODE"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	modeDMAPI = "DMAPI"
	modeSVC   = "SVC"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Debug              bool
	APIKey             string
	Username           string
	Password           string
	APIMode            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		APIMode:            env.GetOrDefaultString(EnvMode, modeDMAPI),
		Debug:              env.GetOrDefaultBool(EnvDebug, false),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 60*time.Second),
		},
	}
}

// NewDNSProvider returns a DNSProvider instance configured for Joker.
// Credentials must be passed in the environment variable JOKER_API_KEY.
func NewDNSProvider() (challenge.ProviderTimeout, error) {
	if os.Getenv(EnvMode) == modeSVC {
		return newSvcProvider()
	}

	return newDmapiProvider()
}

// NewDNSProviderConfig return a DNSProvider instance configured for Joker.
func NewDNSProviderConfig(config *Config) (challenge.ProviderTimeout, error) {
	if config.APIMode == modeSVC {
		return newSvcProviderConfig(config)
	}

	return newDmapiProviderConfig(config)
}
