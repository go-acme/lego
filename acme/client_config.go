package acme

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/xenolf/lego/certcrypto"
	"github.com/xenolf/lego/registration"
)

const (
	// caCertificatesEnvVar is the environment variable name that can be used to
	// specify the path to PEM encoded CA Certificates that can be used to
	// authenticate an ACME server with a HTTPS certificate not issued by a CA in
	// the system-wide trusted root list.
	caCertificatesEnvVar = "LEGO_CA_CERTIFICATES"

	// caServerNameEnvVar is the environment variable name that can be used to
	// specify the CA server name that can be used to
	// authenticate an ACME server with a HTTPS certificate not issued by a CA in
	// the system-wide trusted root list.
	caServerNameEnvVar = "LEGO_CA_SERVER_NAME"

	// LEDirectoryProduction URL to the Let's Encrypt production
	LEDirectoryProduction = "https://acme-v02.api.letsencrypt.org/directory"

	// LEDirectoryStaging URL to the Let's Encrypt staging
	LEDirectoryStaging = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

type Config struct {
	caDirURL   string
	user       registration.User
	keyType    certcrypto.KeyType
	userAgent  string
	HTTPClient *http.Client
}

type ConfigOption func(*Config) error

func NewConfig(user registration.User, opts ...ConfigOption) (*Config, error) {
	config := &Config{
		caDirURL:   LEDirectoryProduction,
		user:       user,
		keyType:    certcrypto.RSA2048,
		HTTPClient: createDefaultHTTPClient(),
	}

	for _, opt := range opts {
		err := opt(config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func WithCADirURL(caDirURL string) ConfigOption {
	return func(c *Config) error {
		_, err := url.Parse(caDirURL)
		if err != nil {
			return err
		}

		c.caDirURL = caDirURL
		return nil
	}
}

func WithUser(user registration.User) ConfigOption {
	return func(c *Config) error {
		c.user = user
		return nil
	}
}

func WithUserAgent(userAgent string) ConfigOption {
	return func(c *Config) error {
		c.userAgent = userAgent
		return nil
	}
}

func WithKeyType(keyType certcrypto.KeyType) ConfigOption {
	return func(c *Config) error {
		c.keyType = keyType
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) ConfigOption {
	return func(c *Config) error {
		if httpClient == nil {
			return errors.New("the HTTP client cannot be nil")
		}

		c.HTTPClient = httpClient
		return nil
	}
}

// createDefaultHTTPClient Creates an HTTP client with a reasonable timeout value
// and potentially a custom *x509.CertPool
// based on the caCertificatesEnvVar environment variable (see the `initCertPool` function)
func createDefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				ServerName: os.Getenv(caServerNameEnvVar),
				RootCAs:    initCertPool(),
			},
		},
	}
}

// initCertPool creates a *x509.CertPool populated with the PEM certificates
// found in the filepath specified in the caCertificatesEnvVar OS environment
// variable. If the caCertificatesEnvVar is not set then initCertPool will
// return nil. If there is an error creating a *x509.CertPool from the provided
// caCertificatesEnvVar value then initCertPool will panic.
func initCertPool() *x509.CertPool {
	if customCACertsPath := os.Getenv(caCertificatesEnvVar); customCACertsPath != "" {
		customCAs, err := ioutil.ReadFile(customCACertsPath)
		if err != nil {
			panic(fmt.Sprintf("error reading %s=%q: %v",
				caCertificatesEnvVar, customCACertsPath, err))
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(customCAs); !ok {
			panic(fmt.Sprintf("error creating x509 cert pool from %s=%q: %v",
				caCertificatesEnvVar, customCACertsPath, err))
		}
		return certPool
	}
	return nil
}
