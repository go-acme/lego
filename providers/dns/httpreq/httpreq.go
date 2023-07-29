// Package httpreq implements a DNS provider for solving the DNS-01 challenge through an HTTP server.
package httpreq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
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
	ctx := context.Background()

	if d.config.Mode == "RAW" {
		msg := &messageRaw{
			Domain:  domain,
			Token:   token,
			KeyAuth: keyAuth,
		}

		err := d.doPost(ctx, "/present", msg)
		if err != nil {
			return fmt.Errorf("httpreq: %w", err)
		}
		return nil
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)
	msg := &message{
		FQDN:  info.EffectiveFQDN,
		Value: info.Value,
	}

	err := d.doPost(ctx, "/present", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	if d.config.Mode == "RAW" {
		msg := &messageRaw{
			Domain:  domain,
			Token:   token,
			KeyAuth: keyAuth,
		}

		err := d.doPost(ctx, "/cleanup", msg)
		if err != nil {
			return fmt.Errorf("httpreq: %w", err)
		}
		return nil
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)
	msg := &message{
		FQDN:  info.EffectiveFQDN,
		Value: info.Value,
	}

	err := d.doPost(ctx, "/cleanup", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %w", err)
	}
	return nil
}

func (d *DNSProvider) doPost(ctx context.Context, uri string, msg any) error {
	reqBody := new(bytes.Buffer)
	err := json.NewEncoder(reqBody).Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to create request JSON body: %w", err)
	}

	endpoint := d.config.Endpoint.JoinPath(uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if d.config.Username != "" && d.config.Password != "" {
		req.SetBasicAuth(d.config.Username, d.config.Password)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}
