// Package infoblox implements a DNS provider for solving the DNS-01 challenge using on prem infoblox DNS.
package infoblox

import (
	"strconv"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/pkg/errors"

	ibclient "github.com/infobloxopen/infoblox-go-client"
)

const (
	DefaultIBClientWorkeres   = 10
	DefaultHttpRequestTimeout = 30
	DefaultUserAgent          = "lego-infoblox-provider-v1.0.0"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	// Host is the URL of the infoblox grid manager
	Host string
	// Port is the Port for the grid manager
	Port string
	// Username the user for accessing infoblox
	Username string
	// Password the password for accessing infoblox
	Password string
	// DNSView is the dns view to put new records and search from
	DNSView string
	// WapiVersion is the version of web api used
	WapiVersion string
	// SSLVerify is whether or not to verify the ssl of the infoblox server being hit
	SSLVerify bool

	// TTL for the created records
	TTL int
}

// BuildConfigFromEnv build a config object, getting all values from the environment.
func BuildConfigFromEnv() Config {
	return Config{
		Host:        env.GetOrFile("INFOBLOX_USER"),
		Username:    env.GetOrFile("INFOBLOX_USER"),
		Password:    env.GetOrFile("INFOBLOX_PASSWORD"),
		TTL:         env.GetOrDefaultInt("INFOBLOX_TTL", 120),
		DNSView:     env.GetOrDefaultString("INFOBLOX_VIEW", "External"),
		WapiVersion: env.GetOrDefaultString("INFOBLOX_WAPI_VERSION", "2.11"),
		Port:        env.GetOrDefaultString("INFOBLOX_PORT", "443"),
		SSLVerify:   env.GetOrDefaultBool("SSL_VERIFY", true),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config     Config
	connection *ibclient.Connector
}

// NewDNSProvider returns a DNSProvider instance configured for Infoblox.
// See infoblox.toml for more information
func NewDNSProvider() (*DNSProvider, error) {
	config := BuildConfigFromEnv()
	ibLibraryConfig := ibclient.HostConfig{
		Host:     config.Host,
		Version:  config.WapiVersion,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
	}

	transportConfig := ibclient.NewTransportConfig(strconv.FormatBool(config.SSLVerify), DefaultHttpRequestTimeout, DefaultIBClientWorkeres)
	requestBuilder := &ibclient.WapiRequestBuilder{}
	requestor := &ibclient.WapiHttpRequestor{}

	conn, err := ibclient.NewConnector(ibLibraryConfig, transportConfig, requestBuilder, requestor)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		config:     config,
		connection: conn,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return errors.Wrapf(err, "infoblox cleanup trying to find authzone for %s", fqdn)
	}

	objMgr := ibclient.NewObjectManager(d.connection, DefaultUserAgent, "")

	_, err = objMgr.CreateTXTRecord(
		authZone,
		value,
		uint(d.config.TTL),
		d.config.DNSView,
	)
	if err != nil {
		return errors.Wrapf(err, "infoblox present could not manage txt record creation for %s", domain)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return errors.Wrapf(err, "infoblox cleanup trying to find authzone for %s", fqdn)
	}

	objMgr := ibclient.NewObjectManager(d.connection, DefaultUserAgent, "")

	record, err := objMgr.GetTXTRecord(authZone)
	if err != nil {
		return errors.Wrapf(err, "infoblox cleanup could not find txt record to delete %s", domain)
	}

	_, err = objMgr.DeleteTXTRecord(record.Ref)
	if err != nil {
		return errors.Wrapf(err, "infoblox cleanup could not manage txt record delete %s", domain)
	}

	return nil
}
