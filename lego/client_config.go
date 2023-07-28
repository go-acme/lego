package lego

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/registration"
)

const (
	// caCertificatesEnvVar is the environment variable name that can be used to
	// specify the path to PEM encoded CA Certificates that can be used to
	// authenticate an ACME server with an HTTPS certificate not issued by a CA in
	// the system-wide trusted root list.
	// Multiple file paths can be added by using os.PathListSeparator as a separator.
	caCertificatesEnvVar = "LEGO_CA_CERTIFICATES"

	// caSystemCertPool is the environment variable name that can be used to define
	// if the certificates pool must use a copy of the system cert pool.
	caSystemCertPool = "LEGO_CA_SYSTEM_CERT_POOL"

	// caServerNameEnvVar is the environment variable name that can be used to
	// specify the CA server name that can be used to
	// authenticate an ACME server with an HTTPS certificate not issued by a CA in
	// the system-wide trusted root list.
	caServerNameEnvVar = "LEGO_CA_SERVER_NAME"

	// LEDirectoryProduction URL to the Let's Encrypt production.
	LEDirectoryProduction = "https://acme-v02.api.letsencrypt.org/directory"

	// LEDirectoryStaging URL to the Let's Encrypt staging.
	LEDirectoryStaging = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

type Config struct {
	CADirURL    string
	User        registration.User
	UserAgent   string
	HTTPClient  *http.Client
	Certificate CertificateConfig
}

func NewConfig(user registration.User) *Config {
	return &Config{
		CADirURL:   LEDirectoryProduction,
		User:       user,
		HTTPClient: createDefaultHTTPClient(),
		Certificate: CertificateConfig{
			KeyType: certcrypto.RSA2048,
			Timeout: 30 * time.Second,
		},
	}
}

type CertificateConfig struct {
	KeyType certcrypto.KeyType
	Timeout time.Duration
}

// createDefaultHTTPClient Creates an HTTP client with a reasonable timeout value
// and potentially a custom *x509.CertPool
// based on the caCertificatesEnvVar environment variable (see the `initCertPool` function).
func createDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			TLSClientConfig: &tls.Config{
				ServerName: os.Getenv(caServerNameEnvVar),
				RootCAs:    initCertPool(),
			},
		},
	}
}

// initCertPool creates a *x509.CertPool populated with the PEM certificates
// found in the filepath specified in the caCertificatesEnvVar OS environment variable.
// If the caCertificatesEnvVar is not set then initCertPool will return nil.
// If there is an error creating a *x509.CertPool from the provided caCertificatesEnvVar value then initCertPool will panic.
// If the caSystemCertPool is set to a "truthy value" (`1`, `t`, `T`, `TRUE`, `true`, `True`) then a copy of system cert pool will be used.
// caSystemCertPool requires caCertificatesEnvVar to be set.
func initCertPool() *x509.CertPool {
	customCACertsPath := os.Getenv(caCertificatesEnvVar)
	if customCACertsPath == "" {
		return nil
	}

	certPool := getCertPool()

	for _, customPath := range strings.Split(customCACertsPath, string(os.PathListSeparator)) {
		customCAs, err := os.ReadFile(customPath)
		if err != nil {
			panic(fmt.Sprintf("error reading %s=%q: %v",
				caCertificatesEnvVar, customPath, err))
		}

		if ok := certPool.AppendCertsFromPEM(customCAs); !ok {
			panic(fmt.Sprintf("error creating x509 cert pool from %s=%q: %v",
				caCertificatesEnvVar, customPath, err))
		}
	}

	return certPool
}

func getCertPool() *x509.CertPool {
	useSystemCertPool, _ := strconv.ParseBool(os.Getenv(caSystemCertPool))
	if !useSystemCertPool {
		return x509.NewCertPool()
	}

	pool, err := x509.SystemCertPool()
	if err == nil {
		return pool
	}
	return x509.NewCertPool()
}
