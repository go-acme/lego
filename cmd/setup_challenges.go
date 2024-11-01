package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/providers/http/memcached"
	"github.com/go-acme/lego/v4/providers/http/s3"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/urfave/cli/v2"
)

func setupChallenges(ctx *cli.Context, client *lego.Client) {
	if !ctx.Bool(flgHTTP) && !ctx.Bool(flgTLS) && !ctx.IsSet(flgDNS) {
		log.Fatalf("No challenge selected. You must specify at least one challenge: `--%s`, `--%s`, `--%s`.", flgHTTP, flgTLS, flgDNS)
	}

	if ctx.Bool(flgHTTP) {
		err := client.Challenge.SetHTTP01Provider(setupHTTPProvider(ctx))
		if err != nil {
			log.Fatal(err)
		}
	}

	if ctx.Bool(flgTLS) {
		err := client.Challenge.SetTLSALPN01Provider(setupTLSProvider(ctx))
		if err != nil {
			log.Fatal(err)
		}
	}

	if ctx.IsSet(flgDNS) {
		err := setupDNS(ctx, client)
		if err != nil {
			log.Fatal(err)
		}
	}
}

//nolint:gocyclo // the complexity is expected.
func setupHTTPProvider(ctx *cli.Context) challenge.Provider {
	switch {
	case ctx.IsSet(flgHTTPWebroot):
		ps, err := webroot.NewHTTPProvider(ctx.String(flgHTTPWebroot))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case ctx.IsSet(flgHTTPMemcachedHost):
		ps, err := memcached.NewMemcachedProvider(ctx.StringSlice(flgHTTPMemcachedHost))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case ctx.IsSet(flgHTTPS3Bucket):
		ps, err := s3.NewHTTPProvider(ctx.String(flgHTTPS3Bucket))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case ctx.IsSet(flgHTTPPort):
		iface := ctx.String(flgHTTPPort)
		if !strings.Contains(iface, ":") {
			log.Fatalf("The --%s switch only accepts interface:port or :port for its argument.", flgHTTPPort)
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal(err)
		}

		srv := http01.NewProviderServer(host, port)
		if header := ctx.String(flgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}
		return srv
	case ctx.Bool(flgHTTP):
		srv := http01.NewProviderServer("", "")
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
			log.Fatalf("The --%s switch only accepts interface:port or :port for its argument.", flgTLSPort)
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal(err)
		}

		return tlsalpn01.NewProviderServer(host, port)
	case ctx.Bool(flgTLS):
		return tlsalpn01.NewProviderServer("", "")
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

	servers := ctx.StringSlice(flgDNSResolvers)

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(len(servers) > 0,
			dns01.AddRecursiveNameservers(dns01.ParseNameservers(ctx.StringSlice(flgDNSResolvers)))),

		dns01.CondOption(ctx.Bool(flgDNSDisableCP) || ctx.Bool(flgDNSPropagationDisableANS),
			dns01.DisableAuthoritativeNssPropagationRequirement()),

		dns01.CondOption(ctx.Duration(flgDNSPropagationWait) > 0,
			// TODO(ldez): inside the next major version we will use flgDNSDisableCP here.
			// This will change the meaning of this flag to really disable all propagation checks.
			dns01.PropagationWait(wait, true)),

		dns01.CondOption(ctx.Bool(flgDNSPropagationRNS),
			dns01.RecursiveNSsPropagationRequirement()),

		dns01.CondOption(ctx.IsSet(flgDNSTimeout),
			dns01.AddDNSTimeout(time.Duration(ctx.Int(flgDNSTimeout))*time.Second)),
	)

	return err
}

func checkPropagationExclusiveOptions(ctx *cli.Context) error {
	if ctx.IsSet(flgDNSDisableCP) {
		log.Printf("The flag '%s' is deprecated use '%s' instead.", flgDNSDisableCP, flgDNSPropagationDisableANS)
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
