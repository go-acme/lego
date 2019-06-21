// Package vegadns implements a DNS provider for solving the DNS-01
// challenge using VegaDNS.
package versio

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDomain = "example.com"

const ipPort = "127.0.0.1:2112"

var envTest = tester.NewEnvTest("VERSIO_USERNAME", "VERSIO_PASSWORD", "VERSIO_ENDPOINT")

type muxCallback func() *http.ServeMux

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"VERSIO_USERNAME": "me@example.com",
				"VERSIO_PASSWORD": "SECRET",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				"VERSIO_PASSWORD": "me@example.com",
			},
			expected: "versio: some credentials information are missing: VERSIO_USERNAME",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				"VERSIO_USERNAME": "TOKEN",
			},
			expected: "versio: some credentials information are missing: VERSIO_PASSWORD",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		config   *Config
		expected string
	}{
		{
			desc: "success",
			config: &Config{
				Username: "me@example.com",
				Password: "PW",
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "versio: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing username",
			config: &Config{
				Password: "PW",
			},
			expected: "versio: the versio username is missing",
		},
		{
			desc: "missing password",
			config: &Config{
				Username: "UN",
			},
			expected: "versio: the versio password is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_TimeoutSuccess(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()
	envVars := map[string]string{
		"VERSIO_USERNAME": "me@example.com",
		"VERSIO_PASSWORD": "secret",
		"VERSIO_ENDPOINT": "http://127.0.0.1:2112",
	}

	ts, err := startTestServer(muxSuccess)
	require.NoError(t, err)

	defer ts.Close()

	envTest.Apply(envVars)
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	timeout, interval := provider.Timeout()
	assert.Equal(t, timeout, time.Duration(60000000000))
	assert.Equal(t, interval, time.Duration(5000000000))
}

func TestDNSProvider_SequentialSuccess(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()
	envVars := map[string]string{
		"VERSIO_USERNAME": "me@example.com",
		"VERSIO_PASSWORD": "secret",
		"VERSIO_ENDPOINT": "http://127.0.0.1:2112",
	}

	ts, err := startTestServer(muxSuccess)
	require.NoError(t, err)

	defer ts.Close()

	envTest.Apply(envVars)
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	sequential := provider.Sequential()
	assert.Equal(t, sequential, time.Duration(60000000000))
}

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		desc          string
		callback      muxCallback
		expectedError string
	}{
		{
			desc:     "Success",
			callback: muxSuccess,
		},
		{
			desc:          "FailToFindZone",
			callback:      muxFailToFindZone,
			expectedError: "versio: 401: request failed: ObjectDoesNotExist|Domain not found",
		},
		{
			desc:          "FailToCreateTXT",
			callback:      muxFailToCreateTXT,
			expectedError: "versio: 400: request failed: ProcessError|DNS record invalid type _acme-challenge.fjmk.eu. TST",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envVars := map[string]string{
				"VERSIO_USERNAME": "me@example.com",
				"VERSIO_PASSWORD": "secret",
				"VERSIO_ENDPOINT": "http://127.0.0.1:2112",
			}

			ts, err := startTestServer(test.callback)
			require.NoError(t, err)

			defer ts.Close()

			envTest.Apply(envVars)
			provider, err := NewDNSProvider()
			require.NoError(t, err)

			err = provider.Present(testDomain, "token", "keyAuth")
			if len(test.expectedError) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	testCases := []struct {
		desc          string
		callback      muxCallback
		expectedError string
	}{
		{
			desc:     "Success",
			callback: muxSuccess,
		},
		{
			desc:          "FailToFindZone",
			callback:      muxFailToFindZone,
			expectedError: "versio: 401: request failed: ObjectDoesNotExist|Domain not found",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envVars := map[string]string{
				"VERSIO_USERNAME": "me@example.com",
				"VERSIO_PASSWORD": "secret",
				"VERSIO_ENDPOINT": "http://127.0.0.1:2112",
			}

			ts, err := startTestServer(test.callback)
			require.NoError(t, err)

			defer ts.Close()

			envTest.Apply(envVars)
			provider, err := NewDNSProvider()
			require.NoError(t, err)

			err = provider.CleanUp(testDomain, "token", "keyAuth")
			if len(test.expectedError) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func muxSuccess() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("show_dns_records") == "true" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/domains/example.com/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Not Found for Request: (%+v)\n\n", r)
	})

	return mux
}

func muxFailToFindZone() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/domains/example.com", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, tokenFailToFindZoneMock)
	})

	return mux
}

func muxFailToCreateTXT() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("show_dns_records") == "true" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/domains/example.com/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, tokenFailToCreateTXTMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	return mux
}

// Starts and returns a test server using a custom ip/port. Defer close() afterwards.
func startTestServer(callback muxCallback) (*httptest.Server, error) {
	err := os.Setenv("SECRET_VEGADNS_KEY", "key")
	if err != nil {
		return nil, err
	}

	err = os.Setenv("SECRET_VEGADNS_SECRET", "secret")
	if err != nil {
		return nil, err
	}

	err = os.Setenv("VEGADNS_URL", "http://"+ipPort)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewUnstartedServer(callback())

	l, err := net.Listen("tcp", ipPort)
	if err != nil {
		return nil, err
	}

	ts.Listener = l
	ts.Start()

	return ts, nil
}
