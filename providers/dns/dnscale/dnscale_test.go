package dnscale

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/providers/dns/dnscale/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIToken).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIToken: "test-token",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIToken: "",
			},
			expected: "dnscale: some credentials information are missing: DNSCALE_API_TOKEN",
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
		apiToken string
		expected string
	}{
		{
			desc:     "success",
			apiToken: "test-token",
		},
		{
			desc:     "missing credentials",
			expected: "dnscale: some credentials information are missing: DNSCALE_API_TOKEN",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIToken = test.apiToken

			p, err := NewDNSProviderConfig(config)

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

func TestNewDNSProviderConfig_NilConfig(t *testing.T) {
	_, err := NewDNSProviderConfig(nil)
	require.EqualError(t, err, "dnscale: the configuration of the DNS provider is nil")
}

func setupMockServer(t *testing.T, zones []internal.Zone) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(internal.ZonesResponse{
			Status: "success",
			Data:   internal.ZonesData{Zones: zones},
		})
	})

	mux.HandleFunc("/v1/zones/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)

			assert.Equal(t, "TXT", body["type"])
			assert.NotEmpty(t, body["name"])
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			content := r.URL.Query().Get("content")
			assert.NotEmpty(t, content)

			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func TestPresent(t *testing.T) {
	server := setupMockServer(t, []internal.Zone{
		{ID: "zone-123", Name: "example.com"},
	})
	defer server.Close()

	config := NewDefaultConfig()
	config.APIToken = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestPresent_Subdomain(t *testing.T) {
	server := setupMockServer(t, []internal.Zone{
		{ID: "zone-123", Name: "example.com"},
	})
	defer server.Close()

	config := NewDefaultConfig()
	config.APIToken = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present("sub.example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestPresent_ZoneNotFound(t *testing.T) {
	server := setupMockServer(t, []internal.Zone{
		{ID: "zone-other", Name: "other.com"},
	})
	defer server.Close()

	config := NewDefaultConfig()
	config.APIToken = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	require.Error(t, err)
}

func TestPresent_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(internal.ZonesResponse{
			Status: "success",
			Data:   internal.ZonesData{Zones: []internal.Zone{{ID: "zone-123", Name: "example.com"}}},
		})
	})
	mux.HandleFunc("/v1/zones/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "forbidden", "message": "insufficient permissions",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	config := NewDefaultConfig()
	config.APIToken = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	require.Error(t, err)
}

func TestCleanUp(t *testing.T) {
	server := setupMockServer(t, []internal.Zone{
		{ID: "zone-123", Name: "example.com"},
	})
	defer server.Close()

	config := NewDefaultConfig()
	config.APIToken = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestCleanUp_ZoneNotFound(t *testing.T) {
	server := setupMockServer(t, []internal.Zone{})
	defer server.Close()

	config := NewDefaultConfig()
	config.APIToken = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp("example.com", "token", "keyAuth")
	require.Error(t, err)
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
