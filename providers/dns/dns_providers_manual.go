//go:build manualdnsonly

package dns

import (
	"fmt"
	"github.com/go-acme/lego/v4/challenge/dns01"

	"github.com/go-acme/lego/v4/challenge"
)

// NewDNSChallengeProviderByName Factory for DNS providers.
func NewDNSChallengeProviderByName(name string) (challenge.Provider, error) {
	switch name {
	case "manual":
		return dns01.NewDNSProviderManual()
	default:
		return nil, fmt.Errorf("unrecognized DNS provider: %s", name)
	}
}
