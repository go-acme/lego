package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/challenge/dnspersist01"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/challenge/tlsalpn01"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/providers/http/memcached"
	"github.com/go-acme/lego/v5/providers/http/s3"
	"github.com/go-acme/lego/v5/providers/http/webroot"
	"github.com/urfave/cli/v3"
)

func setupChallenges(cmd *cli.Command, client *lego.Client) error {
	if cmd.Bool(flags.FlgHTTP) {
		err := setupHTTPProvider(cmd, client)
		if err != nil {
			return fmt.Errorf("HTTP challenge provider: %w", err)
		}
	}

	if cmd.Bool(flags.FlgTLS) {
		err := setupTLSProvider(cmd, client)
		if err != nil {
			return fmt.Errorf("TLS challenge provider: %w", err)
		}
	}

	if cmd.IsSet(flags.FlgDNS) {
		err := setupDNS(cmd, client)
		if err != nil {
			return fmt.Errorf("DNS challenge provider: %w", err)
		}
	}

	if cmd.Bool(flags.FlgDNSPersist) {
		err := setupDNSPersist(cmd, client)
		if err != nil {
			return fmt.Errorf("DNS-PERSIST challenge provider: %w", err)
		}
	}

	return nil
}

func setupHTTPProvider(cmd *cli.Command, client *lego.Client) error {
	provider, err := createHTTPProvider(cmd)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return client.Challenge.SetHTTP01Provider(provider, http01.SetDelay(cmd.Duration(flags.FlgHTTPDelay)))
}

func createHTTPProvider(cmd *cli.Command) (challenge.Provider, error) {
	switch {
	case cmd.IsSet(flags.FlgHTTPWebroot):
		ps, err := webroot.NewHTTPProvider(cmd.String(flags.FlgHTTPWebroot))
		if err != nil {
			return nil, fmt.Errorf("webroot provider (%s): %w", cmd.String(flags.FlgHTTPWebroot), err)
		}

		return ps, nil

	case cmd.IsSet(flags.FlgHTTPMemcachedHost):
		ps, err := memcached.NewMemcachedProvider(cmd.StringSlice(flags.FlgHTTPMemcachedHost))
		if err != nil {
			return nil, fmt.Errorf("memcached provider (%s): %w", strings.Join(cmd.StringSlice(flags.FlgHTTPMemcachedHost), ", "), err)
		}

		return ps, nil

	case cmd.IsSet(flags.FlgHTTPS3Bucket):
		ps, err := s3.NewHTTPProvider(cmd.String(flags.FlgHTTPS3Bucket))
		if err != nil {
			return nil, fmt.Errorf("S3 provider (%s): %w", cmd.String(flags.FlgHTTPS3Bucket), err)
		}

		return ps, nil

	case cmd.IsSet(flags.FlgHTTPAddress):
		host, port, err := parseAddress(cmd, flags.FlgHTTPAddress)
		if err != nil {
			return nil, err
		}

		ps := http01.NewProviderServerWithOptions(http01.Options{
			Network:         getNetworkStack(cmd).Network("tcp"),
			Address:         net.JoinHostPort(host, port),
			ProxyHeaderName: cmd.String(flags.FlgHTTPProxyHeader),
		})

		return ps, nil

	default:
		ps := http01.NewProviderServerWithOptions(http01.Options{
			Network:         getNetworkStack(cmd).Network("tcp"),
			Address:         net.JoinHostPort("", ":80"),
			ProxyHeaderName: cmd.String(flags.FlgHTTPProxyHeader),
		})

		return ps, nil
	}
}

func setupTLSProvider(cmd *cli.Command, client *lego.Client) error {
	options := tlsalpn01.Options{
		Network: getNetworkStack(cmd).Network("tcp"),
	}

	if cmd.IsSet(flags.FlgTLSAddress) {
		host, port, err := parseAddress(cmd, flags.FlgTLSAddress)
		if err != nil {
			return err
		}

		options.Host = host
		options.Port = port
	}

	return client.Challenge.SetTLSALPN01Provider(
		tlsalpn01.NewProviderServerWithOptions(options),
		tlsalpn01.SetDelay(cmd.Duration(flags.FlgTLSDelay)),
	)
}

func setupDNS(cmd *cli.Command, client *lego.Client) error {
	provider, err := dns.NewDNSChallengeProviderByName(cmd.String(flags.FlgDNS))
	if err != nil {
		return err
	}

	opts := &dns01.Options{RecursiveNameservers: cmd.StringSlice(flags.FlgDNSResolvers)}

	if cmd.IsSet(flags.FlgDNSTimeout) {
		opts.Timeout = time.Duration(cmd.Int(flags.FlgDNSTimeout)) * time.Second
	}

	opts.NetworkStack = getNetworkStack(cmd)

	dns01.SetDefaultClient(dns01.NewClient(opts))

	shouldWait := cmd.IsSet(flags.FlgDNSPropagationWait)

	return client.Challenge.SetDNS01Provider(provider,
		dns01.CondOptions(shouldWait,
			dns01.PropagationWait(cmd.Duration(flags.FlgDNSPropagationWait), true),
		),
		dns01.CondOptions(!shouldWait,
			dns01.CondOptions(cmd.Bool(flags.FlgDNSPropagationDisableANS),
				dns01.DisableAuthoritativeNssPropagationRequirement(),
			),
			dns01.CondOptions(cmd.Bool(flags.FlgDNSPropagationDisableRNS),
				dns01.DisableRecursiveNSsPropagationRequirement(),
			),
		),
	)
}

func setupDNSPersist(cmd *cli.Command, client *lego.Client) error {
	opts := &dnspersist01.Options{RecursiveNameservers: cmd.StringSlice(flags.FlgDNSPersistResolvers)}

	if cmd.IsSet(flags.FlgDNSPersistTimeout) {
		opts.Timeout = time.Duration(cmd.Int(flags.FlgDNSPersistTimeout)) * time.Second
	}

	opts.NetworkStack = getNetworkStack(cmd)

	dnspersist01.SetDefaultClient(dnspersist01.NewClient(opts))

	shouldWait := cmd.IsSet(flags.FlgDNSPersistPropagationWait)

	return client.Challenge.SetDNSPersist01(
		dnspersist01.WithIssuerDomainName(cmd.String(flags.FlgDNSPersistIssuerDomainName)),
		dnspersist01.CondOptions(cmd.IsSet(flags.FlgDNSPersistPersistUntil),
			dnspersist01.WithPersistUntil(cmd.Timestamp(flags.FlgDNSPersistPersistUntil)),
		),
		dnspersist01.CondOptions(shouldWait,
			dnspersist01.PropagationWait(cmd.Duration(flags.FlgDNSPersistPropagationWait), true),
		),
		dnspersist01.CondOptions(!shouldWait,
			dnspersist01.CondOptions(cmd.Bool(flags.FlgDNSPersistPropagationDisableANS),
				dnspersist01.DisableAuthoritativeNssPropagationRequirement(),
			),
			dnspersist01.CondOptions(cmd.Bool(flags.FlgDNSPersistPropagationDisableRNS),
				dnspersist01.DisableRecursiveNSsPropagationRequirement(),
			),
		),
	)
}

func getNetworkStack(cmd *cli.Command) challenge.NetworkStack {
	switch {
	case cmd.Bool(flags.FlgIPv4Only):
		return challenge.IPv4Only

	case cmd.Bool(flags.FlgIPv6Only):
		return challenge.IPv6Only

	default:
		return challenge.DualStack
	}
}
