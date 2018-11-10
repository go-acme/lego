package cmd

import (
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/providers/dns"
	"github.com/xenolf/lego/providers/http/memcached"
	"github.com/xenolf/lego/providers/http/webroot"
)

func setupChallenges(client *acme.Client, c *cli.Context, conf *Configuration) {
	if len(c.GlobalStringSlice("exclude")) > 0 {
		client.Challenge.Exclude(conf.ExcludedSolvers())
	}

	if c.GlobalIsSet("webroot") {
		setupWebroot(client, c.GlobalString("webroot"))
	}

	if c.GlobalIsSet("memcached-host") {
		setupMemcached(client, c.GlobalStringSlice("memcached-host"))
	}

	if c.GlobalIsSet("http") {
		setupHTTP(client, c.GlobalString("http"))
	}

	if c.GlobalIsSet("tls") {
		setupTLS(client, c.GlobalString("tls"))
	}

	if c.GlobalIsSet("dns") {
		setupDNS(client, c)
	}
}

func setupWebroot(client *acme.Client, path string) {
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

func setupMemcached(client *acme.Client, hosts []string) {
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

func setupHTTP(client *acme.Client, iface string) {
	if !strings.Contains(iface, ":") {
		log.Fatalf("The --http switch only accepts interface:port or :port for its argument.")
	}

	err := client.Challenge.SetHTTP01Address(iface)
	if err != nil {
		log.Fatal(err)
	}
}

func setupTLS(client *acme.Client, iface string) {
	if !strings.Contains(iface, ":") {
		log.Fatalf("The --tls switch only accepts interface:port or :port for its argument.")
	}

	err := client.Challenge.SetTLSALPN01Address(iface)
	if err != nil {
		log.Fatal(err)
	}
}

func setupDNS(client *acme.Client, c *cli.Context) {
	provider, err := dns.NewDNSChallengeProviderByName(c.GlobalString("dns"))
	if err != nil {
		log.Fatal(err)
	}

	servers := c.GlobalStringSlice("dns-resolvers")
	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(len(servers) > 0,
			dns01.AddRecursiveNameservers(dns01.ParseNameservers(c.GlobalStringSlice("dns-resolvers")))),
		dns01.CondOption(c.GlobalIsSet("dns-disable-cp"),
			dns01.DisableCompletePropagationRequirement()),
		dns01.CondOption(c.GlobalIsSet("dns-timeout"),
			dns01.AddDNSTimeout(time.Duration(c.GlobalInt("dns-timeout"))*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}
}
