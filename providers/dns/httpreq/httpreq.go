// Package httpreq implements a DNS provider for solving the DNS-01 challenge through a HTTP server.
package httpreq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "HTTPREQ_"

	EnvEndpoint = envNamespace + "ENDPOINT"
	EnvMode     = envNamespace + "MODE"
	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

type message struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

type messageRaw struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"keyAuth"`
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Endpoint           *url.URL
	Mode               string
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
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
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvEndpoint)
	if err != nil {
		return nil, fmt.Errorf("httpreq: %w", err)
	}

	endpoint, err := url.Parse(values[EnvEndpoint])
	if err != nil {
		return nil, fmt.Errorf("httpreq: %w", err)
	}

	config := NewDefaultConfig()
	config.Mode = env.GetOrFile(EnvMode)
	config.Username = env.GetOrFile(EnvUsername)
	config.Password = env.GetOrFile(EnvPassword)
	config.Endpoint = endpoint
	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("httpreq: the configuration of the DNS provider is nil")
	}

	if config.Endpoint == nil {
		return nil, errors.New("httpreq: the endpoint is missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	if d.config.Mode == "RAW" {
		msg := &messageRaw{
			Domain:  domain,
			Token:   token,
			KeyAuth: keyAuth,
		}

		err := d.doPost("/present", msg)
		if err != nil {
			return fmt.Errorf("httpreq: %w", err)
		}
		return nil
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	msg := &message{
		FQDN:  fqdn,
		Value: value,
	}

	err := d.doPost("/present", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	if d.config.Mode == "RAW" {
		msg := &messageRaw{
			Domain:  domain,
			Token:   token,
			KeyAuth: keyAuth,
		}

		err := d.doPost("/cleanup", msg)
		if err != nil {
			return fmt.Errorf("httpreq: %w", err)
		}
		return nil
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	msg := &message{
		FQDN:  fqdn,
		Value: value,
	}

	err := d.doPost("/cleanup", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %w", err)
	}
	return nil
}

func (d *DNSProvider) doPost(uri string, msg interface{}) error {
	reqBody := &bytes.Buffer{}
	err := json.NewEncoder(reqBody).Encode(msg)
	if err != nil {
		return err
	}

	endpoint := d.config.Endpoint.JoinPath(uri)

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(d.config.Username) > 0 && len(d.config.Password) > 0 {
		req.SetBasicAuth(d.config.Username, d.config.Password)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%d: failed to read response body: %w", resp.StatusCode, err)
		}

		return fmt.Errorf("%d: request failed: %v", resp.StatusCode, string(body))
	}

	return nil
}
