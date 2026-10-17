//go:build lego_dnslight

package registry

import (
	"errors"
	"maps"
	"slices"

	"github.com/go-acme/lego/v5/challenge"
)

//nolint:gochecknoglobals // Required by the "dnslight" approach.
var store = New()

type Builder func() (challenge.Provider, error)

type Store struct {
	providers       map[string]Builder
	providerAliases map[string]Builder
}

func New() *Store {
	return &Store{
		providers:       map[string]Builder{},
		providerAliases: map[string]Builder{},
	}
}

// Add adds a DNS provider builder to the registry.
func (s *Store) Add(key string, fn Builder, aliases ...string) {
	s.providers[key] = fn

	for _, alias := range aliases {
		s.providerAliases[alias] = fn
	}
}

// Get gets a DNS provider builder from the registry.
func (s *Store) Get(key string) (Builder, error) {
	fn, ok := s.providers[key]
	if ok {
		return fn, nil
	}

	fn, ok = s.providerAliases[key]
	if ok {
		return fn, nil
	}

	return nil, errors.New("not found")
}

// GetCodes returns all available provider codes (without aliases).
func (s *Store) GetCodes() []string {
	return slices.Collect(maps.Keys(s.providers))
}

// Add adds a DNS provider builder to the registry.
func Add(key string, fn Builder, aliases ...string) {
	store.Add(key, fn, aliases...)
}

// Get gets a DNS provider builder from the registry.
func Get(key string) (Builder, error) {
	return store.Get(key)
}

// GetCodes returns all available provider codes (without aliases).
func GetCodes() []string {
	return store.GetCodes()
}
