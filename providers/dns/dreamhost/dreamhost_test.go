package dreamhost

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey).
	WithDomain(envDomain)

const (
	fakeAPIKey         = "asdf1234"
	fakeChallengeToken = "foobar"
	fakeKeyAuth        = "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI"
)

func setupTest(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	config := NewDefaultConfig()
	config.APIKey = fakeAPIKey
	config.BaseURL = server.URL
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
				EnvAPIKey: "123",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "dreamhost: some credentials information are missing: DREAMHOST_API_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				assert.NoError(t, err)
				assert.NotNil(t, p)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
		},
		{
			desc:     "missing credentials",
			expected: "dreamhost: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				assert.NoError(t, err)
				assert.NotNil(t, p)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")

		q := r.URL.Query()
		assert.Equal(t, fakeAPIKey, q.Get("key"))
		assert.Equal(t, "dns-add_record", q.Get("cmd"))
		assert.Equal(t, "json", q.Get("format"))
		assert.Equal(t, "_acme-challenge.example.com", q.Get("record"))
		assert.Equal(t, fakeKeyAuth, q.Get("value"))
		assert.Equal(t, "Managed+By+lego", q.Get("comment"))

		_, err := fmt.Fprintf(w, `{"data":"record_added","result":"success"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := provider.Present("example.com", "", fakeChallengeToken)
	require.NoError(t, err)
}

func TestDNSProvider_PresentFailed(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")

		_, err := fmt.Fprintf(w, `{"data":"record_already_exists_remove_first","result":"error"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := provider.Present("example.com", "", fakeChallengeToken)
	require.EqualError(t, err, "dreamhost: add TXT record failed: record_already_exists_remove_first")
}

func TestDNSProvider_Cleanup(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")

		q := r.URL.Query()
		assert.Equal(t, fakeAPIKey, q.Get("key"), "key mismatch")
		assert.Equal(t, "dns-remove_record", q.Get("cmd"), "cmd mismatch")
		assert.Equal(t, "json", q.Get("format"))
		assert.Equal(t, "_acme-challenge.example.com", q.Get("record"))
		assert.Equal(t, fakeKeyAuth, q.Get("value"), "value mismatch")
		assert.Equal(t, "Managed+By+lego", q.Get("comment"))

		_, err := fmt.Fprintf(w, `{"data":"record_removed","result":"success"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := provider.CleanUp("example.com", "", fakeChallengeToken)
	require.NoError(t, err, "failed to remove TXT record")
}

func TestLivePresentAndCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
