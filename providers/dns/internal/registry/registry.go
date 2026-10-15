//go:build lego_dnslight

package registry

import (
	"errors"
	"maps"
	"slices"

	"github.com/go-acme/lego/v5/challenge"
)

//nolint:gochecknoglobals // Required by the "dnslight" approach.
var providers = map[string]Builder{}

//nolint:gochecknoglobals // Required by the "dnslight" approach.
var providerAliases = map[string]Builder{}

type Builder func() (challenge.Provider, error)

// Add adds a DNS provider builder to the registry.
func Add(key string, fn Builder, aliases ...string) {
	providers[key] = fn

	for _, alias := range aliases {
		providerAliases[alias] = fn
	}
}

// Get gets a DNS provider builder from the registry.
func Get(key string) (Builder, error) {
	fn, ok := providers[key]
	if ok {
		return fn, nil
	}

	fn, ok = providerAliases[key]
	if ok {
		return fn, nil
	}

	return nil, errors.New("not found")
}

func GetCodes() []string {
	return slices.Collect(maps.Keys(providers))
}
