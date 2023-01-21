package cmd

import (
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
		setupDNS(ctx, client)
	}
}

func resolveHTTP01Network(ctx *cli.Context) http01.ProviderNetwork {
	var network http01.ProviderNetwork
	switch {
	case ctx.IsSet("ipv4only") && ctx.IsSet("ipv6only"):
		network = http01.DefaultNetwork
	case ctx.IsSet("ipv4only"):
		network = http01.TCP4Network
	case ctx.IsSet("ipv6only"):
		network = http01.TCP6Network
	}
	return network
}

func setupHTTPProvider(ctx *cli.Context) challenge.Provider {
	network := resolveHTTP01Network(ctx)

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
		srv.SetNetwork(network)
		if header := ctx.String("http.proxy-header"); header != "" {
			srv.SetProxyHeader(header)
		}
		return srv
	case ctx.Bool("http"):
		srv := http01.NewProviderServer("", "")
		srv.SetNetwork(network)
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
	var network tlsalpn01.ProviderNetwork
	switch {
	case ctx.IsSet("ipv4only") && ctx.IsSet("ipv6only"):
		network = tlsalpn01.DefaultNetwork
	case ctx.IsSet("ipv4only"):
		network = tlsalpn01.TCP4Network
	case ctx.IsSet("ipv6only"):
		network = tlsalpn01.TCP6Network
	}

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
		srv.SetNetwork(network)
		return srv
	case ctx.Bool("tls"):
		srv := tlsalpn01.NewProviderServer("", "")
		srv.SetNetwork(network)
		return srv
	default:
		log.Fatal("Invalid HTTP challenge options.")
		return nil
	}
}

func setupDNS(ctx *cli.Context, client *lego.Client) {
	provider, err := dns.NewDNSChallengeProviderByName(ctx.String("dns"))
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case ctx.IsSet("ipv4only") && ctx.IsSet("ipv6only"):
		// If both flags are set then it's as good as not providing either flag,
		// so we default to the OS choice.
		dns01.SetNetworkStack(dns01.DefaultNetworkStack)

	case ctx.IsSet("ipv4only"):
		dns01.SetNetworkStack(dns01.IPv4Only)

	case ctx.IsSet("ipv6only"):
		dns01.SetNetworkStack(dns01.IPv6Only)
	}

	servers := ctx.StringSlice("dns.resolvers")
	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(len(servers) > 0,
			dns01.AddRecursiveNameservers(dns01.ParseNameservers(ctx.StringSlice("dns.resolvers")))),
		dns01.CondOption(ctx.Bool("dns.disable-cp"),
			dns01.DisableCompletePropagationRequirement()),
		dns01.CondOption(ctx.IsSet("dns-timeout"),
			dns01.AddDNSTimeout(time.Duration(ctx.Int("dns-timeout"))*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}
}
