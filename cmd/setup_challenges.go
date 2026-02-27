package cmd

import (
	"fmt"
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

//nolint:gocyclo // challenge setup dispatch is expected to branch by enabled challenge type.
func setupChallenges(cmd *cli.Command, client *lego.Client) {
	if !cmd.Bool(flgHTTP) && !cmd.Bool(flgTLS) && !cmd.IsSet(flgDNS) && !cmd.Bool(flgDNSPersist) {
		log.Fatal(fmt.Sprintf("No challenge selected. You must specify at least one challenge: `--%s`, `--%s`, `--%s`, `--%s`.",
			flgHTTP, flgTLS, flgDNS, flgDNSPersist))
	}

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

//nolint:gocyclo // the complexity is expected.
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
		iface := cmd.String(flgHTTPPort)

		if !strings.Contains(iface, ":") {
			log.Fatal(
				fmt.Sprintf("The --%s switch only accepts interface:port or :port for its argument.", flgHTTPPort),
				slog.String("flag", flgHTTPPort),
				slog.String("port", cmd.String(flgHTTPPort)),
			)
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal("Could not split host and port.", slog.String("iface", iface), log.ErrorAttr(err))
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
		iface := cmd.String(flgTLSPort)
		if !strings.Contains(iface, ":") {
			log.Fatal(fmt.Sprintf("The --%s switch only accepts interface:port or :port for its argument.", flgTLSPort))
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal("Could not split host and port.", slog.String("iface", iface), log.ErrorAttr(err))
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
	err := validatePropagationExclusiveOptions(cmd, flgDNSPropagationWait, flgDNSPropagationDisableANS, flgDNSPropagationDisableRNS)
	if err != nil {
		return err
	}

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
	err := validatePropagationExclusiveOptions(cmd, flgDNSPersistPropagationWait, flgDNSPersistPropagationDisableANS, flgDNSPersistIssuerDomainName)
	if err != nil {
		return err
	}

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

func validatePropagationExclusiveOptions(cmd *cli.Command, flgWait, flgANS, flgDNS string) error {
	if !cmd.IsSet(flgWait) {
		return nil
	}

	if isSetBool(cmd, flgANS) {
		return fmt.Errorf("'%s' and '%s' are mutually exclusive", flgWait, flgANS)
	}

	if isSetBool(cmd, flgDNS) {
		return fmt.Errorf("'%s' and '%s' are mutually exclusive", flgWait, flgDNS)
	}

	return nil
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

func isSetBool(cmd *cli.Command, name string) bool {
	return cmd.IsSet(name) && cmd.Bool(name)
}
