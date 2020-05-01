// Package hover implements a DNS provider for solving the DNS-01 challenge using Hover DNS (past: "TuCows").
//
// This is based on attempting a python->go language conversion, and fit the smart parts from
// Dan Krause into the LeGo API.  See https://gist.github.com/dankrause/5585907
package hover

import (
	"errors"
	"net/url"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/hover/internal"
)

// Environment variables names.
const (
	envNamespace = "HOVER_"

	EnvDebug    = envNamespace + "DEBUG"
	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvFilename = envNamespace + "PASSFILE"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Username string
	Password string
	Server   string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
	ttl                uint
	hover              *internal.Client
	parsed             *url.URL
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig(u, p string) *Config {
	return &Config{
		ttl:                uint(env.GetOrDefaultInt(EnvTTL, 3600)),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		Username:           u,
		Password:           p,
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider is an implementation of the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Hover DNS.
// Credentials (username, password) must be passed in the environment variables; if they are
// undefined, look for a filename credential which will be parsed for the
// username/plaintextpassword (see hover_test.go for examples)
//
// NOTE: that the use of a password file is preferred to increase the burden to reap auth creds
func NewDNSProvider() (*DNSProvider, error) {
	if values, err := env.Get(EnvUsername, EnvPassword); err != nil {
		filename := ""
		if v2, e2 := env.Get(EnvFilename); e2 == nil { // check whether we can fallback
			filename = v2[EnvFilename]
		} else {
			return nil, err // nope; return original error
		}
		log.Infof("username (%s) and/or password (%s) environment variables not populated; reading from %s (%s)", EnvUsername, EnvPassword, EnvFilename, filename)

		if pta, err := internal.ReadConfigFile(filename); err == nil {
			log.Infof("username (%s) and/or password read from %s", pta.Username, filename)
			return NewDNSProviderConfig(NewDefaultConfig(pta.Username, pta.PlaintextPassword))
		}

		// give up: no config provided: return a zero-initialized provider
		return NewDNSProviderConfig(&Config{})
	} else { // successful environment pull
		return NewDNSProviderConfig(NewDefaultConfig(values[EnvUsername], values[EnvPassword]))
	}
}

// NewDNSProviderConfig return a DNSProvider instance configured for Hover
func NewDNSProviderConfig(config *Config) (d *DNSProvider, err error) {
	if config == nil {
		return nil, errors.New("hover: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, errors.New("hover: incomplete credentials, missing Hover Username")
	}
	if config.Password == "" {
		return nil, errors.New("hover: incomplete credentials, missing Hover Password")
	}

	config.hover = internal.NewClient(config.Username, config.Password, "", config.HTTPTimeout)

	return &DNSProvider{config: config, client: config.hover}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	return d.client.Upsert(dns01.UnFqdn(fqdn), domain, value, d.config.ttl)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	log.Infof(`deleting token "%s" from domain "%s" (fqdn: %s) on auth "%s"`, token, domain, fqdn, keyAuth)
	return d.client.Delete(dns01.UnFqdn(fqdn), domain)
}
