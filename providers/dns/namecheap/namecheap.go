// Package namecheap implements a DNS provider for solving the DNS-01 challenge using namecheap DNS.
package namecheap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/namecheap/internal"
	"golang.org/x/net/publicsuffix"
)

// Notes about namecheap's tool API:
// 1. Using the API requires registration.
//    Once registered, use your account name and API key to access the API.
// 2. There is no API to add or modify a single DNS record.
//    Instead, you must read the entire list of records, make modifications,
//    and then write the entire updated list of records. (Yuck.)
// 3. Namecheap's DNS updates can be slow to propagate.
//    I've seen them take as long as an hour.
// 4. Namecheap requires you to whitelist the IP address from which you call its APIs.
//    It also requires all API calls to include the whitelisted IP address as a form or query string value.
//    This code uses a namecheap service to query the client's IP address.

// Environment variables names.
const (
	envNamespace = "NAMECHEAP_"

	EnvAPIUser = envNamespace + "API_USER"
	EnvAPIKey  = envNamespace + "API_KEY"

	EnvSandbox = envNamespace + "SANDBOX"
	EnvDebug   = envNamespace + "DEBUG"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// A challenge represents all the data needed to specify a dns-01 challenge to lets-encrypt.
type challenge struct {
	domain   string
	key      string
	keyFqdn  string
	keyValue string
	tld      string
	sld      string
	host     string
}

// newChallenge builds a challenge record from a domain name and a challenge authentication key.
func newChallenge(domain, keyAuth string) (*challenge, error) {
	domain = dns01.UnFqdn(domain)

	tld, _ := publicsuffix.PublicSuffix(domain)
	if tld == domain {
		return nil, fmt.Errorf("invalid domain name %q", domain)
	}

	parts := strings.Split(domain, ".")
	longest := len(parts) - strings.Count(tld, ".") - 1
	sld := parts[longest-1]

	var host string
	if longest >= 1 {
		host = strings.Join(parts[:longest-1], ".")
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)

	return &challenge{
		domain:   domain,
		key:      "_acme-challenge." + host,
		keyFqdn:  info.EffectiveFQDN,
		keyValue: info.Value,
		tld:      tld,
		sld:      sld,
		host:     host,
	}, nil
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Debug              bool
	BaseURL            string
	APIUser            string
	APIKey             string
	ClientIP           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	baseURL := internal.DefaultBaseURL
	if env.GetOrDefaultBool(EnvSandbox, false) {
		baseURL = internal.SandboxBaseURL
	}

	return &Config{
		BaseURL:            baseURL,
		Debug:              env.GetOrDefaultBool(EnvDebug, false),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 60*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 15*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 60*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for namecheap.
// Credentials must be passed in the environment variables:
// NAMECHEAP_API_USER and NAMECHEAP_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUser, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("namecheap: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values[EnvAPIUser]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Namecheap.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namecheap: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, errors.New("namecheap: credentials missing")
	}

	if config.ClientIP == "" {
		clientIP, err := internal.GetClientIP(context.Background(), config.HTTPClient, config.Debug)
		if err != nil {
			return nil, fmt.Errorf("namecheap: %w", err)
		}
		config.ClientIP = clientIP
	}

	client := internal.NewClient(config.APIUser, config.APIKey, config.ClientIP)
	client.BaseURL = config.BaseURL

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Namecheap can sometimes take a long time to complete an update, so wait up to 60 minutes for the update to propagate.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present installs a TXT record for the DNS challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	// TODO(ldez) replace domain by FQDN to follow CNAME.
	ch, err := newChallenge(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("namecheap: %w", err)
	}

	ctx := context.Background()

	records, err := d.client.GetHosts(ctx, ch.sld, ch.tld)
	if err != nil {
		return fmt.Errorf("namecheap: %w", err)
	}

	record := internal.Record{
		Name:    ch.key,
		Type:    "TXT",
		Address: ch.keyValue,
		MXPref:  "10",
		TTL:     strconv.Itoa(d.config.TTL),
	}

	records = append(records, record)

	if d.config.Debug {
		for _, h := range records {
			log.Printf("%-5.5s %-30.30s %-6s %-70.70s", h.Type, h.Name, h.TTL, h.Address)
		}
	}

	err = d.client.SetHosts(ctx, ch.sld, ch.tld, records)
	if err != nil {
		return fmt.Errorf("namecheap: %w", err)
	}
	return nil
}

// CleanUp removes a TXT record used for a previous DNS challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	// TODO(ldez) replace domain by FQDN to follow CNAME.
	ch, err := newChallenge(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("namecheap: %w", err)
	}

	ctx := context.Background()

	records, err := d.client.GetHosts(ctx, ch.sld, ch.tld)
	if err != nil {
		return fmt.Errorf("namecheap: %w", err)
	}

	// Find the challenge TXT record and remove it if found.
	var found bool
	var newRecords []internal.Record
	for _, h := range records {
		if h.Name == ch.key && h.Type == "TXT" {
			found = true
		} else {
			newRecords = append(newRecords, h)
		}
	}

	if !found {
		return nil
	}

	err = d.client.SetHosts(ctx, ch.sld, ch.tld, newRecords)
	if err != nil {
		return fmt.Errorf("namecheap: %w", err)
	}
	return nil
}
