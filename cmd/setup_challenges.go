package cmd

import (
	"errors"
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
	if !ctx.Bool("http") && !ctx.Bool("tls") && !ctx.IsSet("dns") {
		log.Fatal("No challenge selected. You must specify at least one challenge: `--http`, `--tls`, `--dns`.")
	}

	if ctx.Bool("http") {
		err := client.Challenge.SetHTTP01Provider(setupHTTPProvider(ctx))
		if err != nil {
			log.Fatal(err)
		}
	}

	if ctx.Bool("tls") {
		err := client.Challenge.SetTLSALPN01Provider(setupTLSProvider(ctx))
		if err != nil {
			log.Fatal(err)
		}
	}

	if ctx.IsSet("dns") {
		err := setupDNS(ctx, client)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type networkStackSetter interface {
	SetIPv4Only()
	SetIPv6Only()
	SetDualStack()
}

func setNetwork(ctx *cli.Context, srv networkStackSetter) {
	switch v4, v6 := ctx.IsSet("ipv4only"), ctx.IsSet("ipv6only"); {
	case v4 && !v6:
		srv.SetIPv4Only()
	case !v4 && v6:
		srv.SetIPv6Only()
	default:
		// setting both --ipv4only and --ipv6only is not an error, just a no-op
		srv.SetDualStack()
	}
}

//nolint:gocyclo // the complexity is expected.
func setupHTTPProvider(ctx *cli.Context) challenge.Provider {
	switch {
	case ctx.IsSet("http.webroot"):
		ps, err := webroot.NewHTTPProvider(ctx.String("http.webroot"))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case ctx.IsSet("http.memcached-host"):
		ps, err := memcached.NewMemcachedProvider(ctx.StringSlice("http.memcached-host"))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case ctx.IsSet("http.s3-bucket"):
		ps, err := s3.NewHTTPProvider(ctx.String("http.s3-bucket"))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case ctx.IsSet("http.port"):
		iface := ctx.String("http.port")
		if !strings.Contains(iface, ":") {
			log.Fatalf("The --http switch only accepts interface:port or :port for its argument.")
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal(err)
		}

		srv := http01.NewProviderServer(host, port)
		setNetwork(ctx, srv)
		if header := ctx.String("http.proxy-header"); header != "" {
			srv.SetProxyHeader(header)
		}
		return srv
	case ctx.Bool("http"):
		srv := http01.NewProviderServer("", "")
		setNetwork(ctx, srv)
		if header := ctx.String("http.proxy-header"); header != "" {
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
	case ctx.IsSet("tls.port"):
		iface := ctx.String("tls.port")
		if !strings.Contains(iface, ":") {
			log.Fatalf("The --tls switch only accepts interface:port or :port for its argument.")
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal(err)
		}

		srv := tlsalpn01.NewProviderServer(host, port)
		setNetwork(ctx, srv)
		return srv
	case ctx.Bool("tls"):
		srv := tlsalpn01.NewProviderServer("", "")
		setNetwork(ctx, srv)
		return srv
	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
}

func setupDNS(ctx *cli.Context, client *lego.Client) error {
	if ctx.IsSet("dns.disable-cp") && ctx.Bool("dns.disable-cp") && ctx.IsSet("dns.propagation-wait") {
		return errors.New("'dns.disable-cp' and 'dns.propagation-wait' are mutually exclusive")
	}

	wait := ctx.Duration("dns.propagation-wait")
	if wait < 0 {
		return errors.New("'dns.propagation-wait' cannot be negative")
	}

	provider, err := dns.NewDNSChallengeProviderByName(ctx.String("dns"))
	if err != nil {
		return err
	}

	switch v4, v6 := ctx.IsSet("ipv4only"), ctx.IsSet("ipv6only"); {
	case v4 && !v6:
		dns01.SetIPv4Only()
	case !v4 && v6:
		dns01.SetIPv6Only()
	default:
		// setting both --ipv4only and --ipv6only is not an error, just a no-op
		dns01.SetDualStack()
	}

	servers := ctx.StringSlice("dns.resolvers")

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(len(servers) > 0,
			dns01.AddRecursiveNameservers(dns01.ParseNameservers(ctx.StringSlice("dns.resolvers")))),

		dns01.CondOption(ctx.Bool("dns.disable-cp"),
			dns01.DisableCompletePropagationRequirement()),

		dns01.CondOption(ctx.IsSet("dns.propagation-wait"), dns01.WrapPreCheck(
			func(domain, fqdn, value string, check dns01.PreCheckFunc) (bool, error) {
				time.Sleep(wait)
				return true, nil
			},
		)),

		dns01.CondOption(ctx.IsSet("dns-timeout"),
			dns01.AddDNSTimeout(time.Duration(ctx.Int("dns-timeout"))*time.Second)),
	)

	return err
}
