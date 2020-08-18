// Package inwx implements a DNS provider for solving the DNS-01 challenge using inwx dom robot
package inwx

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/nrdcg/goinwx"
	"github.com/pquerna/otp/totp"
)

// Environment variables names.
const (
	envNamespace = "INWX_"

	EnvUsername     = envNamespace + "USERNAME"
	EnvPassword     = envNamespace + "PASSWORD"
	EnvSharedSecret = envNamespace + "SHARED_SECRET"
	EnvSandbox      = envNamespace + "SANDBOX"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username           string
	Password           string
	SharedSecret       string
	Sandbox            bool
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		Sandbox:            env.GetOrDefaultBool(conf, EnvSandbox, false),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *goinwx.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Dyn DNS.
// Credentials must be passed in the environment variables:
// INWX_USERNAME, INWX_PASSWORD, and INWX_SHARED_SECRET.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("inwx: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]
	config.SharedSecret = env.GetOrFile(conf, EnvSharedSecret)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Dyn DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("inwx: the configuration of the DNS provider is nil")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("inwx: credentials missing")
	}

	if config.Sandbox {
		log.Infof("inwx: sandbox mode is enabled")
	}

	client := goinwx.NewClient(config.Username, config.Password, &goinwx.ClientOptions{Sandbox: config.Sandbox})

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("inwx: %w", err)
	}

	info, err := d.client.Account.Login()
	if err != nil {
		return fmt.Errorf("inwx: %w", err)
	}

	defer func() {
		errL := d.client.Account.Logout()
		if errL != nil {
			log.Infof("inwx: failed to logout: %v", errL)
		}
	}()

	err = d.twoFactorAuth(info)
	if err != nil {
		return fmt.Errorf("inwx: %w", err)
	}

	request := &goinwx.NameserverRecordRequest{
		Domain:  dns01.UnFqdn(authZone),
		Name:    dns01.UnFqdn(fqdn),
		Type:    "TXT",
		Content: value,
		TTL:     d.config.TTL,
	}

	_, err = d.client.Nameservers.CreateRecord(request)
	if err != nil {
		switch er := err.(type) {
		case *goinwx.ErrorResponse:
			if er.Message == "Object exists" {
				return nil
			}
			return fmt.Errorf("inwx: %w", err)
		default:
			return fmt.Errorf("inwx: %w", err)
		}
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("inwx: %w", err)
	}

	info, err := d.client.Account.Login()
	if err != nil {
		return fmt.Errorf("inwx: %w", err)
	}

	defer func() {
		errL := d.client.Account.Logout()
		if errL != nil {
			log.Infof("inwx: failed to logout: %v", errL)
		}
	}()

	err = d.twoFactorAuth(info)
	if err != nil {
		return fmt.Errorf("inwx: %w", err)
	}

	response, err := d.client.Nameservers.Info(&goinwx.NameserverInfoRequest{
		Domain: dns01.UnFqdn(authZone),
		Name:   dns01.UnFqdn(fqdn),
		Type:   "TXT",
	})
	if err != nil {
		return fmt.Errorf("inwx: %w", err)
	}

	var lastErr error
	for _, record := range response.Records {
		err = d.client.Nameservers.DeleteRecord(record.ID)
		if err != nil {
			lastErr = fmt.Errorf("inwx: %w", err)
		}
	}

	return lastErr
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) twoFactorAuth(info *goinwx.LoginResponse) error {
	if info.TFA != "GOOGLE-AUTH" {
		return nil
	}

	if d.config.SharedSecret == "" {
		return errors.New("two factor authentication but no shared secret is given")
	}

	tan, err := totp.GenerateCode(d.config.SharedSecret, time.Now())
	if err != nil {
		return err
	}

	return d.client.Account.Unlock(tan)
}
