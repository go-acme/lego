package webnames

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/webnames/internal"
)

// Environment variables names.
const (
	defaultBaseURL = "https://www.webnames.ru/scripts/json_domain_zone_manager.pl"
	envNamespace   = "WEBNAMES_"

	EnvApiKey             = envNamespace + "APIKEY"
	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ApiKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
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

// NewDNSProvider returns a DNSProvider instance configured for reg.ru.
// Credentials must be passed in the environment variables:
// REGRU_USERNAME and REGRU_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvApiKey)
	if err != nil {
		return nil, fmt.Errorf("webnames: %w", err)
	}

	config := NewDefaultConfig()
	config.ApiKey = values[EnvApiKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for reg.ru.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("webnames: the configuration of the DNS provider is nil")
	}

	if config.ApiKey == "" {
		return nil, errors.New("webnames: incomplete credentials, missing username and/or password")
	}

	client := internal.NewClient(config.ApiKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

func (wn *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("webnames: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}
	subDomain := dns01.UnFqdn(strings.TrimSuffix(fqdn, authZone))

	q := bytes.NewBuffer(nil)
	q.WriteString(fmt.Sprintf("apikey=%s&", wn.config.ApiKey))
	q.WriteString(fmt.Sprintf("domain=%s&", dns01.UnFqdn(authZone)))
	q.WriteString(fmt.Sprintf("type=%s&", "TXT"))
	q.WriteString(fmt.Sprintf("record=%s:%s&", subDomain, value))
	q.WriteString(fmt.Sprintf("action=%s", "add"))

	req, err := http.NewRequest("POST", defaultBaseURL, strings.NewReader(q.String()))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	var r map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if x, ok := r["result"]; ok && x == "OK" {
		return nil
	}

	return fmt.Errorf("can not present")
}

func (wn *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("webnames: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	}
	subDomain := dns01.UnFqdn(strings.TrimSuffix(fqdn, authZone))

	q := bytes.NewBuffer(nil)
	q.WriteString(fmt.Sprintf("apikey=%s&", wn.config.ApiKey))
	q.WriteString(fmt.Sprintf("domain=%s&", dns01.UnFqdn(authZone)))
	q.WriteString(fmt.Sprintf("type=%s&", "TXT"))
	q.WriteString(fmt.Sprintf("record=%s:%s&", subDomain, value))
	q.WriteString(fmt.Sprintf("action=%s", "delete"))

	req, err := http.NewRequest("POST", defaultBaseURL, strings.NewReader(q.String()))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	var r map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if x, ok := r["result"]; ok && x == "OK" {
		return nil
	}

	return fmt.Errorf("can not cleanup")
}

func (wn *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return wn.config.PropagationTimeout, wn.config.PollingInterval
}
