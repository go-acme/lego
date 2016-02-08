package dns_provider

import "github.com/xenolf/lego/acme"

type providerFactory func() (acme.ChallengeProvider, error)

type ProviderRegistry struct {
	entries []ProviderRegistryEntry
}

// The global registry object, accessible outside the package
var Registry ProviderRegistry

// Add a provider to the registry, so external consumers can instantiate them as needed.
func (r *ProviderRegistry) addProvider(name string, configDescription string, factory providerFactory) {
	entry := ProviderRegistryEntry{
		Name:              name,
		ConfigDescription: configDescription,
		factory:           factory,
	}
	r.entries = append(r.entries, entry)
}

// Returns a list of registered providers
func (reg ProviderRegistry) Entries() []ProviderRegistryEntry {
	return reg.entries
}

// Looks up a provider by name, returning nil if not found
func (reg ProviderRegistry) FindEntryByName(name string) *ProviderRegistryEntry {
	for _, e := range reg.entries {
		if e.Name == name {
			return &e
		}
	}

	return nil
}

// Represents a single provider
type ProviderRegistryEntry struct {
	Name              string
	ConfigDescription string
	factory           providerFactory
}

// Given a ProviderRegistryEntry, return a new acme.ChallengeProvider. The challenge provider
// will attempt to configure itself based on environment variables supplied by os.Getenv().
// If this fails, you'll get a nil ChallengeProvider and a non-nil error.
func (pre ProviderRegistryEntry) NewChallengeProvider() (acme.ChallengeProvider, error) {
	return pre.factory()
}
