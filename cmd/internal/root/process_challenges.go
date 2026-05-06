package root

import (
	"fmt"
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
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/providers/http/memcached"
	"github.com/go-acme/lego/v5/providers/http/s3"
	"github.com/go-acme/lego/v5/providers/http/webroot"
)

func setupChallenges(client *lego.Client, chlgConfig *configuration.Challenge, networkStack challenge.NetworkStack) error {
	if chlgConfig.HTTP != nil {
		err := setupHTTPProvider(client, chlgConfig.HTTP, networkStack)
		if err != nil {
			return fmt.Errorf("HTTP challenge provider: %w", err)
		}
	}

	if chlgConfig.TLS != nil {
		err := setupTLSProvider(client, chlgConfig.TLS, networkStack)
		if err != nil {
			return fmt.Errorf("TLS challenge provider: %w", err)
		}
	}

	if chlgConfig.DNS != nil {
		err := setupDNS(client, chlgConfig.DNS, networkStack)
		if err != nil {
			return fmt.Errorf("DNS challenge provider: %w", err)
		}
	}

	if chlgConfig.DNSPersist != nil {
		err := setupDNSPersist(client, chlgConfig.DNSPersist, networkStack)
		if err != nil {
			return fmt.Errorf("DNS-PERSIST challenge provider: %w", err)
		}
	}

	return nil
}

func setupHTTPProvider(client *lego.Client, chlg *configuration.HTTPChallenge, networkStack challenge.NetworkStack) error {
	provider, err := createHTTPProvider(chlg, networkStack)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return client.Challenge.SetHTTP01Provider(provider, http01.SetDelay(chlg.Delay))
}

func createHTTPProvider(chlg *configuration.HTTPChallenge, networkStack challenge.NetworkStack) (challenge.Provider, error) {
	switch {
	case chlg.Webroot != "":
		ps, err := webroot.NewHTTPProvider(chlg.Webroot)
		if err != nil {
			return nil, fmt.Errorf("webroot provider (%s): %w", chlg.Webroot, err)
		}

		return ps, nil

	case len(chlg.MemcachedHosts) > 0:
		ps, err := memcached.NewMemcachedProvider(chlg.MemcachedHosts)
		if err != nil {
			return nil, fmt.Errorf("memcached provider (%s): %w", strings.Join(chlg.MemcachedHosts, ", "), err)
		}

		return ps, nil

	case chlg.S3Bucket != "":
		ps, err := s3.NewHTTPProvider(chlg.S3Bucket)
		if err != nil {
			return nil, fmt.Errorf("S3 provider (%s): %w", chlg.S3Bucket, err)
		}

		return ps, nil

	case chlg.Address != "":
		host, port, err := parseAddress(chlg.Address)
		if err != nil {
			return nil, err
		}

		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network:         networkStack.Network("tcp"),
			Address:         net.JoinHostPort(host, port),
			ProxyHeaderName: chlg.ProxyHeader,
		})

		return srv, nil

	default:
		srv := http01.NewProviderServerWithOptions(http01.Options{
			Network:         networkStack.Network("tcp"),
			Address:         net.JoinHostPort("", ":80"),
			ProxyHeaderName: chlg.ProxyHeader,
		})

		return srv, nil
	}
}

func setupTLSProvider(client *lego.Client, chlg *configuration.TLSChallenge, networkStack challenge.NetworkStack) error {
	options := tlsalpn01.Options{
		Network: networkStack.Network("tcp"),
	}

	if chlg.Address != "" {
		host, port, err := parseAddress(chlg.Address)
		if err != nil {
			return err
		}

		options.Host = host
		options.Port = port
	}

	return client.Challenge.SetTLSALPN01Provider(
		tlsalpn01.NewProviderServerWithOptions(options),
		tlsalpn01.SetDelay(chlg.Delay),
	)
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
