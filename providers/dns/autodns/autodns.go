package autodns

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

const (
	envAPIUser            = "AUTODNS_API_USER"
	envAPIPassword        = "AUTODNS_API_PASSWORD"
	envAPIEndpoint        = "AUTODNS_ENDPOINT"
	envAPIEndpointContext = "AUTODNS_CONTEXT"
	envTTL                = "AUTODNS_TTL"
	envPropagationTimeout = "AUTODNS_PROPAGATION_TIMEOUT"
	envPollingInterval    = "AUTODNS_POLLING_INTERVAL"
	envHTTPTimeout        = "AUTODNS_HTTP_TIMEOUT"
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
	endpoint, _ := url.Parse(env.GetOrDefaultString(envAPIEndpoint, defaultEndpoint))

	return &Config{
		Endpoint:           endpoint,
		Context:            env.GetOrDefaultInt(envAPIEndpointContext, defaultEndpointContext),
		TTL:                env.GetOrDefaultInt(envTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(envPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(envPollingInterval, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(envHTTPTimeout, 30*time.Second),
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
	values, err := env.Get(envAPIUser, envAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("autodns: %v", err)
	}

	config := NewDefaultConfig()
	config.Username = values[envAPIUser]
	config.Password = values[envAPIPassword]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("autodns: config is nil")
	}

	if config.Username == "" {
		return nil, fmt.Errorf("autodns: missing user")
	}

	if config.Password == "" {
		return nil, fmt.Errorf("autodns: missing password")
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
		return fmt.Errorf("autodns: %v", err)
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
		return fmt.Errorf("autodns: %v", err)
	}

	return nil
}
