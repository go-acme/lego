//go:build lego_dnslight

package dns

import (
	"fmt"

	"github.com/go-acme/lego/v5/challenge"
	_ "github.com/go-acme/lego/v5/providers/dns/internal/dnslight"
	"github.com/go-acme/lego/v5/providers/dns/internal/registry"
)

/**
Experimental DNS provider loading.

ex: `go build -tags lego_dnslight,lego_dns_cloudflare -trimpath -ldflags '-s -w -X "main.version=${VERSION}"' -o ${BIN_OUTPUT} ${MAIN_DIRECTORY}`
*/

// NewDNSChallengeProviderByName Factory for DNS providers.
func NewDNSChallengeProviderByName(name string) (challenge.Provider, error) {
	builder, err := registry.Get(name)
	if err != nil {
		return nil, fmt.Errorf("unrecognized DNS provider (lego_dnslight) %q: %w", name, err)
	}

	return builder()
}

// GetProviderCodes returns the list of available DNS providers.
func GetProviderCodes() []string {
	return registry.GetCodes()
}
