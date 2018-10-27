package acme

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	// UserAgent (if non-empty) will be tacked onto the User-Agent string in requests.
	UserAgent string

	// HTTPClient is an HTTP client with a reasonable timeout value and
	// potentially a custom *x509.CertPool based on the caCertificatesEnvVar
	// environment variable (see the `initCertPool` function)
	HTTPClient = http.Client{
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
)

const (
	// ourUserAgent is the User-Agent of this underlying library package.
	// NOTE: Update this with each tagged release.
	ourUserAgent = "xenolf-acme/1.2.1"

	// ourUserAgentComment is part of the UA comment linked to the version status of this underlying library package.
	// values: detach|release
	// NOTE: Update this with each tagged release.
	ourUserAgentComment = "detach"

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

type requestOption func(*http.Request) error

func contentType(ct string) requestOption {
	return func(req *http.Request) error {
		req.Header.Set("Content-Type", ct)
		return nil
	}
}

func newRequest(method, uri string, body io.Reader, opts ...requestOption) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", userAgent())

	for _, opt := range opts {
		err = opt(req)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

// httpHead performs a HEAD request with a proper User-Agent string.
// The response body (resp.Body) is already closed when this function returns.
func httpHead(url string) (resp *http.Response, err error) {
	req, err := newRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err = HTTPClient.Do(req)
	if err != nil {
		return resp, fmt.Errorf("failed to do head %q: %v", url, err)
	}

	resp.Body.Close()

	return resp, nil
}

// httpPost performs a POST request with a proper User-Agent string.
// Callers should close resp.Body when done reading from it.
func httpPost(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	req, err := newRequest(http.MethodPost, url, body, contentType(bodyType))
	if err != nil {
		return nil, err
	}

	return HTTPClient.Do(req)
}

// httpGet performs a GET request with a proper User-Agent string.
// Callers should close resp.Body when done reading from it.
func httpGet(url string) (resp *http.Response, err error) {
	req, err := newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return HTTPClient.Do(req)
}

// getJSON performs an HTTP GET request and parses the response body
// as JSON, into the provided respBody object.
func getJSON(uri string, respBody interface{}) (http.Header, error) {
	resp, err := httpGet(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get json %q: %v", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return resp.Header, handleHTTPError(resp)
	}

	return resp.Header, json.NewDecoder(resp.Body).Decode(respBody)
}

// userAgent builds and returns the User-Agent string to use in requests.
func userAgent() string {
	ua := fmt.Sprintf("%s %s (%s; %s; %s)", UserAgent, ourUserAgent, ourUserAgentComment, runtime.GOOS, runtime.GOARCH)
	return strings.TrimSpace(ua)
}
