package easydns

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authorizationHeader = "Authorization"

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvEndpoint,
	EnvToken,
	EnvKey).
	WithDomain(envDomain)

func setupTest(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	endpoint, err := url.Parse(server.URL)
	require.NoError(t, err)

	config := NewDefaultConfig()
	config.Token = "TOKEN"
	config.Key = "SECRET"
	config.Endpoint = endpoint
	config.HTTPClient = server.Client()

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	return provider, mux
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
				EnvToken: "TOKEN",
				EnvKey:   "SECRET",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvKey: "SECRET",
			},
			expected: "easydns: some credentials information are missing: EASYDNS_TOKEN",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				EnvToken: "TOKEN",
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

			if test.expected == "" {
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

			if test.expected == "" {
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
	provider, mux := setupTest(t)

	mux.HandleFunc("/zones/records/add/example.com/TXT", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get(authorizationHeader), authorizationHeader)

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
			return
		}
	})

	err := provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)
	require.Contains(t, provider.recordIDs, "_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM")
}

func TestDNSProvider_Cleanup_WhenRecordIdNotSet_NoOp(t *testing.T) {
	provider, _ := setupTest(t)

	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_WhenRecordIdSet_DeletesTxtRecord(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/zones/records/example.com/123456", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get(authorizationHeader), authorizationHeader)

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
			return
		}
	})

	provider.recordIDs["_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"] = "123456"
	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_WhenHttpError_ReturnsError(t *testing.T) {
	provider, mux := setupTest(t)

	errorMessage := `{
		"error": {
			"code": 406,
			"message": "Provided id is invalid or you do not have permission to access it."
		}
	}`
	mux.HandleFunc("/zones/records/example.com/123456", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "method")
		assert.Equal(t, "format=json", r.URL.RawQuery, "query")
		assert.Equal(t, "Basic VE9LRU46U0VDUkVU", r.Header.Get(authorizationHeader), authorizationHeader)

		w.WriteHeader(http.StatusNotAcceptable)
		_, err := fmt.Fprint(w, errorMessage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	provider.recordIDs["_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"] = "123456"
	err := provider.CleanUp("example.com", "token", "keyAuth")
	expectedError := fmt.Sprintf("easydns: unexpected status code: [status code: 406] body: %v", errorMessage)
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
