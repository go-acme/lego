package root

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
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/providers/http/memcached"
	"github.com/go-acme/lego/v5/providers/http/s3"
	"github.com/go-acme/lego/v5/providers/http/webroot"
)

func setupChallenges(client *lego.Client, chlgConfig *configuration.Challenge, networkStack challenge.NetworkStack) {
	if chlgConfig.HTTP != nil {
		err := client.Challenge.SetHTTP01Provider(setupHTTPProvider(chlgConfig.HTTP, networkStack), http01.SetDelay(chlgConfig.HTTP.Delay))
		if err != nil {
			log.Fatal("Could not set HTTP challenge provider.", log.ErrorAttr(err))
		}
	}

	if chlgConfig.TLS != nil {
		err := client.Challenge.SetTLSALPN01Provider(setupTLSProvider(chlgConfig.TLS, networkStack), tlsalpn01.SetDelay(chlgConfig.TLS.Delay))
		if err != nil {
			log.Fatal("Could not set TLS challenge provider.", log.ErrorAttr(err))
		}
	}

	if chlgConfig.DNS != nil {
		err := setupDNS(client, chlgConfig.DNS, networkStack)
		if err != nil {
			log.Fatal("Could not set DNS challenge provider.", log.ErrorAttr(err))
		}
	}

	if chlgConfig.DNSPersist != nil {
		err := setupDNSPersist(client, chlgConfig.DNSPersist, networkStack)
		if err != nil {
			log.Fatal("Could not set DNS-PERSIST challenge provider.", log.ErrorAttr(err))
		}
	}
}

func setupHTTPProvider(chlg *configuration.HTTPChallenge, networkStack challenge.NetworkStack) challenge.Provider {
	switch {
	case chlg.Webroot != "":
		ps, err := webroot.NewHTTPProvider(chlg.Webroot)
		if err != nil {
			log.Fatal("Could not create the webroot provider.",
				slog.String("webRoot", chlg.Webroot),
				log.ErrorAttr(err),
			)
		}

		return ps

	case len(chlg.MemcachedHosts) > 0:
		ps, err := memcached.NewMemcachedProvider(chlg.MemcachedHosts)
		if err != nil {
			log.Fatal("Could not create the memcached provider.",
				slog.String("memcachedHosts", strings.Join(chlg.MemcachedHosts, ", ")),
				log.ErrorAttr(err),
			)
		}

		return ps

	case chlg.S3Bucket != "":
		ps, err := s3.NewHTTPProvider(chlg.S3Bucket)
		if err != nil {
			log.Fatal("Could not create the S3 provider.",
				slog.String("bucket", chlg.S3Bucket),
				log.ErrorAttr(err),
			)
		}

		return ps

	case chlg.Address != "":
		host, port, err := parseAddress(chlg.Address)
		if err != nil {
			log.Fatal("Could not split host and port.", slog.String("iface", chlg.Address), log.ErrorAttr(err))
		}

		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network: networkStack.Network("tcp"),
			Address: net.JoinHostPort(host, port),
		})

		if header := chlg.ProxyHeader; header != "" {
			srv.SetProxyHeader(header)
		}

		return srv

	default:
		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network: networkStack.Network("tcp"),
			Address: net.JoinHostPort("", ":80"),
		})

		if header := chlg.ProxyHeader; header != "" {
			srv.SetProxyHeader(header)
		}

		return srv
	}
}

func setupTLSProvider(chlg *configuration.TLSChallenge, networkStack challenge.NetworkStack) challenge.Provider {
	switch {
	case chlg.Address != "":
		host, port, err := parseAddress(chlg.Address)
		if err != nil {
			log.Fatal("Could not split host and port.", slog.String("iface", chlg.Address), log.ErrorAttr(err))
		}

		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			Network: networkStack.Network("tcp"),
			Host:    host,
			Port:    port,
		})

	default:
		return tlsalpn01.NewProviderServerWithOptions(tlsalpn01.Options{
			Network: networkStack.Network("tcp"),
		})
	}
}

func setupDNS(client *lego.Client, chlg *configuration.DNSChallenge, networkStack challenge.NetworkStack) error {
	provider, err := dns.NewDNSChallengeProviderByName(chlg.Provider)
	if err != nil {
		return err
	}

	opts := &dns01.Options{RecursiveNameservers: chlg.Resolvers}

	if chlg.DNSTimeout > 0 {
		opts.Timeout = time.Duration(chlg.DNSTimeout) * time.Second
	}

	opts.NetworkStack = networkStack

	dns01.SetDefaultClient(dns01.NewClient(opts))

	return client.Challenge.SetDNS01Provider(provider,
		dns01.LazyCondOption(chlg.Propagation != nil, func() dns01.ChallengeOption {
			if chlg.Propagation.Wait > 0 {
				return dns01.PropagationWait(chlg.Propagation.Wait, true)
			}

			return dns01.CombineOptions(
				dns01.CondOptions(chlg.Propagation.DisableAuthoritativeNameservers,
					dns01.DisableAuthoritativeNssPropagationRequirement(),
				),
				dns01.CondOptions(chlg.Propagation.DisableRecursiveNameservers,
					dns01.DisableRecursiveNSsPropagationRequirement(),
				),
			)
		}),
	)
}

func setupDNSPersist(client *lego.Client, chlg *configuration.DNSPersistChallenge, networkStack challenge.NetworkStack) error {
	opts := &dns01.Options{RecursiveNameservers: chlg.Resolvers}

	if chlg.DNSTimeout > 0 {
		opts.Timeout = time.Duration(chlg.DNSTimeout) * time.Second
	}

	opts.NetworkStack = networkStack

	dnspersist01.SetDefaultClient(dnspersist01.NewClient(opts))

	return client.Challenge.SetDNSPersist01(
		dnspersist01.WithIssuerDomainName(chlg.IssuerDomainName),
		dnspersist01.CondOptions(!chlg.PersistUntil.IsZero(),
			dnspersist01.WithPersistUntil(chlg.PersistUntil),
		),
		dnspersist01.LazyCondOption(chlg.Propagation != nil, func() dnspersist01.ChallengeOption {
			if chlg.Propagation.Wait > 0 {
				return dnspersist01.PropagationWait(chlg.Propagation.Wait, true)
			}

			return dnspersist01.CombineOptions(
				dnspersist01.CondOptions(chlg.Propagation.DisableAuthoritativeNameservers,
					dnspersist01.DisableAuthoritativeNssPropagationRequirement(),
				),
				dnspersist01.CondOptions(chlg.Propagation.DisableRecursiveNameservers,
					dnspersist01.DisableRecursiveNSsPropagationRequirement(),
				),
			)
		}),
	)
}
