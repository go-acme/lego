package storage

import (
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"gopkg.in/yaml.v3"
)

// ConfigurationStorage is a storage for the configuration.
// The backup is the effective configuration and cannot be used directly.
type ConfigurationStorage struct {
	backupPath string
}

// NewConfigurationStorage creates a new ConfigurationStorage.
func NewConfigurationStorage(basePath string) *ConfigurationStorage {
	return &ConfigurationStorage{
		backupPath: filepath.Join(basePath, ".lego.bck.yaml"),
	}
}

// Backup saves the configuration to the storage.
func (s *ConfigurationStorage) Backup(cfg *configuration.Configuration) error {
	file, err := os.Create(s.backupPath)
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	defer func() { _ = encoder.Close() }()

	return encoder.Encode(cfg)
}

// ReadBackup reads the backup configuration.
func (s *ConfigurationStorage) ReadBackup() (*configuration.Configuration, error) {
	file, err := os.Open(s.backupPath)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	cfg := &configuration.Configuration{}

	err = yaml.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
