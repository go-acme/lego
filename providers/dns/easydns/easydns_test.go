package easydns

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"EASYDNS_HOSTNAME",
	"EASYDNS_TOKEN",
	"EASYDNS_KEY",
	"EASYDNS_PROPAGATION_TIMEOUT",
	"EASYDNS_POLLING_INTERVAL").
	WithDomain("EASYDNS_DOMAIN")

func setup() (*DNSProvider, *http.ServeMux, func()) {
	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	url, err := url.Parse(server.URL)
	if err != nil {
		panic(err)
	}

	config := NewDefaultConfig()
	config.Token = "TOKEN"
	config.Key = "SECRET"
	config.URL = url

	provider, err := NewDNSProviderConfig(config)
	if err != nil {
		panic(err)
	}

	return provider, handler, server.Close
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"EASYDNS_TOKEN": "TOKEN",
				"EASYDNS_KEY":   "SECRET",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				"EASYDNS_KEY": "SECRET",
			},
			expected: "easydns: the API token is missing: EASYDNS_TOKEN",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				"EASYDNS_TOKEN": "TOKEN",
			},
			expected: "easydns: the API key is missing: EASYDNS_KEY",
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
				Token: "TOKEN",
				Key:   "KEY",
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "easydns: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing token",
			config: &Config{
				Key: "KEY",
			},
			expected: "easydns: the API token is missing: EASYDNS_TOKEN",
		},
		{
			desc: "missing key",
			config: &Config{
				Token: "TOKEN",
			},
			expected: "easydns: the API key is missing: EASYDNS_KEY",
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

func TestDNSProvider_Present_WhenTxtRecordNotFound_CreatesTxtRecord(t *testing.T) {
	provider, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/zones/records/all/example.com", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get("Authorization"), "Authorization")

		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, `{
			"data": [],
			"status": 200
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/zones/records/add/example.com/TXT", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get("Authorization"), "Authorization")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		expectedReqBody := "{\"domain\":\"example.com\",\"host\":\"_acme-challenge\",\"ttl\":\"120\",\"prio\":\"0\",\"type\":\"TXT\",\"rdata\":\"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM\"}\n"
		assert.Equal(t, expectedReqBody, string(reqBody))

		w.WriteHeader(http.StatusCreated)
		_, err = fmt.Fprintf(w, `{
			"msg": "OK",
			"tm": 1554681934,
			"data": {
				"host": "_acme-challenge",
				"geozone_id": 0,
				"ttl": "120",
				"prio": "0",
				"rdata": "pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM",
				"revoked": 0,
				"id": "123456789",
				"new_host": "_acme-challenge.example.com"
			},
			"status": 201
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	err := provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Present_WhenTxtRecordFound_UpdatesTxtRecord(t *testing.T) {
	provider, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/zones/records/all/example.com", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get("Authorization"), "Authorization")

		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, `{
			"data": [
				{
					"id": "123456",
					"domain": "example.com",
					"host": "_acme-challenge",
					"ttl": "300",
					"prio": "0",
					"type": "TXT",
					"rdata": "",
					"geozone_id": "0",
					"last_mod": "2019-04-08 00:59:37"
				}
			],
			"status": 200
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/zones/records/123456", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get("Authorization"), "Authorization")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		expectedReqBody := "{\"domain\":\"example.com\",\"host\":\"_acme-challenge\",\"ttl\":\"120\",\"prio\":\"0\",\"type\":\"TXT\",\"rdata\":\"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM\"}\n"
		assert.Equal(t, expectedReqBody, string(reqBody))

		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, `{
			"msg": "Record updated successfully.",
			"tm": 1554685520,
			"data": {
				"id": "123456",
				"domain": "example.com",
				"host": "_acme-challenge-test",
				"ttl": "300",
				"prio": "0",
				"type": "TXT",
				"rdata": "update",
				"geozone_id": "0",
				"last_mod": "2019-04-08 01:05:20"
			},
			"status": 200
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	err := provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Present_WhenHttpError_ReturnsError(t *testing.T) {
	provider, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/zones/records/all/example.com", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")

		w.WriteHeader(http.StatusForbidden)
		_, err := fmt.Fprintf(w, `{
			"error": {
				"code": 403,
				"message": "Access to resource denied due to permissions"
			}
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	err := provider.Present("example.com", "token", "keyAuth")
	require.Error(t, err)
}

func TestDNSProvider_Present_Live(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_WhenRecordIdNotSet_NoOp(t *testing.T) {
	provider, _, tearDown := setup()
	defer tearDown()

	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_WhenRecordIdSet_DeletesTxtRecord(t *testing.T) {
	provider, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/zones/records/example.com/123456", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get("Authorization"), "Authorization")

		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, `{
			"msg": "OK",
			"data": {
				"domain": "example.com",
				"id": "123456"
			},
			"status": 200
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	provider.recordID = "123456"
	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_Live(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Timeout(t *testing.T) {
	testCases := []struct {
		desc                       string
		envVars                    map[string]string
		expectedPropagationTimeout time.Duration
		expectedPollingInterval    time.Duration
	}{
		{
			desc: "defaults",
			envVars: map[string]string{
				"EASYDNS_TOKEN": "TOKEN",
				"EASYDNS_KEY":   "SECRET",
			},
			expectedPropagationTimeout: dns01.DefaultPropagationTimeout,
			expectedPollingInterval:    dns01.DefaultPollingInterval,
		},
		{
			desc: "custom propagation timeout",
			envVars: map[string]string{
				"EASYDNS_TOKEN":               "TOKEN",
				"EASYDNS_KEY":                 "SECRET",
				"EASYDNS_PROPAGATION_TIMEOUT": "1",
			},
			expectedPropagationTimeout: 1000000000,
			expectedPollingInterval:    dns01.DefaultPollingInterval,
		},
		{
			desc: "custom polling interval",
			envVars: map[string]string{
				"EASYDNS_TOKEN":            "TOKEN",
				"EASYDNS_KEY":              "SECRET",
				"EASYDNS_POLLING_INTERVAL": "1",
			},
			expectedPropagationTimeout: dns01.DefaultPropagationTimeout,
			expectedPollingInterval:    1000000000,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()
			actualPropagationTimeout, actualPollingInterval := p.Timeout()

			require.NoError(t, err)
			require.NotNil(t, p)
			require.Equal(t, test.expectedPropagationTimeout, actualPropagationTimeout)
			require.Equal(t, test.expectedPollingInterval, actualPollingInterval)
		})
	}
}

func TestSplitFqdn(t *testing.T) {
	testCases := []struct {
		desc           string
		fqdn           string
		expectedHost   string
		expectedDoamin string
	}{
		{
			desc:           "domain only",
			fqdn:           "domain.com.",
			expectedHost:   "",
			expectedDoamin: "domain.com",
		},
		{
			desc:           "single-part host",
			fqdn:           "_acme-challenge.domain.com.",
			expectedHost:   "_acme-challenge",
			expectedDoamin: "domain.com",
		},
		{
			desc:           "multi-part host",
			fqdn:           "_acme-challenge.sub.domain.com.",
			expectedHost:   "_acme-challenge.sub",
			expectedDoamin: "domain.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actualHost, actualDomain := splitFqdn(test.fqdn)

			require.Equal(t, test.expectedHost, actualHost)
			require.Equal(t, test.expectedDoamin, actualDomain)
		})
	}
}
