package cmd

import (
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/lego"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/providers/dns"
	"github.com/xenolf/lego/providers/http/memcached"
	"github.com/xenolf/lego/providers/http/webroot"
)

func setupChallenges(ctx *cli.Context, client *lego.Client) {
	if len(ctx.GlobalStringSlice("exclude")) > 0 {
		excludedSolvers(ctx, client)
	}

	if ctx.GlobalIsSet("webroot") {
		setupWebroot(client, ctx.GlobalString("webroot"))
	}

	if ctx.GlobalIsSet("memcached-host") {
		setupMemcached(client, ctx.GlobalStringSlice("memcached-host"))
	}

	if ctx.GlobalIsSet("http") {
		setupHTTP(client, ctx.GlobalString("http"))
	}

	if ctx.GlobalIsSet("tls") {
		setupTLS(client, ctx.GlobalString("tls"))
	}

	if ctx.GlobalIsSet("dns") {
		setupDNS(ctx, client)
	}
}

func excludedSolvers(ctx *cli.Context, client *lego.Client) {
	var cc []challenge.Type
	for _, s := range ctx.GlobalStringSlice("exclude") {
		cc = append(cc, challenge.Type(s))
	}
	client.Challenge.Exclude(cc)
}

func setupWebroot(client *lego.Client, path string) {
	provider, err := webroot.NewHTTPProvider(path)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Challenge.SetHTTP01Provider(provider)
	if err != nil {
		log.Fatal(err)
	}

	// --webroot=foo indicates that the user specifically want to do a HTTP challenge
	// infer that the user also wants to exclude all other challenges
	client.Challenge.Exclude([]challenge.Type{challenge.DNS01, challenge.TLSALPN01})
}

func setupMemcached(client *lego.Client, hosts []string) {
	provider, err := memcached.NewMemcachedProvider(hosts)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Challenge.SetHTTP01Provider(provider)
	if err != nil {
		log.Fatal(err)
	}

	// --memcached-host=foo:11211 indicates that the user specifically want to do a HTTP challenge
	// infer that the user also wants to exclude all other challenges
	client.Challenge.Exclude([]challenge.Type{challenge.DNS01, challenge.TLSALPN01})
}

func setupHTTP(client *lego.Client, iface string) {
	if !strings.Contains(iface, ":") {
		log.Fatalf("The --http switch only accepts interface:port or :port for its argument.")
	}

	err := client.Challenge.SetHTTP01Address(iface)
	if err != nil {
		log.Fatal(err)
	}
}

func setupTLS(client *lego.Client, iface string) {
	if !strings.Contains(iface, ":") {
		log.Fatalf("The --tls switch only accepts interface:port or :port for its argument.")
	}

	err := client.Challenge.SetTLSALPN01Address(iface)
	if err != nil {
		log.Fatal(err)
	}
}

func setupDNS(ctx *cli.Context, client *lego.Client) {
	provider, err := dns.NewDNSChallengeProviderByName(ctx.GlobalString("dns"))
	if err != nil {
		log.Fatal(err)
	}

	servers := ctx.GlobalStringSlice("dns-resolvers")
	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(len(servers) > 0,
			dns01.AddRecursiveNameservers(dns01.ParseNameservers(ctx.GlobalStringSlice("dns-resolvers")))),
		dns01.CondOption(ctx.GlobalIsSet("dns-disable-cp"),
			dns01.DisableCompletePropagationRequirement()),
		dns01.CondOption(ctx.GlobalIsSet("dns-timeout"),
			dns01.AddDNSTimeout(time.Duration(ctx.GlobalInt("dns-timeout"))*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}
}
