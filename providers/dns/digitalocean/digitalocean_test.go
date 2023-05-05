package digitalocean

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAuthToken)

func setupTest(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	config := NewDefaultConfig()
	config.AuthToken = "asdf1234"
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
				EnvAuthToken: "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAuthToken: "",
			},
			expected: "digitalocean: some credentials information are missing: DO_AUTH_TOKEN",
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
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		authToken string
		expected  string
	}{
		{
			desc:      "success",
			authToken: "123",
		},
		{
			desc:     "missing credentials",
			expected: "digitalocean: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthToken = test.authToken

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
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
	provider, mux := setupTest(t)

	mux.HandleFunc("/v2/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "method")

		assert.Equal(t, "application/json", r.Header.Get("Accept"), "Accept")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Bearer asdf1234", r.Header.Get("Authorization"), "Authorization")

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		expectedReqBody := `{"type":"TXT","name":"_acme-challenge.example.com.","data":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":30}`
		assert.Equal(t, expectedReqBody, string(bytes.TrimSpace(reqBody)))

		w.WriteHeader(http.StatusCreated)
		_, err = fmt.Fprintf(w, `{
			"domain_record": {
				"id": 1234567,
				"type": "TXT",
				"name": "_acme-challenge",
				"data": "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
				"priority": null,
				"port": null,
				"weight": null
			}
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := provider.Present("example.com", "", "foobar")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/v2/domains/example.com/records/1234567", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "method")

		assert.Equal(t, "/v2/domains/example.com/records/1234567", r.URL.Path, "Path")

		assert.Equal(t, "application/json", r.Header.Get("Accept"), "Accept")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Bearer asdf1234", r.Header.Get("Authorization"), "Authorization")

		w.WriteHeader(http.StatusNoContent)
	})

	provider.recordIDsMu.Lock()
	provider.recordIDs["token"] = 1234567
	provider.recordIDsMu.Unlock()

	err := provider.CleanUp("example.com", "token", "")
	require.NoError(t, err, "fail to remove TXT record")
}
