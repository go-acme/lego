package stackpath

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultBaseURL = "https://gateway.stackpath.com/dns/v1/stacks"
	defaultAuthURL = "https://gateway.stackpath.com/identity/v1/oauth2/token"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *http.Client
	config *Config
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	ClientID           string
	ClientSecret       string
	StackID            string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("STACKPATH_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("STACKPATH_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("STACKPATH_POLLING_INTERVAL", acme.DefaultPollingInterval),
	}
}

// NewDNSProviderConfig return a DNSProvider instance configured for Stackpath.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if len(config.ClientID) == 0 || len(config.ClientSecret) == 0 {
		return nil, errors.New("credentials missing")
	}

	if len(config.StackID) == 0 {
		return nil, errors.New("stack id missing")
	}

	return &DNSProvider{
		client: getOathClient(config),
		config: config,
	}, nil
}

func getOathClient(config *Config) *http.Client {
	oathConfig := &clientcredentials.Config{
		TokenURL:     defaultAuthURL,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
	}

	return oathConfig.Client(context.Background())
}

// NewDNSProvider returns a DNSProvider instance configured for Stackpath.
// Credentials must be passed in the environment variables:
// STACKPATH_CLIENT_ID, STACKPATH_CLIENT_SECRET, and STACKPATH_STACK_ID.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("STACKPATH_CLIENT_ID", "STACKPATH_CLIENT_SECRET", "STACKPATH_STACK_ID")
	if err != nil {
		return nil, err
	}

	return NewDNSProviderConfig(
		&Config{
			ClientID:     values["STACKPATH_CLIENT_ID"],
			ClientSecret: values["STACKPATH_CLIENT_SECRET"],
			StackID:      values["STACKPATH_STACK_ID"],
		},
	)
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	zone, err := d.getZoneForDomain(domain)
	if err != nil {
		return err
	}

	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	parts := strings.Split(fqdn, ".")

	body := Record{
		Name: parts[0],
		Type: "TXT",
		TTL:  d.config.TTL,
		Data: value,
	}

	return d.httpPost(fmt.Sprintf("/zones/%s/records", zone.ID), body)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	zone, err := d.getZoneForDomain(domain)
	if err != nil {
		return err
	}

	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	parts := strings.Split(fqdn, ".")

	record, err := d.getRecordForZone(parts[0], zone)
	if err != nil {
		return err
	}

	return d.httpDelete(fmt.Sprintf("/zones/%s/records/%s", zone.ID, record.ID))
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
