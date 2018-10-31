package emca

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/xenolf/lego/emca/certificate"
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
)

type Config struct {
	caDirURL   string
	user       User
	keyType    certificate.KeyType
	userAgent  string
	HTTPClient *http.Client
}

func NewDefaultConfig(user User) *Config {
	return &Config{
		caDirURL:   "https://acme-v02.api.letsencrypt.org/directory",
		user:       user,
		keyType:    certificate.RSA2048,
		HTTPClient: createDefaultHTTPClient(),
	}
}

func (c *Config) WithCADirURL(caDirURL string) *Config {
	c.caDirURL = caDirURL
	return c
}

func (c *Config) WithUser(user User) *Config {
	c.user = user
	return c
}

func (c *Config) WithUserAgent(userAgent string) *Config {
	c.userAgent = userAgent
	return c
}

func (c *Config) WithKeyType(keyType certificate.KeyType) *Config {
	c.keyType = keyType
	return c
}

func (c *Config) WithHTTPClient(httpClient *http.Client) *Config {
	c.HTTPClient = httpClient
	return c
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
