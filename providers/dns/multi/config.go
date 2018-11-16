package multi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// ProviderConfig is the configuration for a multiple provider setup. This is expected to be given in json format via
// MULTI_CONFIG environment variable, or in a file location specified by MULTI_CONFIG_FILE.
type ProviderConfig struct {
	// Domain names to list of provider names
	Domains map[string][]string
	// Provider Name -> Key/Value pairs for environment
	Providers map[string]map[string]string
}

// providerNamesForDomain chooses the most appropriate domain from the config and returns its' list of dns providers
// looks for most specific match to least specific, one dot at a time. Finally folling back to "default" domain.
func (m *ProviderConfig) providerNamesForDomain(domain string) ([]string, error) {
	parts := strings.Split(domain, ".")
	var names []string
	for i := 0; i < len(parts); i++ {
		partial := strings.Join(parts[i:], ".")
		if names = m.Domains[partial]; names != nil {
			break
		}
	}
	if names == nil {
		names = m.Domains["default"]
	}
	if names == nil {
		return nil, fmt.Errorf("Couldn't find any suitable dns provider for domain %s", domain)
	}
	return names, nil
}

func getConfig() (*ProviderConfig, error) {
	var rawJSON []byte
	var err error
	if cfg := os.Getenv("MULTI_CONFIG"); cfg != "" {
		rawJSON = []byte(cfg)
	} else if path := os.Getenv("MULTI_CONFIG_FILE"); path != "" {
		if rawJSON, err = ioutil.ReadFile(path); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("'multi' provider requires json config in MULTI_CONFIG or MULTI_CONFIG_FILE")
	}
	cfg := &ProviderConfig{}
	if err = json.Unmarshal(rawJSON, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
