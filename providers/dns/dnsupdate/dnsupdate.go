// Package dnsupdate implements a DNS provider for solving the DNS-01 challenge using the RFC2136 dynamic update.
package dnsupdate

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/bodgit/tsig"
	"github.com/bodgit/tsig/gss"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/dnsupdate/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "DNSUPDATE_"

	EnvNameserver = envNamespace + "NAMESERVER"
	EnvDNSTimeout = envNamespace + "DNS_TIMEOUT"

	// Old environment variable name from lego v0.
	// TODO(ldez): remove in the future.
	envTimeout = envNamespace + "TIMEOUT"

	EnvZones = envNamespace + "ZONES"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Environment variables names related to TSIG.
const (
	envTSIG = envNamespace + "TSIG_"

	EnvTSIGFile = envTSIG + "FILE"

	EnvTSIGKey       = envTSIG + "KEY"
	EnvTSIGSecret    = envTSIG + "SECRET"
	EnvTSIGAlgorithm = envTSIG + "ALGORITHM"
)

// Environment variables names related to GSS-TSIG.
const (
	envSubTSIGGSS = "TSIG_GSS_"

	envTSIGGSS = envNamespace + envSubTSIGGSS

	EnvTSIGGSSRealm      = envTSIGGSS + "REALM"
	EnvTSIGGSSUsername   = envTSIGGSS + "USERNAME"
	EnvTSIGGSSPassword   = envTSIGGSS + "PASSWORD"
	EnvTSIGGSSKeytabFile = envTSIGGSS + "KEYTAB_FILE"
)

const (
	actionRemove = "REMOVE"
	actionInsert = "INSERT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Nameserver string
	DNSTimeout time.Duration

	Zones []string

	TSIGFile string

	TSIGAlgorithm string
	TSIGKey       string
	TSIGSecret    string

	TSIGGSSRealm      string
	TSIGGSSUsername   string
	TSIGGSSPassword   string
	TSIGGSSKeytabFile string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	SequenceInterval   time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TSIGAlgorithm: getOrDefaultString(EnvTSIGAlgorithm, dns.HmacSHA1),
		DNSTimeout:    getOrDefaultSecond(EnvDNSTimeout, 10*time.Second),
		TTL:           getOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOneWithFallback(
			EnvPropagationTimeout,
			dns01.DefaultPropagationTimeout,
			env.ParseSecond,
			slices.Concat(altEnvNames(EnvPropagationTimeout), []string{envTimeout})...,
		),
		PollingInterval:  getOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval: getOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for dnsupdate (RFC2136)
// dynamic update. Configured with environment variables:
// DNSUPDATE_NAMESERVER: Network address in the form "host" or "host:port".
// DNSUPDATE_TSIG_ALGORITHM: Defaults to hmac-md5.sig-alg.reg.int. (HMAC-MD5).
// See https://github.com/miekg/dns/blob/master/tsig.go for supported values.
// DNSUPDATE_TSIG_KEY: Name of the secret key as defined in DNS server configuration.
// DNSUPDATE_TSIG_SECRET: Secret key payload.
// DNSUPDATE_PROPAGATION_TIMEOUT: DNS propagation timeout in time.ParseDuration format. (60s)
// To disable TSIG authentication, leave the DNSUPDATE_TSIG* variables unset.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.GetWithFallback(
		slices.Concat([]string{EnvNameserver}, altEnvNames(EnvNameserver)),
	)
	if err != nil {
		return nil, fmt.Errorf("dnsupdate: %w", err)
	}

	config := NewDefaultConfig()
	config.Nameserver = values[EnvNameserver]

	config.Zones = getEnvStringSlice(EnvZones)

	config.TSIGFile = getEnvString(EnvTSIGFile)

	config.TSIGKey = getEnvString(EnvTSIGKey)
	config.TSIGSecret = getEnvString(EnvTSIGSecret)

	config.TSIGGSSRealm = getEnvString(EnvTSIGGSSRealm)
	config.TSIGGSSUsername = getEnvString(EnvTSIGGSSUsername)
	config.TSIGGSSPassword = getEnvString(EnvTSIGGSSPassword)
	config.TSIGGSSKeytabFile = getEnvString(EnvTSIGGSSKeytabFile)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for rfc2136.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsupdate: the configuration of the DNS provider is nil")
	}

	if config.Nameserver == "" {
		return nil, errors.New("dnsupdate: nameserver missing")
	}

	// Append the default DNS port if none is specified.
	if _, _, err := net.SplitHostPort(config.Nameserver); err != nil {
		if strings.Contains(err.Error(), "missing port") {
			config.Nameserver = net.JoinHostPort(config.Nameserver, "53")
		} else {
			return nil, fmt.Errorf("dnsupdate: %w", err)
		}
	}

	err := setupTSIG(config)
	if err != nil {
		return nil, fmt.Errorf("dnsupdate: %w", err)
	}

	slices.SortFunc(config.Zones, func(a, b string) int {
		return cmp.Compare(len(dns.Split(b)), len(dns.Split(a)))
	})

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
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	err := d.changeRecord(ctx, actionInsert, info.EffectiveFQDN, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("dnsupdate: failed to insert: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	err := d.changeRecord(ctx, actionRemove, info.EffectiveFQDN, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("dnsupdate: failed to remove: %w", err)
	}

	return nil
}

