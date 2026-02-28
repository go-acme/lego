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
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/providers/http/memcached"
	"github.com/go-acme/lego/v5/providers/http/s3"
	"github.com/go-acme/lego/v5/providers/http/webroot"
	"github.com/urfave/cli/v3"
)

func setupChallenges(cmd *cli.Command, client *lego.Client) {
	if cmd.Bool(flgHTTP) {
		err := client.Challenge.SetHTTP01Provider(setupHTTPProvider(cmd), http01.SetDelay(cmd.Duration(flgHTTPDelay)))
		if err != nil {
			log.Fatal("Could not set HTTP challenge provider.", log.ErrorAttr(err))
		}
	}

	if cmd.Bool(flgTLS) {
		err := client.Challenge.SetTLSALPN01Provider(setupTLSProvider(cmd), tlsalpn01.SetDelay(cmd.Duration(flgTLSDelay)))
		if err != nil {
			log.Fatal("Could not set TLS challenge provider.", log.ErrorAttr(err))
		}
	}

	if cmd.IsSet(flgDNS) {
		err := setupDNS(cmd, client)
		if err != nil {
			log.Fatal("Could not set DNS challenge provider.", log.ErrorAttr(err))
		}
	}

	if cmd.Bool(flgDNSPersist) {
		err := setupDNSPersist(cmd, client)
		if err != nil {
			log.Fatal("Could not set DNS-PERSIST-01 challenge provider.", log.ErrorAttr(err))
		}
	}
}

func setupHTTPProvider(cmd *cli.Command) challenge.Provider {
	switch {
	case cmd.IsSet(flgHTTPWebroot):
		ps, err := webroot.NewHTTPProvider(cmd.String(flgHTTPWebroot))
		if err != nil {
			log.Fatal("Could not create the webroot provider.",
				slog.String("flag", flgHTTPWebroot),
				slog.String("webRoot", cmd.String(flgHTTPWebroot)),
				log.ErrorAttr(err),
			)
		}

		return ps

	case cmd.IsSet(flgHTTPMemcachedHost):
		ps, err := memcached.NewMemcachedProvider(cmd.StringSlice(flgHTTPMemcachedHost))
		if err != nil {
			log.Fatal("Could not create the memcached provider.",
				slog.String("flag", flgHTTPMemcachedHost),
				slog.String("memcachedHosts", strings.Join(cmd.StringSlice(flgHTTPMemcachedHost), ", ")),
				log.ErrorAttr(err),
			)
		}

		return ps

	case cmd.IsSet(flgHTTPS3Bucket):
		ps, err := s3.NewHTTPProvider(cmd.String(flgHTTPS3Bucket))
		if err != nil {
			log.Fatal("Could not create the S3 provider.",
				slog.String("flag", flgHTTPS3Bucket),
				slog.String("bucket", cmd.String(flgHTTPS3Bucket)),
				log.ErrorAttr(err),
			)
		}

		return ps

	case cmd.IsSet(flgHTTPPort):
		host, port, err := parseAddress(cmd, flgHTTPPort)
		if err != nil {
			log.Fatal("Invalid address.", log.ErrorAttr(err))
		}

		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
			Address: net.JoinHostPort(host, port),
		})

		if header := cmd.String(flgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}

		return srv

	case cmd.Bool(flgHTTP):
		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
			Address: net.JoinHostPort("", ":80"),
		})

		if header := cmd.String(flgHTTPProxyHeader); header != "" {
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
	case cmd.IsSet(flgTLSPort):
		host, port, err := parseAddress(cmd, flgTLSPort)
		if err != nil {
			log.Fatal("Invalid address.", log.ErrorAttr(err))
		}

		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
			Host:    host,
			Port:    port,
		})

	case cmd.Bool(flgTLS):
		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			Network: getNetworkStack(cmd).Network("tcp"),
		})

	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
}

func setupDNS(cmd *cli.Command, client *lego.Client) error {
	provider, err := dns.NewDNSChallengeProviderByName(cmd.String(flgDNS))
	if err != nil {
		return err
	}

	opts := &dns01.Options{RecursiveNameservers: cmd.StringSlice(flgDNSResolvers)}

	if cmd.IsSet(flgDNSTimeout) {
		opts.Timeout = time.Duration(cmd.Int(flgDNSTimeout)) * time.Second
	}

	opts.NetworkStack = getNetworkStack(cmd)

	dns01.SetDefaultClient(dns01.NewClient(opts))

	shouldWait := cmd.IsSet(flgDNSPropagationWait)

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOptions(shouldWait,
			dns01.PropagationWait(cmd.Duration(flgDNSPropagationWait), true),
		),
		dns01.CondOptions(!shouldWait,
			dns01.CondOptions(cmd.Bool(flgDNSPropagationDisableANS),
				dns01.DisableAuthoritativeNssPropagationRequirement(),
			),
			dns01.CondOptions(cmd.Bool(flgDNSPropagationDisableRNS),
				dns01.DisableRecursiveNSsPropagationRequirement(),
			),
		),
	)

	return err
}

func setupDNSPersist(cmd *cli.Command, client *lego.Client) error {
	opts := &dnspersist01.Options{RecursiveNameservers: cmd.StringSlice(flgDNSPersistResolvers)}

	if cmd.IsSet(flgDNSPersistTimeout) {
		opts.Timeout = time.Duration(cmd.Int(flgDNSPersistTimeout)) * time.Second
	}

	opts.NetworkStack = getNetworkStack(cmd)

	dnspersist01.SetDefaultClient(dnspersist01.NewClient(opts))

	shouldWait := cmd.IsSet(flgDNSPersistPropagationWait)

	return client.Challenge.SetDNSPersist01(
		dnspersist01.WithIssuerDomainName(cmd.String(flgDNSPersistIssuerDomainName)),
		dnspersist01.CondOptions(cmd.IsSet(flgDNSPersistPersistUntil),
			dnspersist01.WithPersistUntil(cmd.Timestamp(flgDNSPersistPersistUntil)),
		),
		dnspersist01.CondOptions(shouldWait,
			dnspersist01.PropagationWait(cmd.Duration(flgDNSPersistPropagationWait), true),
		),
		dnspersist01.CondOptions(!shouldWait,
			dnspersist01.CondOptions(cmd.Bool(flgDNSPersistPropagationDisableANS),
				dnspersist01.DisableAuthoritativeNssPropagationRequirement(),
			),
			dnspersist01.CondOptions(cmd.Bool(flgDNSSPersistPropagationDisableRNS),
				dnspersist01.DisableRecursiveNSsPropagationRequirement(),
			),
		),
	)
}

func getNetworkStack(cmd *cli.Command) challenge.NetworkStack {
	switch {
	case cmd.Bool(flgIPv4Only):
		return challenge.IPv4Only

	case cmd.Bool(flgIPv6Only):
		return challenge.IPv6Only

	default:
		return challenge.DualStack
	}
}
