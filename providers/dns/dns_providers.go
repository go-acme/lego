package dns

import (
	"github.com/go-acme/lego/v4/challenge"
)

// NewDNSChallengeProviderByName Factory for DNS providers.
// Return an ErrUnrecognizedProvider if name doesn't match a supported proivder.
func NewDNSChallengeProviderByName(name string) (challenge.Provider, error) {
	var p SupportedProvider
	if e := p.set(name); e != nil {
		return nil, e
	}

	return _provider2func[p]()
}

// IsProviderSupported return true if the name param match a
// supported provider.
func IsProviderSupported(name string) bool {
	var p SupportedProvider

	return p.set(name) == nil
}

// GetSupportedProvider return a list of supported provider name.
func GetSupportedProvider() []string {
	i, out := 0, make([]string, len(_str2provider))
	for name := range _str2provider {
		out[i] = name
		i++
	}

	return out
}
