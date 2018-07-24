// package multi implements a dynamic challenge provider that can select different dns providers for different domains,
// and even multiple distinct dns providers and accounts for each individual domain. This can be useful if:
//
// - Multiple dns providers are used for active-active redundant dns service
//
// - You need a single certificate issued for different domains, each using different dns services
//
// Configuration is given by selecting DNS provider type "multi", and by giving further per-domain information via a json object:
//
//  {
//    "Providers": {
//      "cloudflare": {
//        "CLOUDFLARE_EMAIL": "myacct@example.com",
//        "CLOUDFLARE_API_KEY": "123qwerty"
//      },
//      "digitalocean":{
//        "DO_AUTH_TOKEN": "456uiop"
//      }
//    }
//    "Domains": {
//      "example.com": ["digitalocean"],
//      "example.org": ["cloudflare"],
//      "example.net": ["digitalocean, cloudflare"]
//    }
//  }
//
// In the above json, each "Provider" is a named provider instance along with the associated credentials. The credentials will be set as environment
// variables as appropriate when the provider is instantiated for the first time.
//
// If the provider name is the same as a registered provider type (like "cloudflare"), the type will be inferred. If it is not the same (perhaps in cases where multiple
// different accounts are involved), you may specify it with the `type` field on the provider object.
//
// Domains are then linked to one or more of the named providers by name. Challenges will be filled on every provider specified for the domain. When looking for a domain
// configuration, config domains will be checked from most specific to least specific by each dot. For example, to fill a challenge for `foo.example.com`,
// a configured domain for `foo.example.com` will be looked for, failing that it will look for `example.com` and `com` in that order. If there is still no match and a
// domain with the name `default` is found, that will be used. Otherwise an error will be returned.
//
// The json configuration for domains can be specified directly via environment variable (`MULTI_CONFIG`), or from a file referenced by `MULTI_CONFIG_FILE`.
package multi

import (
	"fmt"
	"os"
	"time"

	"github.com/xenolf/lego/acme"
)

// NewDNSChallengeProviderByName is defined here to avoid recursive imports, this must be injected by the dns package so that
// the delegated dns providers may be dynamically instantiated
var NewDNSChallengeProviderByName func(string) (acme.ChallengeProvider, error)

type MultiProvider struct {
	config    *MultiProviderConfig
	providers map[string]acme.ChallengeProvider
}

// AggregateProvider is simply a list of dns providers. All Challenges are filled by all members of the aggregate.
type AggregateProvider []acme.ChallengeProvider

func (a AggregateProvider) Present(domain, token, keyAuth string) error {
	for _, p := range a {
		if err := p.Present(domain, token, keyAuth); err != nil {
			return err
		}
	}
	return nil
}
func (a AggregateProvider) CleanUp(domain, token, keyAuth string) error {
	for _, p := range a {
		if err := p.CleanUp(domain, token, keyAuth); err != nil {
			return err
		}
	}
	return nil
}

// AggregateProviderTimeout is simply a list of dns providers. This type will be chosen when any of the 'subproviders' implement Timeout control.
// All Challenges are filled by all members of the aggregate.
// Timeout returned will be the maximum time of any child provider.
type AggregateProviderTimeout struct {
	AggregateProvider
}

func (a AggregateProviderTimeout) Timeout() (timeout, interval time.Duration) {
	for _, p := range a.AggregateProvider {
		if to, ok := p.(acme.ChallengeProviderTimeout); ok {
			t, i := to.Timeout()
			if t > timeout {
				timeout = t
			}
			if i > interval {
				interval = i
			}
		}
	}
	return
}

func (m *MultiProvider) getProviderForDomain(domain string) (acme.ChallengeProvider, error) {
	names, err := m.config.providerNamesForDomain(domain)
	if err != nil {
		return nil, err
	}
	var agg AggregateProvider
	anyTimeouts := false
	for _, n := range names {
		p, err := m.providerByName(n)
		if err != nil {
			return nil, err
		}
		if _, ok := p.(acme.ChallengeProviderTimeout); ok {
			anyTimeouts = true
		}
		agg = append(agg, p)
	}
	// don't wrap provider in aggregate if there is only one
	if len(agg) == 1 {
		return agg[0], nil
	}
	if anyTimeouts {
		return AggregateProviderTimeout{agg}, nil
	}
	return agg, nil
}

func (m *MultiProvider) providerByName(name string) (acme.ChallengeProvider, error) {
	if p, ok := m.providers[name]; ok {
		return p, nil
	}
	if params, ok := m.config.Providers[name]; ok {
		return m.buildProvider(name, params)
	}
	return nil, fmt.Errorf("Couldn't find appropriate config for dns provider named '%s'", name)
}

func (m *MultiProvider) buildProvider(name string, params map[string]string) (acme.ChallengeProvider, error) {
	pType := name
	origEnv := map[string]string{}

	// copy parameters into environment, keeping track of previous values
	for k, v := range params {
		if k == "type" {
			pType = v
			continue
		}
		if oldVal, ok := os.LookupEnv(k); ok {
			origEnv[k] = oldVal
		}
		os.Setenv(k, v)
	}
	// restore previous values
	defer func() {
		for k := range params {
			if k == "type" {
				continue
			}
			if oldVal, ok := origEnv[k]; ok {
				os.Setenv(k, oldVal)
			} else {
				os.Unsetenv(k)
			}
		}
	}()
	prv, err := NewDNSChallengeProviderByName(pType)
	if err != nil {
		return nil, err
	}
	m.providers[name] = prv
	return prv, nil
}

func New() (*MultiProvider, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	return &MultiProvider{
		providers: map[string]acme.ChallengeProvider{},
		config:    config,
	}, nil
}

func (m *MultiProvider) Present(domain, token, keyAuth string) error {
	provider, err := m.getProviderForDomain(domain)
	if err != nil {
		return err
	}
	return provider.Present(domain, token, keyAuth)
}

func (m *MultiProvider) CleanUp(domain, token, keyAuth string) error {
	provider, err := m.getProviderForDomain(domain)
	if err != nil {
		return err
	}
	return provider.CleanUp(domain, token, keyAuth)
}
