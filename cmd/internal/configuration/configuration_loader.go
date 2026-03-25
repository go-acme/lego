package configuration

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"gopkg.in/yaml.v3"
)

// ReadConfiguration reads the configuration file and returns a Configuration struct.
func ReadConfiguration(filename string) (*Configuration, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open the configuration file: %w", err)
	}

	defer func() { _ = file.Close() }()

	cfg := new(Configuration)

	err = yaml.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not decode the configuration file: %w", err)
	}

	return cfg, nil
}

// FindDefaultConfigurationFile returns the path of the default configuration file.
func FindDefaultConfigurationFile() (string, error) {
	extensions := []string{".yml", ".yaml"}

	for _, ext := range extensions {
		filename := ".lego" + ext

		ok, err := exists(filename)
		if err != nil {
			return "", err
		}

		if !ok {
			continue
		}

		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return filepath.Join(dir, filename), nil
	}

	return "", errors.New("no configuration file found")
}

func exists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return !stat.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// GetServerConfig returns the server configuration for the given server name or URL.
func GetServerConfig(cfg *Configuration, accountID string) *Server {
	if len(cfg.Accounts) == 0 {
		return &Server{URL: lego.DirectoryURLLetsEncrypt}
	}

	accountConfig := cfg.Accounts[accountID]

	if len(cfg.Servers) == 0 {
		log.Debug("No server configuration",
			slog.String("account", accountID),
			slog.String("server", accountConfig.Server),
		)

		return &Server{
			URL:                 accountConfig.Server,
			OverallRequestLimit: certificate.DefaultOverallRequestLimit,
		}
	}

	if _, ok := cfg.Servers[accountConfig.Server]; ok {
		return cfg.Servers[accountConfig.Server]
	}

	log.Debug("Server configuration not found.",
		slog.String("account", accountID),
		slog.String("server", accountConfig.Server),
	)

	directoryURL, err := lego.GetDirectoryURL(accountConfig.Server)
	if err == nil {
		return &Server{
			URL:                 directoryURL,
			OverallRequestLimit: certificate.DefaultOverallRequestLimit,
		}
	}

	log.Debug("Server shortcode not found.",
		slog.String("account", accountID),
		slog.String("server", accountConfig.Server),
		log.ErrorAttr(err),
	)

	return &Server{
		URL:                 accountConfig.Server,
		OverallRequestLimit: certificate.DefaultOverallRequestLimit,
	}
}
