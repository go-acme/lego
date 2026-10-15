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

type Builder func() (challenge.Provider, error)

// Add adds a DNS provider builder to the registry.
func Add(key string, fn Builder) {
	providers[key] = fn
}

// Get gets a DNS provider builder from the registry.
func Get(key string) (Builder, error) {
	fn, ok := providers[key]
	if !ok {
		return nil, errors.New("not found")
	}

	return fn, nil
}

func GetCodes() []string {
	return slices.Collect(maps.Keys(providers))
}
