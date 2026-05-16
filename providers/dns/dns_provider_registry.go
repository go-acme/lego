//go:build lego_lightdns

package dns

import (
	"fmt"
	"sync"

	"github.com/go-acme/lego/v5/challenge"
)

/**
Experimental DNS provider loading.
Used to build a "light" CLI.

ex: `go build -tags lego_lightdns,lego_cloudflare -trimpath -ldflags '-s -w -X "main.version=${VERSION}"' -o ${BIN_OUTPUT} ${MAIN_DIRECTORY}`
*/

var providers = sync.Map{}

type Builder func() (challenge.Provider, error)

// Add adds a DNS provider builder to the registry.
func Add(key string, fn Builder) {
	providers.Store(key, fn)
}

// Get gets a DNS provider builder from the registry.
func Get(key string) (Builder, error) {
	fn, ok := providers.Load(key)
	if !ok {

		return nil, fmt.Errorf("%q not found", key)
	}

	return fn.(Builder), nil
}

// NewDNSChallengeProviderByName Factory for DNS providers.
func NewDNSChallengeProviderByName(name string) (challenge.Provider, error) {
	builder, err := Get(name)
	if err != nil {
		return nil, fmt.Errorf("unrecognized DNS provider %q: %w", name, err)
	}

	return builder()
}
