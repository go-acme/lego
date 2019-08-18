package liquidweb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest("LIQUID_WEB_URL", "LIQUID_WEB_USERNAME", "LIQUID_WEB_PASSWORD", "LIQUID_WEB_TIMEOUT", "LIQUID_WEB_ZONE").WithDomain("LIQUID_WEB_DOMAIN")

func setupTest() (*DNSProvider, *http.ServeMux, func()) {
	handler := http.NewServeMux()
	server := httptest.NewServer(handler)
	config := NewDefaultConfig()
	config.Username = "blars"
	config.Password = "tacoman"
	config.URL = server.URL
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
				"LIQUID_WEB_URL":      "https://storm.com",
				"LIQUID_WEB_USERNAME": "blars",
				"LIQUID_WEB_PASSWORD": "tacoman",
			},
		},
		{
			desc:     "missing url",
			envVars:  map[string]string{},
			expected: "liquidweb: url is missing",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				"LIQUID_WEB_URL": "https://storm.com",
			},
			expected: "liquidweb: username is missing",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				"LIQUID_WEB_URL":      "https://storm.com",
				"LIQUID_WEB_USERNAME": "blars",
			}, expected: "liquidweb: password is missing",
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
		assert.Equal(t, http.MethodPost, r.Method, "method")

		assert.Equal(t, "Basic YmxhcnM6dGFjb21hbg==", r.Header.Get("Authorization"), "Authorization")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		expectedReqBody := `{"params":{"name":"_acme-challenge.tacoman.com","rdata":"\"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU\"","type":"TXT","zone":"tacoman.com"}}`
		assert.Equal(t, expectedReqBody, string(reqBody))

		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, `{
			"type": "TXT",
			"name": "_acme-challenge.tacoman.com",
			"rdata": "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
			"id": 1234567,
			"prio": null
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	err := provider.Present("tacoman.com", "", "")
	fmt.Printf("%+v", err)
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider, mux, tearDown := setupTest()
	defer tearDown()

	mux.HandleFunc("/v1/Network/DNS/Record/delete", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "method")

		assert.Equal(t, "/v1/Network/DNS/Record/delete", r.URL.Path, "Path")
		assert.Equal(t, "Basic YmxhcnM6dGFjb21hbg==", r.Header.Get("Authorization"), "Authorization")

		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, `{
			"deleted": "123"
		}`)
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
