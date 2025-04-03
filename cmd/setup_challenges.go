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
	"github.com/urfave/cli/v3"
)

func setupChallenges(cmd *cli.Command, client *lego.Client) {
	if !cmd.Bool(flgHTTP) && !cmd.Bool(flgTLS) && !cmd.IsSet(flgDNS) {
		log.Fatalf("No challenge selected. You must specify at least one challenge: `--%s`, `--%s`, `--%s`.", flgHTTP, flgTLS, flgDNS)
	}

	if cmd.Bool(flgHTTP) {
		err := client.Challenge.SetHTTP01Provider(setupHTTPProvider(cmd), http01.SetDelay(cmd.Duration(flgHTTPDelay)))
		if err != nil {
			log.Fatal(err)
		}
	}

	if cmd.Bool(flgTLS) {
		err := client.Challenge.SetTLSALPN01Provider(setupTLSProvider(cmd), tlsalpn01.SetDelay(cmd.Duration(flgTLSDelay)))
		if err != nil {
			log.Fatal(err)
		}
	}

	if cmd.IsSet(flgDNS) {
		err := setupDNS(cmd, client)
		if err != nil {
			log.Fatal(err)
		}
	}
}

//nolint:gocyclo // the complexity is expected.
func setupHTTPProvider(cmd *cli.Command) challenge.Provider {
	switch {
	case cmd.IsSet(flgHTTPWebroot):
		ps, err := webroot.NewHTTPProvider(cmd.String(flgHTTPWebroot))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case cmd.IsSet(flgHTTPMemcachedHost):
		ps, err := memcached.NewMemcachedProvider(cmd.StringSlice(flgHTTPMemcachedHost))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case cmd.IsSet(flgHTTPS3Bucket):
		ps, err := s3.NewHTTPProvider(cmd.String(flgHTTPS3Bucket))
		if err != nil {
			log.Fatal(err)
		}
		return ps
	case cmd.IsSet(flgHTTPPort):
		iface := cmd.String(flgHTTPPort)
		if !strings.Contains(iface, ":") {
			log.Fatalf("The --%s switch only accepts interface:port or :port for its argument.", flgHTTPPort)
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal(err)
		}

		srv := http01.NewProviderServer(host, port)
		if header := cmd.String(flgHTTPProxyHeader); header != "" {
			srv.SetProxyHeader(header)
		}
		return srv
	case cmd.Bool(flgHTTP):
		srv := http01.NewProviderServer("", "")
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
			log.Fatalf("The --%s switch only accepts interface:port or :port for its argument.", flgTLSPort)
		}

		host, port, err := net.SplitHostPort(iface)
		if err != nil {
			log.Fatal(err)
		}

		return tlsalpn01.NewProviderServer(host, port)
	case cmd.Bool(flgTLS):
		return tlsalpn01.NewProviderServer("", "")
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

	servers := cmd.StringSlice(flgDNSResolvers)

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(len(servers) > 0,
			dns01.AddRecursiveNameservers(dns01.ParseNameservers(cmd.StringSlice(flgDNSResolvers)))),

		dns01.CondOption(cmd.Bool(flgDNSDisableCP) || cmd.Bool(flgDNSPropagationDisableANS),
			dns01.DisableAuthoritativeNssPropagationRequirement()),

		dns01.CondOption(cmd.Duration(flgDNSPropagationWait) > 0,
			// TODO(ldez): inside the next major version we will use flgDNSDisableCP here.
			// This will change the meaning of this flag to really disable all propagation checks.
			dns01.PropagationWait(wait, true)),

		dns01.CondOption(cmd.Bool(flgDNSPropagationRNS),
			dns01.RecursiveNSsPropagationRequirement()),

		dns01.CondOption(cmd.IsSet(flgDNSTimeout),
			dns01.AddDNSTimeout(time.Duration(cmd.Int(flgDNSTimeout))*time.Second)),
	)

	return err
}

func checkPropagationExclusiveOptions(cmd *cli.Command) error {
	if cmd.IsSet(flgDNSDisableCP) {
		log.Printf("The flag '%s' is deprecated use '%s' instead.", flgDNSDisableCP, flgDNSPropagationDisableANS)
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
