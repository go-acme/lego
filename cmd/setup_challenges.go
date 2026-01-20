package cmd

import (
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
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
	if !cmd.Bool(flgHTTP) && !cmd.Bool(flgTLS) && !cmd.IsSet(flgDNS) {
		log.Fatal(fmt.Sprintf("No challenge selected. You must specify at least one challenge: `--%s`, `--%s`, `--%s`.", flgHTTP, flgTLS, flgDNS))
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
			// TODO(ldez): set network stack
			Network: "tcp",
			Address: net.JoinHostPort(host, port),
		})

		if header := cmd.String(flgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}

		return srv

	case cmd.Bool(flgHTTP):
		srv := http01.NewProviderServerWithOptions(http01.Options{
			// TODO(ldez): set network stack
			Network: "tcp",
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
			// TODO(ldez): set network stack
			Network: "tcp",
			Host:    host,
			Port:    port,
		})

	case cmd.Bool(flgTLS):
		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			// TODO(ldez): set network stack
			Network: "tcp",
		})

	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
}

func setupDNS(cmd *cli.Command, client *lego.Client) error {
	err := checkPropagationExclusiveOptions(cmd)
	if err != nil {
		return err
	}

	wait := cmd.Duration(flgDNSPropagationWait)
	if wait < 0 {
		return fmt.Errorf("'%s' cannot be negative", flgDNSPropagationWait)
	}

	provider, err := dns.NewDNSChallengeProviderByName(cmd.String(flgDNS))
	if err != nil {
		return err
	}

	opts := &dns01.Options{RecursiveNameservers: cmd.StringSlice(flgDNSResolvers)}

	if cmd.IsSet(flgDNSTimeout) {
		opts.Timeout = time.Duration(cmd.Int(flgDNSTimeout)) * time.Second
	}

	dns01.SetDefaultClient(dns01.NewClient(opts))

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(cmd.Bool(flgDNSDisableCP) || cmd.Bool(flgDNSPropagationDisableANS),
			dns01.DisableAuthoritativeNssPropagationRequirement()),

		dns01.CondOption(cmd.Duration(flgDNSPropagationWait) > 0,
			// TODO(ldez): inside the next major version we will use flgDNSDisableCP here.
			// This will change the meaning of this flag to really disable all propagation checks.
			dns01.PropagationWait(wait, true)),

		dns01.CondOption(cmd.Bool(flgDNSPropagationRNS),
			dns01.RecursiveNSsPropagationRequirement()),
	)

	return err
}

func checkPropagationExclusiveOptions(cmd *cli.Command) error {
	if cmd.IsSet(flgDNSDisableCP) {
		log.Warnf(log.LazySprintf("The flag '%s' is deprecated use '%s' instead.", flgDNSDisableCP, flgDNSPropagationDisableANS))
	}

	if (isSetBool(cmd, flgDNSDisableCP) || isSetBool(cmd, flgDNSPropagationDisableANS)) && cmd.IsSet(flgDNSPropagationWait) {
		return fmt.Errorf("'%s' and '%s' are mutually exclusive", flgDNSPropagationDisableANS, flgDNSPropagationWait)
	}

	if isSetBool(cmd, flgDNSPropagationRNS) && cmd.IsSet(flgDNSPropagationWait) {
		return fmt.Errorf("'%s' and '%s' are mutually exclusive", flgDNSPropagationRNS, flgDNSPropagationWait)
	}

	return nil
}

func isSetBool(cmd *cli.Command, name string) bool {
	return cmd.IsSet(name) && cmd.Bool(name)
}
