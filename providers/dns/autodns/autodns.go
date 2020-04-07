package autodns

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "AUTODNS_"

	EnvAPIUser            = envNamespace + "API_USER"
	EnvAPIPassword        = envNamespace + "API_PASSWORD"
	EnvAPIEndpoint        = envNamespace + "ENDPOINT"
	EnvAPIEndpointContext = envNamespace + "CONTEXT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	defaultEndpointContext int = 4
	defaultTTL             int = 600
)

type Config struct {
	Endpoint           *url.URL
	Username           string
	Password           string
	Context            int
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

func NewDefaultConfig() *Config {
	endpoint, _ := url.Parse(env.GetOrDefaultString(EnvAPIEndpoint, defaultEndpoint))

	return &Config{
		Endpoint:           endpoint,
		Context:            env.GetOrDefaultInt(EnvAPIEndpointContext, defaultEndpointContext),
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

type DNSProvider struct {
	config *Config
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUser, EnvAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("autodns: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvAPIUser]
	config.Password = values[EnvAPIPassword]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("autodns: config is nil")
	}

	if config.Username == "" {
		return nil, errors.New("autodns: missing user")
	}

	if config.Password == "" {
		return nil, errors.New("autodns: missing password")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	records := []*ResourceRecord{{
		Name:  fqdn,
		TTL:   int64(d.config.TTL),
		Type:  "TXT",
		Value: value,
	}}

	_, err := d.addTxtRecord(domain, records)
	if err != nil {
		return fmt.Errorf("autodns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record previously created
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	records := []*ResourceRecord{{
		Name:  fqdn,
		TTL:   int64(d.config.TTL),
		Type:  "TXT",
		Value: value,
	}}

	if err := d.removeTXTRecord(domain, records); err != nil {
		return fmt.Errorf("autodns: %w", err)
	}

	return nil
}
