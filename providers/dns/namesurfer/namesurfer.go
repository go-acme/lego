// Package namesurfer implements a DNS provider for solving the DNS-01 challenge using FusionLayer NameSurfer API.
package namesurfer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/namesurfer/internal"
)

// Environment variables names.
const (
	envNamespace = "NAMESURFER_"

	EnvBaseURL   = envNamespace + "BASE_URL"
	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"
	EnvView      = envNamespace + "VIEW"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvInsecureSkipVerify = envNamespace + "INSECURE_SKIP_VERIFY"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL   string
	APIKey    string
	APISecret string
	View      string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for FusionLayer NameSurfer.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvBaseURL, EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("namesurfer: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvBaseURL]
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]
	config.View = env.GetOrDefaultString(EnvView, "")

	if env.GetOrDefaultBool(EnvInsecureSkipVerify, false) {
		config.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for FusionLayer NameSurfer.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namesurfer: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.BaseURL, config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("namesurfer: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("namesurfer: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("namesurfer: %w", err)
	}

	record := internal.DNSNode{
		Name: subDomain,
		Type: "TXT",
		TTL:  d.config.TTL,
		Data: info.Value,
	}

	err = d.client.AddDNSRecord(ctx, zone, d.config.View, record)
	if err != nil {
		return fmt.Errorf("namesurfer: add DNS record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("namesurfer: %w", err)
	}

	existing, err := d.client.SearchDNSHosts(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("namesurfer: search DNS hosts: %w", err)
	}

	for _, node := range existing {
		if node.Type != "TXT" || node.Data != info.Value {
			continue
		}

		err = d.client.UpdateDNSHost(ctx, zone, d.config.View, node, internal.DNSNode{})
		if err != nil {
			return fmt.Errorf("namesurfer: update DNS host: %w", err)
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (string, error) {
	zones, err := d.client.ListZones(ctx, "forward")
	if err != nil {
		return "", fmt.Errorf("list zones: %w", err)
	}

	domain := dns01.UnFqdn(fqdn)

	var zone string

	for _, z := range zones {
		if strings.HasSuffix(domain, z.Name) && len(z.Name) > len(zone) {
			zone = z.Name
		}
	}

	if zone == "" {
		return "", fmt.Errorf("no zone found for %s", fqdn)
	}

	return zone, nil
}
