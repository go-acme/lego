// Package acmedns implements a DNS provider for solving DNS-01 challenges using Joohoi's acme-dns project.
// For more information see the ACME-DNS homepage: https://github.com/joohoi/acme-dns
package acmedns

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/acmedns/internal"
	"github.com/nrdcg/goacmedns"
	"github.com/nrdcg/goacmedns/storage"
)

const (
	// envNamespace is the prefix for ACME-DNS environment variables.
	envNamespace = "ACME_DNS_"

	// EnvAPIBase is the environment variable name for the ACME-DNS API address.
	// (e.g. https://acmedns.your-domain.com).
	EnvAPIBase = envNamespace + "API_BASE"

	// EnvAllowList are source networks using CIDR notation,
	// e.g. "192.168.100.1/24,1.2.3.4/32,2002:c0a8:2a00::0/40".
	EnvAllowList = envNamespace + "ALLOWLIST"

	// EnvStoragePath is the environment variable name for the ACME-DNS JSON account data file.
	// A per-domain account will be registered/persisted to this file and used for TXT updates.
	EnvStoragePath = envNamespace + "STORAGE_PATH"

	// EnvStorageBaseURL  is the environment variable name for the ACME-DNS JSON account data.
	// The URL to the storage server.
	EnvStorageBaseURL = envNamespace + "STORAGE_BASE_URL"
)

var _ challenge.Provider = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIBase        string
	AllowList      []string
	StoragePath    string
	StorageBaseURL string
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{}
}

// acmeDNSClient is an interface describing the goacmedns.Client functions the DNSProvider uses.
// It makes it easier for tests to shim a mock Client into the DNSProvider.
type acmeDNSClient interface {
	// UpdateTXTRecord updates the provided account's TXT record
	// to the given value or returns an error.
	UpdateTXTRecord(ctx context.Context, account goacmedns.Account, value string) error
	// RegisterAccount registers and returns a new account
	// with the given allowFrom restriction or returns an error.
	RegisterAccount(ctx context.Context, allowFrom []string) (goacmedns.Account, error)
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config  *Config
	client  acmeDNSClient
	storage goacmedns.Storage
}

// NewDNSProvider returns a DNSProvider instance configured for Joohoi's acme-dns.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIBase)
	if err != nil {
		return nil, fmt.Errorf("acme-dns: %w", err)
	}

	config := NewDefaultConfig()
	config.APIBase = values[EnvAPIBase]
	config.StoragePath = env.GetOrFile(EnvStoragePath)
	config.StorageBaseURL = env.GetOrFile(EnvStorageBaseURL)

	allowList := env.GetOrFile(EnvAllowList)
	if allowList != "" {
		config.AllowList = strings.Split(allowList, ",")
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Joohoi's acme-dns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("acme-dns: the configuration of the DNS provider is nil")
	}

	st, err := getStorage(config)
	if err != nil {
		return nil, fmt.Errorf("acme-dns: %w", err)
	}

	client, err := goacmedns.NewClient(config.APIBase)
	if err != nil {
		return nil, fmt.Errorf("acme-dns: new client: %w", err)
	}

	return &DNSProvider{
		config:  config,
		client:  client,
		storage: st,
	}, nil
}

// NewDNSProviderClient creates an ACME-DNS DNSProvider with the given acmeDNSClient and [goacmedns.Storage].
//
// Deprecated: use [NewDNSProviderConfig] instead.
func NewDNSProviderClient(client acmeDNSClient, store goacmedns.Storage) (*DNSProvider, error) {
	if client == nil {
		return nil, errors.New("acme-dns: Client must be not nil")
	}

	if store == nil {
		return nil, errors.New("acme-dns: Storage must be not nil")
	}

	return &DNSProvider{
		config:  NewDefaultConfig(),
		client:  client,
		storage: store,
	}, nil
}

// ErrCNAMERequired is returned by Present when the Domain indicated had no
// existing ACME-DNS account in the Storage and additional setup is required.
// The user must create a CNAME in the DNS zone for Domain that aliases FQDN
// to Target in order to complete setup for the ACME-DNS account that was created.
type ErrCNAMERequired struct {
	// The Domain that is being issued for.
	Domain string
	// The alias of the CNAME (left hand DNS label).
	FQDN string
	// The RDATA of the CNAME (right hand side, canonical name).
	Target string
}

