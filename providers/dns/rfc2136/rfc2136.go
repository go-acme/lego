// Package rfc2136 implements a DNS provider for solving the DNS-01 challenge using the rfc2136 dynamic update.
package rfc2136

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/rfc2136/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "RFC2136_"

	EnvTSIGFile = envNamespace + "TSIG_FILE"

	EnvTSIGKey       = envNamespace + "TSIG_KEY"
	EnvTSIGSecret    = envNamespace + "TSIG_SECRET"
	EnvTSIGAlgorithm = envNamespace + "TSIG_ALGORITHM"

	EnvNameserver = envNamespace + "NAMESERVER"
	EnvDNSTimeout = envNamespace + "DNS_TIMEOUT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Nameserver string

	TSIGFile string

	TSIGAlgorithm string
	TSIGKey       string
	TSIGSecret    string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	SequenceInterval   time.Duration
	DNSTimeout         time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TSIGAlgorithm:      env.GetOrDefaultString(EnvTSIGAlgorithm, dns.HmacSHA1),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, env.GetOrDefaultSecond("RFC2136_TIMEOUT", 60*time.Second)),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
		DNSTimeout:         env.GetOrDefaultSecond(EnvDNSTimeout, 10*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for rfc2136
// dynamic update. Configured with environment variables:
// RFC2136_NAMESERVER: Network address in the form "host" or "host:port".
// RFC2136_TSIG_ALGORITHM: Defaults to hmac-md5.sig-alg.reg.int. (HMAC-MD5).
// See https://github.com/miekg/dns/blob/master/tsig.go for supported values.
// RFC2136_TSIG_KEY: Name of the secret key as defined in DNS server configuration.
// RFC2136_TSIG_SECRET: Secret key payload.
// RFC2136_PROPAGATION_TIMEOUT: DNS propagation timeout in time.ParseDuration format. (60s)
// To disable TSIG authentication, leave the RFC2136_TSIG* variables unset.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvNameserver)
	if err != nil {
		return nil, fmt.Errorf("rfc2136: %w", err)
	}

	config := NewDefaultConfig()
	config.Nameserver = values[EnvNameserver]

	config.TSIGFile = env.GetOrDefaultString(EnvTSIGFile, "")

	config.TSIGKey = env.GetOrFile(EnvTSIGKey)
	config.TSIGSecret = env.GetOrFile(EnvTSIGSecret)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for rfc2136.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rfc2136: the configuration of the DNS provider is nil")
	}

	if config.Nameserver == "" {
		return nil, errors.New("rfc2136: nameserver missing")
	}

	if config.TSIGFile != "" {
		key, err := internal.ReadTSIGFile(config.TSIGFile)
		if err != nil {
			return nil, fmt.Errorf("rfc2136: read TSIG file %s: %w", config.TSIGFile, err)
		}

		config.TSIGAlgorithm = key.Algorithm
		config.TSIGKey = key.Name
		config.TSIGSecret = key.Secret
	}

	// Append the default DNS port if none is specified.
	if _, _, err := net.SplitHostPort(config.Nameserver); err != nil {
		if strings.Contains(err.Error(), "missing port") {
			config.Nameserver = net.JoinHostPort(config.Nameserver, "53")
		} else {
			return nil, fmt.Errorf("rfc2136: %w", err)
		}
	}

	if config.TSIGKey == "" || config.TSIGSecret == "" {
		config.TSIGKey = ""
		config.TSIGSecret = ""
	} else {
		// zonename must be in canonical form (lowercase, fqdn, see RFC 4034 Section 6.2)
		config.TSIGKey = strings.ToLower(dns.Fqdn(config.TSIGKey))
	}

	if config.TSIGAlgorithm == "" {
		config.TSIGAlgorithm = dns.HmacSHA1
	} else {
		// To be compatible with https://github.com/miekg/dns/blob/master/tsig.go
		config.TSIGAlgorithm = dns.Fqdn(config.TSIGAlgorithm)
	}

	switch config.TSIGAlgorithm {
	case dns.HmacSHA1, dns.HmacSHA224, dns.HmacSHA256, dns.HmacSHA384, dns.HmacSHA512:
		// valid algorithm
	default:
		return nil, fmt.Errorf("rfc2136: unsupported TSIG algorithm: %s", config.TSIGAlgorithm)
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.changeRecord("INSERT", info.EffectiveFQDN, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("rfc2136: failed to insert: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	err := d.changeRecord("REMOVE", info.EffectiveFQDN, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("rfc2136: failed to remove: %w", err)
	}
	return nil
}

func (d *DNSProvider) changeRecord(action, fqdn, value string, ttl int) error {
	// Find the zone for the given fqdn
	zone, err := dns01.FindZoneByFqdnCustom(fqdn, []string{d.config.Nameserver})
	if err != nil {
		return err
	}

	// Create RR
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: fqdn, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: uint32(ttl)}
	rr.Txt = []string{value}
	rrs := []dns.RR{rr}

	// Create dynamic update packet
	m := new(dns.Msg)
	m.SetUpdate(zone)
	switch action {
	case "INSERT":
		// Always remove old challenge left over from who knows what.
		m.RemoveRRset(rrs)
		m.Insert(rrs)
	case "REMOVE":
		m.Remove(rrs)
	default:
		return fmt.Errorf("unexpected action: %s", action)
	}

	// Setup client
	c := &dns.Client{Timeout: d.config.DNSTimeout}

	// TSIG authentication / msg signing
	if d.config.TSIGKey != "" && d.config.TSIGSecret != "" {
		m.SetTsig(d.config.TSIGKey, d.config.TSIGAlgorithm, 300, time.Now().Unix())

		// Secret(s) for TSIG map[<zonename>]<base64 secret>.
		c.TsigSecret = map[string]string{d.config.TSIGKey: d.config.TSIGSecret}
	}

	// Send the query
	reply, _, err := c.Exchange(m, d.config.Nameserver)
	if err != nil {
		return fmt.Errorf("DNS update failed: %w", err)
	}
	if reply != nil && reply.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update failed: server replied: %s", dns.RcodeToString[reply.Rcode])
	}

	return nil
}
