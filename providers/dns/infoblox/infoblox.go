// Package infoblox implements a DNS provider for solving the DNS-01 challenge using on prem infoblox DNS.
package infoblox

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	ibclient "github.com/infobloxopen/infoblox-go-client"
)

const (
	// DefaultIBClientWorkers is the number of workers the infoblox client uses for connections.
	DefaultIBClientWorkers = 1
	// DefaultHTTPRequestTimeout is the http timeout used by the infoblox client.
	DefaultHTTPRequestTimeout = 30
	// DefaultUserAgent is the user agent this package will use.
	DefaultUserAgent = "lego-infoblox-provider-v1.0.0"
)

const (
	envNamespace = "INFOBLOX_"

	EnvUser        = envNamespace + "USER"
	EnvHost        = envNamespace + "HOST"
	EnvPassword    = envNamespace + "PASSWORD"
	EnvTTL         = envNamespace + "TTL"
	EnvView        = envNamespace + "View"
	EnvWapiVersion = envNamespace + "WAPI_VERSION"
	EnvPort        = envNamespace + "PORT"
	SSL_VERIFY     = "SSL_VERIFY"
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
func BuildConfigFromEnv() (Config, error) {
	host := env.GetOrFile(EnvHost)
	if host == "" {
		return Config{}, fmt.Errorf("infoblox build config from env could not find value for " + EnvHost)
	}
	user := env.GetOrFile(EnvUser)
	if user == "" {
		return Config{}, fmt.Errorf("infoblox build config from env could not find value for " + EnvUser)
	}
	password := env.GetOrFile(EnvPassword)
	if password == "" {
		return Config{}, fmt.Errorf("infoblox build config from env could not find value for " + EnvPassword)
	}

	return Config{
		Host:        host,
		Username:    user,
		Password:    password,
		TTL:         env.GetOrDefaultInt(EnvTTL, 120),
		DNSView:     env.GetOrDefaultString(EnvView, "External"),
		WapiVersion: env.GetOrDefaultString(EnvWapiVersion, "2.11"),
		Port:        env.GetOrDefaultString(EnvPort, "443"),
		SSLVerify:   env.GetOrDefaultBool(SSL_VERIFY, true),
	}, nil
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	Config Config
}

// NewDNSProvider returns a DNSProvider instance configured for Infoblox. Primarily used by the CLI.
// See infoblox.toml for more information.
func NewDNSProvider() (*DNSProvider, error) {
	config, err := BuildConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("infoblox new dns provider could not get config from env: %v", err)
	}

	return &DNSProvider{
		Config: config,
	}, nil
}

// buildConnection passes the config to the infoblox client go library and returns an authenticated connection to the host specified in the config. Be sure to close the connection when done.
func (d *DNSProvider) buildConnection() (*ibclient.Connector, error) {
	ibLibraryConfig := ibclient.HostConfig{
		Host:     d.Config.Host,
		Version:  d.Config.WapiVersion,
		Port:     d.Config.Port,
		Username: d.Config.Username,
		Password: d.Config.Password,
	}

	transportConfig := ibclient.NewTransportConfig(strconv.FormatBool(d.Config.SSLVerify), DefaultHTTPRequestTimeout, DefaultIBClientWorkers)
	requestBuilder := &ibclient.WapiRequestBuilder{}
	requestor := &ibclient.WapiHttpRequestor{}

	conn, err := ibclient.NewConnector(ibLibraryConfig, transportConfig, requestBuilder, requestor)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	fqdn = strings.TrimSuffix(fqdn, ".")

	conn, err := d.buildConnection()
	defer conn.Logout()

	if err != nil {
		return fmt.Errorf("infoblox present could not connect to infoblox with supplied config %v", err)
	}

	objMgr := ibclient.NewObjectManager(conn, DefaultUserAgent, "")

	_, err = objMgr.CreateTXTRecord(
		fqdn,
		value,
		uint(d.Config.TTL),
		d.Config.DNSView,
	)
	if err != nil {
		return fmt.Errorf("infoblox present could not manage txt record creation for %s: %v", domain, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	fqdn = strings.TrimSuffix(fqdn, ".")

	conn, err := d.buildConnection()
	defer conn.Logout()

	if err != nil {
		return fmt.Errorf("infoblox cleanup could not connect to infoblox with supplied config %v", err)
	}

	objMgr := ibclient.NewObjectManager(conn, DefaultUserAgent, "")

	record, err := objMgr.GetTXTRecord(fqdn)
	if err != nil {
		return fmt.Errorf("infoblox cleanup could not find txt record to delete %s: %v", domain, err)
	}

	_, err = objMgr.DeleteTXTRecord(record.Ref)
	if err != nil {
		return fmt.Errorf("infoblox cleanup could not delete txt record %s: %v", domain, err)
	}

	return nil
}