// Error returns a descriptive message for the ErrCNAMERequired instance telling
// the user that a CNAME needs to be added to the DNS zone of c.Domain before
// the ACME-DNS hook will work.
// The CNAME to be created should be of the form: {{ c.FQDN }} 	CNAME	{{ c.Target }}.
func (e ErrCNAMERequired) Error() string {
	return fmt.Sprintf("acme-dns: new account created for %q. "+
		"To complete setup for %q you must provision the following "+
		"CNAME in your DNS zone and re-run this provider when it is "+
		"in place:\n"+
		"%s CNAME %s.",
		e.Domain, e.Domain, e.FQDN, e.Target)
}

// Present creates a TXT record to fulfill the DNS-01 challenge.
// If there is an existing account for the domain in the provider's storage
// then it will be used to set the challenge response TXT record with the ACME-DNS server and issuance will continue.
// If there is not an account for the given domain present in the DNSProvider storage
// one will be created and registered with the ACME DNS server and an ErrCNAMERequired error is returned.
// This will halt issuance and indicate to the user that a one-time manual setup is required for the domain.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	ctx := context.Background()

	// Compute the challenge response FQDN and TXT value for the domain based on the keyAuth.
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// Check if credentials were previously saved for this domain.
	account, err := d.storage.Fetch(ctx, domain)
	if err != nil {
		if !errors.Is(err, storage.ErrDomainNotFound) {
			return err
		}

		// The account did not exist.
		// Create a new one and return an error indicating the required one-time manual CNAME setup.
		account, err = d.register(ctx, domain, info.FQDN)
		if err != nil {
			return err
		}
	}

	// Update the acme-dns TXT record.
	return d.client.UpdateTXTRecord(ctx, account, info.Value)
}

// CleanUp removes the record matching the specified parameters. It is not
// implemented for the ACME-DNS provider.
func (d *DNSProvider) CleanUp(_, _, _ string) error {
	// ACME-DNS doesn't support the notion of removing a record.
	// For users of ACME-DNS it is expected the stale records remain in-place.
	return nil
}

// register creates a new ACME-DNS account for the given domain.
// If account creation works as expected a ErrCNAMERequired error is returned describing
// the one-time manual CNAME setup required to complete setup of the ACME-DNS hook for the domain.
// If any other error occurs it is returned as-is.
func (d *DNSProvider) register(ctx context.Context, domain, fqdn string) (goacmedns.Account, error) {
	newAcct, err := d.client.RegisterAccount(ctx, d.config.AllowList)
	if err != nil {
		return goacmedns.Account{}, err
	}

	var cnameCreated bool

	// Store the new account in the storage and call save to persist the data.
	err = d.storage.Put(ctx, domain, newAcct)
	if err != nil {
		cnameCreated = errors.Is(err, internal.ErrCNAMEAlreadyCreated)
		if !cnameCreated {
			return goacmedns.Account{}, err
		}
	}

	err = d.storage.Save(ctx)
	if err != nil {
		return goacmedns.Account{}, err
	}

	if cnameCreated {
		return newAcct, nil
	}

	// Stop issuance by returning an error.
	// The user needs to perform a manual one-time CNAME setup in their DNS zone
	// to complete the setup of the new account we created.
	return goacmedns.Account{}, ErrCNAMERequired{
		Domain: domain,
		FQDN:   fqdn,
		Target: newAcct.FullDomain,
	}
}

func getStorage(config *Config) (goacmedns.Storage, error) {
	if config.StoragePath == "" && config.StorageBaseURL == "" {
		return nil, errors.New("storagePath or storageBaseURL is not set")
	}

	if config.StoragePath != "" && config.StorageBaseURL != "" {
		return nil, errors.New("storagePath and storageBaseURL cannot be used at the same time")
	}

	if config.StoragePath != "" {
		return storage.NewFile(config.StoragePath, 0o600), nil
	}

	st, err := internal.NewHTTPStorage(config.StorageBaseURL)
	if err != nil {
		return nil, fmt.Errorf("new HTTP storage: %w", err)
	}

	return st, nil
}
