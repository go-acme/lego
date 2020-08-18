package versio

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDomain = "example.com"

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword, EnvEndpoint).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "me@example.com",
				EnvPassword: "SECRET",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvPassword: "me@example.com",
			},
			expected: "versio: some credentials information are missing: VERSIO_USERNAME",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				EnvUsername: "TOKEN",
			},
			expected: "versio: some credentials information are missing: VERSIO_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "versio: some credentials information are missing: VERSIO_USERNAME,VERSIO_PASSWORD",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider(nil)

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

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		desc          string
		handler       http.Handler
		expectedError string
	}{
		{
			desc:    "Success",
			handler: muxSuccess(),
		},
		{
			desc:          "FailToFindZone",
			handler:       muxFailToFindZone(),
			expectedError: `versio: 401: request failed: ObjectDoesNotExist|Domain not found`,
		},
		{
			desc:          "FailToCreateTXT",
			handler:       muxFailToCreateTXT(),
			expectedError: `versio: 400: request failed: ProcessError|DNS record invalid type _acme-challenge.example.eu. TST`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			baseURL, tearDown := startTestServer(test.handler)
			defer tearDown()

			envTest.Apply(map[string]string{
				EnvUsername: "me@example.com",
				EnvPassword: "secret",
				EnvEndpoint: baseURL,
			})
			provider, err := NewDNSProvider(nil)
			require.NoError(t, err)

			err = provider.Present(testDomain, "token", "keyAuth")
			if len(test.expectedError) == 0 {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	testCases := []struct {
		desc          string
		handler       http.Handler
		expectedError string
	}{
		{
			desc:    "Success",
			handler: muxSuccess(),
		},
		{
			desc:          "FailToFindZone",
			handler:       muxFailToFindZone(),
			expectedError: `versio: 401: request failed: ObjectDoesNotExist|Domain not found`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			baseURL, tearDown := startTestServer(test.handler)
			defer tearDown()

			envTest.Apply(map[string]string{
				EnvUsername: "me@example.com",
				EnvPassword: "secret",
				EnvEndpoint: baseURL,
			})

			provider, err := NewDNSProvider(nil)
			require.NoError(t, err)

			err = provider.CleanUp(testDomain, "token", "keyAuth")
			if len(test.expectedError) == 0 {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func muxSuccess() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("show_dns_records") == "true" {
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/domains/example.com/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Not Found for Request: (%+v)\n\n", r)
		http.NotFound(w, r)
	})

	return mux
}

func muxFailToFindZone() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/domains/example.com", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, tokenFailToFindZoneMock, http.StatusUnauthorized)
	})

	return mux
}

func muxFailToCreateTXT() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("show_dns_records") == "true" {
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/domains/example.com/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			http.Error(w, tokenFailToCreateTXTMock, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	return mux
}

func startTestServer(handler http.Handler) (string, func()) {
	ts := httptest.NewServer(handler)
	return ts.URL, func() {
		ts.Close()
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
