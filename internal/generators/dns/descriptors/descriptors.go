package descriptors

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Providers struct {
	Providers []Provider
}

type Provider struct {
	Name          string         // Real name of the DNS provider
	Code          string         // DNS code
	Aliases       []string       // DNS code aliases (for compatibility/deprecation)
	Since         string         // First lego version
	URL           string         // DNS provider URL
	Description   string         // Provider summary
	Example       string         // CLI example
	Configuration *Configuration // Environment variables
	Links         *Links         // Links
	Additional    string         // Extra documentation
	GeneratedFrom string         // Source file
}

type Configuration struct {
	Credentials map[string]string
	Additional  map[string]string
}

type Links struct {
	API      string
	GoClient string
}

// GetProviderInformation extract provider information from TOML description files.
func GetProviderInformation(root string) (*Providers, error) {
	models := &Providers{}

	err := filepath.Walk(filepath.Join(root, "providers", "dns"), walker(root, models))
	if err != nil {
		return nil, err
	}

	return models, nil
}

func walker(root string, prs *Providers) func(string, os.FileInfo, error) error {
	return func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".toml" {
			return nil
		}

		m := Provider{}

		m.GeneratedFrom, err = filepath.Rel(root, path)
		if err != nil {
			return err
		}

		_, err = toml.DecodeFile(path, &m)
		if err != nil {
			return err
		}

		prs.Providers = append(prs.Providers, m)

		return nil
	}
}