func (d *DNSProvider) changeRecord(ctx context.Context, action, fqdn, value string, ttl int) error {
	// Find the zone for the given fqdn
	zone, err := d.findZone(ctx, fqdn)
	if err != nil {
		return err
	}

	// Create RR
	rrs := []dns.RR{&dns.TXT{
		Hdr: dns.RR_Header{Name: fqdn, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: uint32(ttl)},
		Txt: []string{value},
	}}

	// Create dynamic update packet
	m := new(dns.Msg).SetUpdate(zone)

	switch action {
	case actionInsert:
		// Always remove old challenge left over from who knows what.
		m.RemoveRRset(rrs)
		m.Insert(rrs)
	case actionRemove:
		m.Remove(rrs)
	default:
		return fmt.Errorf("unexpected action: %s", action)
	}

	// Setup client
	c := &dns.Client{Timeout: d.config.DNSTimeout}

	// TSIG authentication / msg signing
	if d.config.TSIGAlgorithm == tsig.GSS {
		c.Net = "tcp"

		var gssClient *gss.Client

		gssClient, err = gss.NewClient(c)
		if err != nil {
			return fmt.Errorf("create GSS client: %w", err)
		}

		defer func() { _ = gssClient.Close() }()

		var keyName string

		keyName, err = d.negotiate(gssClient)
		if err != nil {
			return err
		}

		defer func() { _ = gssClient.DeleteContext(keyName) }()

		c.TsigProvider = gssClient

		m.SetTsig(keyName, tsig.GSS, 300, time.Now().Unix())
	} else if d.config.TSIGKey != "" && d.config.TSIGSecret != "" {
		m.SetTsig(d.config.TSIGKey, d.config.TSIGAlgorithm, 300, time.Now().Unix())

		// Secret(s) for TSIG map[<zonename>]<base64 secret>.
		c.TsigSecret = map[string]string{d.config.TSIGKey: d.config.TSIGSecret}
	}

	// Send the query
	reply, _, err := c.ExchangeContext(ctx, m, d.config.Nameserver)
	if err != nil {
		return fmt.Errorf("DNS update failed: %w", err)
	}

	if reply != nil && reply.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update failed: server %s replied %s for %s %s", d.config.Nameserver, dns.RcodeToString[reply.Rcode], action, zone)
	}

	return nil
}

func (d *DNSProvider) negotiate(client *gss.Client) (string, error) {
	if d.config.TSIGGSSKeytabFile != "" {
		keyName, _, err := client.NegotiateContextWithKeytab(
			d.config.Nameserver,
			d.config.TSIGGSSRealm,
			d.config.TSIGGSSUsername,
			d.config.TSIGGSSKeytabFile,
		)
		if err != nil {
			return "", fmt.Errorf("negotiate GSS context with keytab: %w", err)
		}

		return keyName, nil
	}

	keyName, _, err := client.NegotiateContextWithCredentials(
		d.config.Nameserver,
		d.config.TSIGGSSRealm,
		d.config.TSIGGSSUsername,
		d.config.TSIGGSSPassword,
	)
	if err != nil {
		return "", fmt.Errorf("negotiate GSS context with credentials: %w", err)
	}

	return keyName, nil
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (string, error) {
	if len(d.config.Zones) == 0 {
		return dns01.DefaultClient().FindZoneByFqdnCustom(ctx, fqdn, []string{d.config.Nameserver})
	}

	for potentialZone := range dns01.DomainsSeq(fqdn) {
		for _, zone := range d.config.Zones {
			z := dns.Fqdn(zone)

			if strings.HasSuffix(potentialZone, z) {
				return z, nil
			}
		}
	}

	return "", fmt.Errorf("zone for %s not found", fqdn)
}

func setupTSIG(config *Config) error {
	if dns.Fqdn(config.TSIGAlgorithm) == tsig.GSS {
		err := validateTSIGGSS(config)
		if err != nil {
			return fmt.Errorf("TSIG GSS: %w", err)
		}
	} else {
		err := prepareTSIG(config)
		if err != nil {
			return err
		}
	}

	if config.TSIGAlgorithm == "" {
		config.TSIGAlgorithm = dns.HmacSHA1
	} else {
		// To be compatible with https://github.com/miekg/dns/blob/master/tsig.go
		config.TSIGAlgorithm = dns.Fqdn(config.TSIGAlgorithm)
	}

	switch config.TSIGAlgorithm {
	case dns.HmacSHA1, dns.HmacSHA224, dns.HmacSHA256, dns.HmacSHA384, dns.HmacSHA512, tsig.GSS:
		// valid algorithm
	default:
		return fmt.Errorf("unsupported TSIG algorithm: %s", config.TSIGAlgorithm)
	}

	return nil
}

func validateTSIGGSS(config *Config) error {
	if config.TSIGGSSUsername == "" {
		return errors.New("username is required")
	}

	if config.TSIGGSSPassword == "" && config.TSIGGSSKeytabFile == "" {
		return errors.New("password or keytab path is required")
	}

	if config.TSIGGSSPassword != "" && config.TSIGGSSKeytabFile != "" {
		return errors.New("only one of the password and keytab paths can be set")
	}

	if config.TSIGFile != "" {
		return errors.New("the TSIG file is not supported")
	}

	if config.TSIGKey != "" || config.TSIGSecret != "" {
		return errors.New("SIG key and secret are not supported")
	}

	return nil
}

func prepareTSIG(config *Config) error {
	if config.TSIGFile != "" {
		key, err := internal.ReadTSIGFile(config.TSIGFile)
		if err != nil {
			return fmt.Errorf("read TSIG file %s: %w", config.TSIGFile, err)
		}

		config.TSIGAlgorithm = key.Algorithm
		config.TSIGKey = key.Name
		config.TSIGSecret = key.Secret
	}

	if config.TSIGKey == "" || config.TSIGSecret == "" {
		config.TSIGKey = ""
		config.TSIGSecret = ""
	} else {
		// zonename must be in canonical form (lowercase, fqdn, see RFC 4034 Section 6.2)
		config.TSIGKey = dns.CanonicalName(config.TSIGKey)
	}

	return nil
}
