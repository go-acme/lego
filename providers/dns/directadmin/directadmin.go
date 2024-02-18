package certbotdnsdirectadmin

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns"
)

// Environment variables for authentication
const (
	envDirectAdminEndpoint = "DIRECTADMIN_ENDPOINT"
	envDirectAdminUsername = "DIRECTADMIN_USERNAME"
	envDirectAdminPassword = "DIRECTADMIN_PASSWORD"
)

// DNSProvider is an implementation of the challenge.Provider interface
type DNSProvider struct {
	config      *Config
	directAdmin *DirectAdmin
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewConfig()

	// Load configuration from environment variables
	err := env.Parse(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate required fields
	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}

	return &DNSProvider{
		config:      config,
		directAdmin: NewDirectAdmin(config.Endpoint, config.Username, config.Password),
	}, nil
}

// Present creates a TXT record to fulfill the DNS challenge.
func (p *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := p.directAdmin.findZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("failed to find zone for %s: %w", fqdn, err)
	}

	// Create the TXT record
	err = p.directAdmin.createTXTRecord(zone, fqdn, value)
	if err != nil {
		return fmt.Errorf("failed to create TXT record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record after the challenge is complete.
func (p *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := p.directAdmin.findZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("failed to find zone for %s: %w", fqdn, err)
	}

	// Delete the TXT record
	err = p.directAdmin.deleteTXTRecord(zone, fqdn)
	if err != nil {
		return fmt.Errorf("failed to delete TXT record: %w", err)
	}

	return nil
}

// Timeout is the maximum time allowed for the Present and CleanUp methods.
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 30 * time.Minute, 15 * time.Second
}

// GetConfig returns the DNS provider configuration.
func (p *DNSProvider) GetConfig() challenge.Config {
	return challenge.Config{Present: p.Present, CleanUp: p.CleanUp, Timeout: p.Timeout}
}

// Config holds the configuration details for the DirectAdmin DNSProvider.
type Config struct {
	Endpoint string `env:"DIRECTADMIN_ENDPOINT"`
	Username string `env:"DIRECTADMIN_USERNAME"`
	Password string `env:"DIRECTADMIN_PASSWORD"`
}

// NewConfig returns a new Config instance.
func NewConfig() *Config {
	return &Config{}
}

// Validate checks if the required configuration fields are set.
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("DIRECTADMIN_ENDPOINT is not set")
	}
	if c.Username == "" {
		return fmt.Errorf("DIRECTADMIN_USERNAME is not set")
	}
	if c.Password == "" {
		return fmt.Errorf("DIRECTADMIN_PASSWORD is not set")
	}
	return nil
}

// DirectAdmin represents the DirectAdmin API client.
type DirectAdmin struct {
	endpoint string
	username string
	password string
}

// NewDirectAdmin returns a new DirectAdmin instance.
func NewDirectAdmin(endpoint, username, password string) *DirectAdmin {
	return &DirectAdmin{
		endpoint: endpoint,
		username: username,
		password: password,
	}
}

// findZoneByFqdn finds the DNS zone for a given FQDN.
func (da *DirectAdmin) findZoneByFqdn(fqdn string) (string, error) {
	// Implement logic to find the DNS zone for a given FQDN
	// ...

	return "", fmt.Errorf("not implemented")
}

// createTXTRecord creates a TXT record for the DNS challenge.
func (da *DirectAdmin) createTXTRecord(zone, fqdn, value string) error {
	// Implement logic to create a TXT record for the DNS challenge
	// ...

	return fmt.Errorf("not implemented")
}

// deleteTXTRecord deletes the TXT record after the challenge is complete.
func (da *DirectAdmin) deleteTXTRecord(zone, fqdn string) error {
	// Implement logic to delete the TXT record after the challenge is complete
	// ...

	return fmt.Errorf("not implemented")
}
