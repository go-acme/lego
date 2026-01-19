package cmd

import (
	"fmt"
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
	"github.com/urfave/cli/v2"
)

func setupChallenges(ctx *cli.Context, client *lego.Client) {
	if !ctx.Bool(flgHTTP) && !ctx.Bool(flgTLS) && !ctx.IsSet(flgDNS) {
		log.Fatal(fmt.Sprintf("No challenge selected. You must specify at least one challenge: `--%s`, `--%s`, `--%s`.", flgHTTP, flgTLS, flgDNS))
	}

	if ctx.Bool(flgHTTP) {
		err := client.Challenge.SetHTTP01Provider(setupHTTPProvider(ctx), http01.SetDelay(ctx.Duration(flgHTTPDelay)))
		if err != nil {
			log.Fatal("Could not set HTTP challenge provider.", "error", err)
		}
	}

	if ctx.Bool(flgTLS) {
		err := client.Challenge.SetTLSALPN01Provider(setupTLSProvider(ctx), tlsalpn01.SetDelay(ctx.Duration(flgTLSDelay)))
		if err != nil {
			log.Fatal("Could not set TLS challenge provider.", "error", err)
		}
	}

	if ctx.IsSet(flgDNS) {
		err := setupDNS(ctx, client)
		if err != nil {
			log.Fatal("Could not set DNS challenge provider.", "error", err)
		}
	}
}

//nolint:gocyclo // the complexity is expected.
func setupHTTPProvider(ctx *cli.Context) challenge.Provider {
	switch {
	case ctx.IsSet(flgHTTPWebroot):
		ps, err := webroot.NewHTTPProvider(ctx.String(flgHTTPWebroot))
		if err != nil {
			log.Fatal("Could not create the webroot provider.",
				"flag", flgHTTPWebroot, "webRoot", ctx.String(flgHTTPWebroot), "error", err)
		}

		return ps

	case ctx.IsSet(flgHTTPMemcachedHost):
		ps, err := memcached.NewMemcachedProvider(ctx.StringSlice(flgHTTPMemcachedHost))
		if err != nil {
			log.Fatal("Could not create the memcached provider.",
				"flag", flgHTTPMemcachedHost, "memcachedHosts", strings.Join(ctx.StringSlice(flgHTTPMemcachedHost), ", "), "error", err)
		}

		return ps

	case ctx.IsSet(flgHTTPS3Bucket):
		ps, err := s3.NewHTTPProvider(ctx.String(flgHTTPS3Bucket))
		if err != nil {
			log.Fatal("Could not create the S3 provider.",
				"flag", flgHTTPS3Bucket, "bucket", ctx.String(flgHTTPS3Bucket), "error", err)
		}

		return ps

	case ctx.IsSet(flgHTTPPort):
		iface := ctx.String(flgHTTPPort)

		if !strings.Contains(iface, ":") {
			log.Fatal(
				fmt.Sprintf("The --%s switch only accepts interface:port or :port for its argument.", flgHTTPPort),
				"flag", flgHTTPPort, "port", ctx.String(flgHTTPPort),
			)
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal("Could not split host and port.", "iface", iface, "error", err)
		}

		srv := http01.NewProviderServerWithOptions(http01.Options{
			// TODO(ldez): set network stack
			Network: "tcp",
			Address: net.JoinHostPort(host, port),
		})

		if header := ctx.String(flgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}

		return srv

	case ctx.Bool(flgHTTP):
		srv := http01.NewProviderServerWithOptions(http01.Options{
			// TODO(ldez): set network stack
			Network: "tcp",
			Address: net.JoinHostPort("", ":80"),
		})

		if header := ctx.String(flgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}

		return srv

	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
}

func setupTLSProvider(ctx *cli.Context) challenge.Provider {
	switch {
	case ctx.IsSet(flgTLSPort):
		iface := ctx.String(flgTLSPort)
		if !strings.Contains(iface, ":") {
			log.Fatal(fmt.Sprintf("The --%s switch only accepts interface:port or :port for its argument.", flgTLSPort))
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal("Could not split host and port.", "iface", iface, "error", err)
		}

		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			// TODO(ldez): set network stack
			Network: "tcp",
			Host:    host,
			Port:    port,
		})

	case ctx.Bool(flgTLS):
		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			// TODO(ldez): set network stack
			Network: "tcp",
		})

	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
}

func setupDNS(ctx *cli.Context, client *lego.Client) error {
	err := checkPropagationExclusiveOptions(ctx)
	if err != nil {
		return err
	}

	wait := ctx.Duration(flgDNSPropagationWait)
	if wait < 0 {
		return fmt.Errorf("'%s' cannot be negative", flgDNSPropagationWait)
	}

	provider, err := dns.NewDNSChallengeProviderByName(ctx.String(flgDNS))
	if err != nil {
		return err
	}

	opts := &dns01.Options{RecursiveNameservers: ctx.StringSlice(flgDNSResolvers)}

	if ctx.IsSet(flgDNSTimeout) {
		opts.Timeout = time.Duration(ctx.Int(flgDNSTimeout)) * time.Second
	}

	dns01.SetDefaultClient(dns01.NewClient(opts))

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(ctx.Bool(flgDNSDisableCP) || ctx.Bool(flgDNSPropagationDisableANS),
			dns01.DisableAuthoritativeNssPropagationRequirement()),

		dns01.CondOption(ctx.Duration(flgDNSPropagationWait) > 0,
			// TODO(ldez): inside the next major version we will use flgDNSDisableCP here.
			// This will change the meaning of this flag to really disable all propagation checks.
			dns01.PropagationWait(wait, true)),

		dns01.CondOption(ctx.Bool(flgDNSPropagationRNS),
			dns01.RecursiveNSsPropagationRequirement()),
	)

	return err
}

func checkPropagationExclusiveOptions(ctx *cli.Context) error {
	if ctx.IsSet(flgDNSDisableCP) {
		log.Warnf(log.LazySprintf("The flag '%s' is deprecated use '%s' instead.", flgDNSDisableCP, flgDNSPropagationDisableANS))
	}

	if (isSetBool(ctx, flgDNSDisableCP) || isSetBool(ctx, flgDNSPropagationDisableANS)) && ctx.IsSet(flgDNSPropagationWait) {
		return fmt.Errorf("'%s' and '%s' are mutually exclusive", flgDNSPropagationDisableANS, flgDNSPropagationWait)
	}

	if isSetBool(ctx, flgDNSPropagationRNS) && ctx.IsSet(flgDNSPropagationWait) {
		return fmt.Errorf("'%s' and '%s' are mutually exclusive", flgDNSPropagationRNS, flgDNSPropagationWait)
	}

	return nil
}

func isSetBool(ctx *cli.Context, name string) bool {
	return ctx.IsSet(name) && ctx.Bool(name)
}
