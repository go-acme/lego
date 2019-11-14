package easydns

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"EASYDNS_ENDPOINT",
	"EASYDNS_TOKEN",
	"EASYDNS_KEY").
	WithDomain("EASYDNS_DOMAIN")

func setup() (*DNSProvider, *http.ServeMux, func()) {
	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	endpoint, err := url.Parse(server.URL)
	if err != nil {
		panic(err)
	}

	config := NewDefaultConfig()
	config.Token = "TOKEN"
	config.Key = "SECRET"
	config.Endpoint = endpoint

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
			expected: "easydns: some credentials information are missing: EASYDNS_TOKEN",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				"EASYDNS_TOKEN": "TOKEN",
			},
			expected: "easydns: some credentials information are missing: EASYDNS_KEY",
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
			expected: "easydns: the API token is missing",
		},
		{
			desc: "missing key",
			config: &Config{
				Token: "TOKEN",
			},
			expected: "easydns: the API key is missing",
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
	provider, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/zones/records/add/example.com/TXT", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get("Authorization"), "Authorization")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		expectedReqBody := `{"domain":"example.com","host":"_acme-challenge","ttl":"120","prio":"0","type":"TXT","rdata":"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"}
`
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
	require.Contains(t, provider.recordIDs, "_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM")
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

	provider.recordIDs["_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"] = "123456"
	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_WhenHttpError_ReturnsError(t *testing.T) {
	provider, mux, tearDown := setup()
	defer tearDown()

	errorMessage := `{
		"error": {
			"code": 406,
			"message": "Provided id is invalid or you do not have permission to access it."
		}
	}`
	mux.HandleFunc("/zones/records/example.com/123456", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get("Authorization"), "Authorization")

		w.WriteHeader(http.StatusNotAcceptable)
		_, err := fmt.Fprint(w, errorMessage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	provider.recordIDs["_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"] = "123456"
	err := provider.CleanUp("example.com", "token", "keyAuth")
	expectedError := fmt.Sprintf("easydns: 406: request failed: %v", errorMessage)
	require.EqualError(t, err, expectedError)
}

func TestSplitFqdn(t *testing.T) {
	testCases := []struct {
		desc           string
		fqdn           string
		expectedHost   string
		expectedDomain string
	}{
		{
			desc:           "domain only",
			fqdn:           "domain.com.",
			expectedHost:   "",
			expectedDomain: "domain.com",
		},
		{
			desc:           "single-part host",
			fqdn:           "_acme-challenge.domain.com.",
			expectedHost:   "_acme-challenge",
			expectedDomain: "domain.com",
		},
		{
			desc:           "multi-part host",
			fqdn:           "_acme-challenge.sub.domain.com.",
			expectedHost:   "_acme-challenge.sub",
			expectedDomain: "domain.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actualHost, actualDomain := splitFqdn(test.fqdn)

			require.Equal(t, test.expectedHost, actualHost)
			require.Equal(t, test.expectedDomain, actualDomain)
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
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
