package liquidweb

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvURL,
	EnvUsername,
	EnvPassword,
	EnvZone).
	WithDomain(envDomain)

func setupTest() (*DNSProvider, *http.ServeMux, func()) {
	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	config := NewDefaultConfig()
	config.Username = "blars"
	config.Password = "tacoman"
	config.BaseURL = server.URL
	config.Zone = "tacoman.com"

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
				EnvURL:      "https://storm.com",
				EnvUsername: "blars",
				EnvPassword: "tacoman",
				EnvZone:     "blars.com",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "liquidweb: some credentials information are missing: LIQUID_WEB_USERNAME,LIQUID_WEB_PASSWORD,LIQUID_WEB_ZONE",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvPassword: "tacoman",
				EnvZone:     "blars.com",
			},
			expected: "liquidweb: some credentials information are missing: LIQUID_WEB_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvZone:     "blars.com",
			}, expected: "liquidweb: some credentials information are missing: LIQUID_WEB_PASSWORD",
		},
		{
			desc: "missing zone",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvPassword: "tacoman",
			}, expected: "liquidweb: some credentials information are missing: LIQUID_WEB_ZONE",
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
				require.NotNil(t, p.client)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		username string
		password string
		zone     string
		expected string
	}{
		{
			desc:     "success",
			username: "acme",
			password: "secret",
			zone:     "example.com",
		},
		{
			desc:     "missing credentials",
			username: "",
			password: "",
			zone:     "",
			expected: "liquidweb: zone is missing",
		},
		{
			desc:     "missing username",
			username: "",
			password: "secret",
			zone:     "example.com",
			expected: "liquidweb: username is missing",
		},
		{
			desc:     "missing password",
			username: "acme",
			password: "",
			zone:     "example.com",
			expected: "liquidweb: password is missing",
		},
		{
			desc:     "missing zone",
			username: "acme",
			password: "secret",
			zone:     "",
			expected: "liquidweb: zone is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.Zone = test.zone

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	provider, mux, tearDown := setupTest()
	defer tearDown()

	mux.HandleFunc("/v1/Network/DNS/Record/create", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		username, password, ok := r.BasicAuth()
		assert.Equal(t, "blars", username)
		assert.Equal(t, "tacoman", password)
		assert.True(t, ok)

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		expectedReqBody := `
			{
				"params": {
					"name": "_acme-challenge.tacoman.com",
					"rdata": "\"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU\"",
					"ttl": 300,
					"type": "TXT",
					"zone": "tacoman.com"
				}
			}`
		assert.JSONEq(t, expectedReqBody, string(reqBody))

		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, `{
			"type": "TXT",
			"name": "_acme-challenge.tacoman.com",
			"rdata": "\"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU\"",
			"ttl": 300,
			"id": 1234567,
			"prio": null
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	err := provider.Present("tacoman.com", "", "")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider, mux, tearDown := setupTest()
	defer tearDown()

	mux.HandleFunc("/v1/Network/DNS/Record/delete", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		username, password, ok := r.BasicAuth()
		assert.Equal(t, "blars", username)
		assert.Equal(t, "tacoman", password)
		assert.True(t, ok)

		_, err := fmt.Fprintf(w, `{"deleted": "123"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	provider.recordIDs["123"] = 1234567

	err := provider.CleanUp("tacoman.com.", "123", "")
	require.NoError(t, err, "fail to remove TXT record")
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
