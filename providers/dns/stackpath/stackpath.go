package stackpath

import (
	"context"
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

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Stackpath.
func NewDNSProviderCredentials(clientID, clientSecret, stackID string) (*DNSProvider, error) {
	defaultConfig := NewDefaultConfig()
	defaultConfig.StackID = stackID

	oathConfig := &clientcredentials.Config{
		TokenURL:     defaultAuthURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	httpClient := oathConfig.Client(context.Background())

	return &DNSProvider{
		client: httpClient,
		config: defaultConfig,
	}, nil
}

// NewDNSProvider returns a DNSProvider instance configured for Stackpath.
// Credentials must be passed in the environment variables:
// STACKPATH_CLIENT_ID, STACKPATH_CLIENT_SECRET, and STACKPATH_STACK_ID.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("STACKPATH_CLIENT_ID", "STACKPATH_CLIENT_SECRET", "STACKPATH_STACK_ID")
	if err != nil {
		return nil, err
	}

	return NewDNSProviderCredentials(
		values["STACKPATH_CLIENT_ID"],
		values["STACKPATH_CLIENT_SECRET"],
		values["STACKPATH_STACK_ID"],
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

	body := map[string]interface{}{
		"name": parts[0],
		"type": "TXT",
		"ttl":  d.config.TTL,
		"data": value,
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
