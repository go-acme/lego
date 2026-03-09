package cmd

import (
	"log/slog"
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
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/providers/http/memcached"
	"github.com/go-acme/lego/v5/providers/http/s3"
	"github.com/go-acme/lego/v5/providers/http/webroot"
	"github.com/urfave/cli/v3"
)

func setupChallenges(cmd *cli.Command, client *lego.Client) {
	if cmd.Bool(flags.FlgHTTP) {
		err := client.Challenge.SetHTTP01Provider(setupHTTPProvider(cmd), http01.SetDelay(cmd.Duration(flags.FlgHTTPDelay)))
		if err != nil {
			log.Fatal("Could not set HTTP challenge provider.", log.ErrorAttr(err))
		}
	}

	if cmd.Bool(flags.FlgTLS) {
		err := client.Challenge.SetTLSALPN01Provider(setupTLSProvider(cmd), tlsalpn01.SetDelay(cmd.Duration(flags.FlgTLSDelay)))
		if err != nil {
			log.Fatal("Could not set TLS challenge provider.", log.ErrorAttr(err))
		}
	}

	if cmd.IsSet(flags.FlgDNS) {
		err := setupDNS(cmd, client)
		if err != nil {
			log.Fatal("Could not set DNS challenge provider.", log.ErrorAttr(err))
		}
	}

	if cmd.Bool(flags.FlgDNSPersist) {
		err := setupDNSPersist(cmd, client)
		if err != nil {
			log.Fatal("Could not set DNS-PERSIST-01 challenge provider.", log.ErrorAttr(err))
		}
	}
}

func setupHTTPProvider(cmd *cli.Command) challenge.Provider {
	switch {
	case cmd.IsSet(flags.FlgHTTPWebroot):
		ps, err := webroot.NewHTTPProvider(cmd.String(flags.FlgHTTPWebroot))
		if err != nil {
			log.Fatal("Could not create the webroot provider.",
				slog.String("flag", flags.FlgHTTPWebroot),
				slog.String("webRoot", cmd.String(flags.FlgHTTPWebroot)),
				log.ErrorAttr(err),
			)
		}

		return ps

	case cmd.IsSet(flags.FlgHTTPMemcachedHost):
		ps, err := memcached.NewMemcachedProvider(cmd.StringSlice(flags.FlgHTTPMemcachedHost))
		if err != nil {
			log.Fatal("Could not create the memcached provider.",
				slog.String("flag", flags.FlgHTTPMemcachedHost),
				slog.String("memcachedHosts", strings.Join(cmd.StringSlice(flags.FlgHTTPMemcachedHost), ", ")),
				log.ErrorAttr(err),
			)
		}

		return ps

	case cmd.IsSet(flags.FlgHTTPS3Bucket):
		ps, err := s3.NewHTTPProvider(cmd.String(flags.FlgHTTPS3Bucket))
		if err != nil {
			log.Fatal("Could not create the S3 provider.",
				slog.String("flag", flags.FlgHTTPS3Bucket),
				slog.String("bucket", cmd.String(flags.FlgHTTPS3Bucket)),
				log.ErrorAttr(err),
			)
		}

		return ps

	case cmd.IsSet(flags.FlgHTTPPort):
		host, port, err := parseAddress(cmd, flags.FlgHTTPPort)
		if err != nil {
			log.Fatal("Invalid address.", log.ErrorAttr(err))
		}

		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
			Address: net.JoinHostPort(host, port),
		})

		if header := cmd.String(flags.FlgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}

		return srv

	case cmd.Bool(flags.FlgHTTP):
		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
			Address: net.JoinHostPort("", ":80"),
		})

		if header := cmd.String(flags.FlgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}

		return srv

	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
}

func setupTLSProvider(cmd *cli.Command) challenge.Provider {
	switch {
	case cmd.IsSet(flags.FlgTLSPort):
		host, port, err := parseAddress(cmd, flags.FlgTLSPort)
		if err != nil {
			log.Fatal("Invalid address.", log.ErrorAttr(err))
		}

		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
			Host:    host,
			Port:    port,
		})

	case cmd.Bool(flags.FlgTLS):
		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
		})

	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
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

	err = client.Challenge.SetDNS01Provider(provider,
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

	return err
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
