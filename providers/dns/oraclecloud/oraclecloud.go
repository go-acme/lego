package oraclecloud

import (
	"context"
	"crypto/rsa"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	ocicommon "github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/dns"
	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	CompartmentID      string
	OCIConfigProvider  ocicommon.ConfigurationProvider
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("OCI_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("OCI_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("OCI_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("OCI_HTTP_TIMEOUT", 60*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dns.DnsClient
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for OracleCloud.
func NewDNSProvider() (*DNSProvider, error) {
	compartmentID := env.GetOrDefaultString("OCI_COMPARTMENT_OCID", "")
	ociConfigProvider := getOCIConfigProvider()

	config := NewDefaultConfig()
	config.CompartmentID = compartmentID
	config.OCIConfigProvider = ociConfigProvider

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OracleCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("oraclecloud: the configuration of the DNS provider is nil")
	}

	if config.CompartmentID == "" {
		return nil, errors.New("oraclecloud: CompartmentID is missing")
	}

	if config.OCIConfigProvider == nil {
		return nil, errors.New("oraclecloud: OCIConfigProvider is missing")
	}

	if _, err := config.OCIConfigProvider.PrivateRSAKey(); err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if _, err := config.OCIConfigProvider.KeyID(); err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if _, err := config.OCIConfigProvider.TenancyOCID(); err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if _, err := config.OCIConfigProvider.UserOCID(); err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if _, err := config.OCIConfigProvider.KeyFingerprint(); err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if _, err := config.OCIConfigProvider.Region(); err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	client, err := dns.NewDnsClientWithConfigurationProvider(config.OCIConfigProvider)
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{client: &client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// generate request RecordDetails
	recordDetails := dns.RecordDetails{
		Domain:      ocicommon.String(dns01.UnFqdn(fqdn)),
		Rdata:       ocicommon.String(value),
		Rtype:       ocicommon.String("TXT"),
		Ttl:         ocicommon.Int(30),
		IsProtected: ocicommon.Bool(false),
	}

	request := dns.UpdateDomainRecordsRequest{
		ZoneNameOrId: ocicommon.String(domain),
		Domain:       ocicommon.String(dns01.UnFqdn(fqdn)),
		UpdateDomainRecordsDetails: dns.UpdateDomainRecordsDetails{
			Items: []dns.RecordDetails{
				recordDetails,
			},
		},
		CompartmentId: ocicommon.String(d.config.CompartmentID),
	}

	_, err := d.client.UpdateDomainRecords(context.Background(), request)
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	request := dns.DeleteDomainRecordsRequest{
		ZoneNameOrId:  ocicommon.String(domain),
		Domain:        ocicommon.String(dns01.UnFqdn(fqdn)),
		CompartmentId: ocicommon.String(d.config.CompartmentID),
	}

	_, err := d.client.DeleteDomainRecords(context.Background(), request)
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	return nil
}

func getOCIConfigProvider() ocicommon.ConfigurationProvider {
	return ociConfigProvider{}
}

type ociConfigProvider struct {
}

const (
	envPrivKeyEncoded    = "OCI_PRIVKEY_BASE64"
	envPrivKeyPassphrase = "OCI_PRIVKEY_PASS"
	envTenancyID         = "OCI_TENANCY_OCID"
	envUserID            = "OCI_USER_OCID"
	envPubKeyFingerPrint = "OCI_PUBKEY_FINGERPRINT"
	envRegion            = "OCI_REGION"
)

func (p ociConfigProvider) PrivateRSAKey() (key *rsa.PrivateKey, err error) {
	var privateKeyEncoded string
	var privateKeyPassphrase string
	var ok bool

	if privateKeyEncoded, ok = os.LookupEnv(envPrivKeyEncoded); !ok {
		err = fmt.Errorf("can not read PrivateKeyEncoded from environment variable %s", envPrivKeyEncoded)
		return nil, err
	}

	if privateKeyPassphrase, ok = os.LookupEnv(envPrivKeyPassphrase); !ok {
		privateKeyPassphrase = ""
	}

	privateKeyDecoded, _ := b64.StdEncoding.DecodeString(privateKeyEncoded)

	key, err = ocicommon.PrivateKeyFromBytes(privateKeyDecoded, &privateKeyPassphrase)
	return key, nil
}

func (p ociConfigProvider) KeyID() (keyID string, err error) {
	ocid, err := p.TenancyOCID()
	if err != nil {
		return "", err
	}

	userocid, err := p.UserOCID()
	if err != nil {
		return "", err
	}

	fingerprint, err := p.KeyFingerprint()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", ocid, userocid, fingerprint), nil
}

func (p ociConfigProvider) TenancyOCID() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envTenancyID); !ok {
		err = fmt.Errorf("can not read Tenancy from environment variable %s", envTenancyID)
		return "", err
	}

	return value, nil
}

func (p ociConfigProvider) UserOCID() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envUserID); !ok {
		err = fmt.Errorf("can not read user id from environment variable %s", envUserID)
		return "", err
	}

	return value, nil
}

func (p ociConfigProvider) KeyFingerprint() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envPubKeyFingerPrint); !ok {
		err = fmt.Errorf("can not read fingerprint from environment variable %s", envPubKeyFingerPrint)
		return "", err
	}

	return value, nil
}

func (p ociConfigProvider) Region() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envRegion); !ok {
		err = fmt.Errorf("can not read region from environment variable %s", envRegion)
		return "", err
	}

	return value, nil
}
