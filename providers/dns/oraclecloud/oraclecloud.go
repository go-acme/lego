// Package oraclecloud implements a DNS provider for solving the DNS-01 challenge using Oracle Cloud DNS.
package oraclecloud

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/oraclecloud/internal"
	"github.com/youmark/pkcs8"
)

// Environment variables names.
const (
	envNamespace = "OCI_"

	EnvCompartmentOCID   = envNamespace + "COMPARTMENT_OCID"
	envPrivKey           = envNamespace + "PRIVKEY"
	EnvPrivKeyFile       = envPrivKey + "_FILE"
	EnvPrivKeyPass       = envPrivKey + "_PASS"
	EnvTenancyOCID       = envNamespace + "TENANCY_OCID"
	EnvUserOCID          = envNamespace + "USER_OCID"
	EnvPubKeyFingerprint = envNamespace + "PUBKEY_FINGERPRINT"
	EnvRegion            = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// RecordOperationRemove is the operation to remove a record.
const RecordOperationRemove = "REMOVE"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	CompartmentID      string
	PrivateKey         *rsa.PrivateKey
	KeyID              string
	Region             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, time.Minute),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for OracleCloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(envPrivKey, EnvTenancyOCID, EnvUserOCID, EnvPubKeyFingerprint, EnvRegion, EnvCompartmentOCID)
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %w", err)
	}

	privateKey, err := loadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %w", err)
	}

	config := NewDefaultConfig()
	config.CompartmentID = values[EnvCompartmentOCID]
	config.PrivateKey = privateKey
	config.KeyID = fmt.Sprintf("%s/%s/%s", values[EnvTenancyOCID], values[EnvUserOCID], values[EnvPubKeyFingerprint])
	config.Region = values[EnvRegion]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OracleCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("oraclecloud: the configuration of the DNS provider is nil")
	}

	if config.PrivateKey == nil {
		return nil, errors.New("oraclecloud: PrivateKey is missing")
	}

	if config.KeyID == "" {
		return nil, errors.New("oraclecloud: KeyID is missing")
	}

	if config.Region == "" {
		return nil, errors.New("oraclecloud: Region is missing")
	}

	if config.CompartmentID == "" {
		return nil, errors.New("oraclecloud: CompartmentID is missing")
	}

	client := internal.NewClient(config.HTTPClient, config.PrivateKey, config.KeyID, config.Region, config.CompartmentID)

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneNameOrID, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("oraclecloud: could not find zone for domain %q: %w", domain, err)
	}

	// First, try to delete any existing challenge records to avoid duplicates
	err = d.client.FindAndDeleteDomainRecord(context.Background(), zoneNameOrID, dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		fmt.Printf("oraclecloud: warning: failed to clean up existing records: %v\n", err)
		// Continue even if deletion fails
	}

	// generate request to update records
	recordOperation := internal.RecordOperation{
		Domain:      dns01.UnFqdn(info.EffectiveFQDN),
		RData:       info.Value,
		RType:       "TXT",
		TTL:         d.config.TTL,
		IsProtected: false,
	}

	request := internal.PatchRecordsRequest{
		Items: []internal.RecordOperation{recordOperation},
	}

	err = d.client.PatchDomainRecords(context.Background(), zoneNameOrID, dns01.UnFqdn(info.EffectiveFQDN), request)
	if err != nil {
		return fmt.Errorf("oraclecloud: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneNameOrID, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("oraclecloud: could not find zone for domain %q: %w", domain, err)
	}

	err = d.client.FindAndDeleteDomainRecord(context.Background(), zoneNameOrID, dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		return fmt.Errorf("oraclecloud: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// loadPrivateKey gets the private key from the environment.
func loadPrivateKey() (*rsa.PrivateKey, error) {
	privateKeyData, err := getPrivateKeyData()
	if err != nil {
		return nil, err
	}

	privateKeyPass := env.GetOrFile(EnvPrivKeyPass)

	key, err := parsePrivateKey(privateKeyData, privateKeyPass)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// getPrivateKeyData gets the private key data from the environment.
func getPrivateKeyData() ([]byte, error) {
	envVarValue := os.Getenv(envPrivKey)
	if envVarValue != "" {
		bytes, err := base64.StdEncoding.DecodeString(envVarValue)
		if err != nil {
			return nil, fmt.Errorf("failed to read base64 value %s (defined by env var %s): %w", envVarValue, envPrivKey, err)
		}
		return bytes, nil
	}

	fileVar := EnvPrivKeyFile
	fileVarValue := os.Getenv(fileVar)
	if fileVarValue == "" {
		return nil, fmt.Errorf("no value provided for: %s or %s", envPrivKey, EnvPrivKeyFile)
	}

	fileContents, err := os.ReadFile(fileVarValue)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file %s (defined by env var %s): %w", fileVarValue, EnvPrivKeyFile, err)
	}

	return fileContents, nil
}

// parsePrivateKey parses a PEM encoded private key.
func parsePrivateKey(privateKey []byte, privateKeyPass string) (*rsa.PrivateKey, error) {
	keyBlock, _ := pem.Decode(privateKey)
	if keyBlock == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	// Handle different private key formats and encryption
	switch {
	// Unencrypted PKCS1 private key
	case keyBlock.Type == "RSA PRIVATE KEY" && !strings.Contains(string(keyBlock.Headers["Proc-Type"]), "ENCRYPTED"):
		key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err == nil {
			return key, nil
		}
		return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)

	// Unencrypted PKCS8 private key
	case keyBlock.Type == "PRIVATE KEY":
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err == nil {
			rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("private key is not an RSA private key")
			}
			return rsaKey, nil
		}
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)

	// PKCS8 Encrypted private key
	case keyBlock.Type == "ENCRYPTED PRIVATE KEY":
		if privateKeyPass == "" {
			return nil, errors.New("private key is encrypted but no password was provided")
		}

		// Use youmark/pkcs8 to parse encrypted PKCS8 key
		key, err := pkcs8.ParsePKCS8PrivateKey(keyBlock.Bytes, []byte(privateKeyPass))
		if err != nil {
			return nil, fmt.Errorf("failed to parse encrypted PKCS8 private key: %w", err)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not an RSA private key")
		}
		return rsaKey, nil

	// Encrypted PKCS1 private key (legacy OpenSSL format)
	case keyBlock.Type == "RSA PRIVATE KEY" && len(keyBlock.Headers) > 0:
		if privateKeyPass == "" {
			return nil, errors.New("private key is encrypted but no password was provided")
		}

		// Handle traditional OpenSSL encryption (for backward compatibility)
		// This is deprecated but we need to support existing keys
		// This will only be used if the PEM header contains encryption info
		// #nosec G501 - We need to support legacy encryption formats
		if x509.IsEncryptedPEMBlock(keyBlock) {
			// #nosec G501 - We need to support legacy encryption formats
			decryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(privateKeyPass))
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt private key: %w", err)
			}

			// Try PKCS1 format first
			key, err := x509.ParsePKCS1PrivateKey(decryptedKey)
			if err == nil {
				return key, nil
			}

			// Then try PKCS8
			pkcs8Key, err := x509.ParsePKCS8PrivateKey(decryptedKey)
			if err == nil {
				rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
				if !ok {
					return nil, errors.New("private key is not an RSA private key")
				}
				return rsaKey, nil
			}

			return nil, fmt.Errorf("failed to parse decrypted private key: %w", err)
		}

		return nil, errors.New("unsupported encrypted private key format")

	default:
		return nil, fmt.Errorf("unsupported private key type: %s", keyBlock.Type)
	}
}
