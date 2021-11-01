package auroradns

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	EnvUserID,
	EnvKey)

func setupTest(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	config := NewDefaultConfig()
	config.UserID = "asdf1234"
	config.Key = "key"
	config.BaseURL = server.URL

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
				EnvUserID: "123",
				EnvKey:    "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvUserID: "",
				EnvKey:    "",
			},
			expected: "aurora: some credentials information are missing: AURORA_USER_ID,AURORA_KEY",
		},
		{
			desc: "missing user id",
			envVars: map[string]string{
				EnvUserID: "",
				EnvKey:    "456",
			},
			expected: "aurora: some credentials information are missing: AURORA_USER_ID",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				EnvUserID: "123",
				EnvKey:    "",
			},
			expected: "aurora: some credentials information are missing: AURORA_KEY",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		userID   string
		key      string
		expected string
	}{
		{
			desc:   "success",
			userID: "123",
			key:    "456",
		},
		{
			desc:     "missing credentials",
			userID:   "",
			key:      "",
			expected: "aurora: some credentials information are missing",
		},
		{
			desc:     "missing user id",
			userID:   "",
			key:      "456",
			expected: "aurora: some credentials information are missing",
		},
		{
			desc:     "missing key",
			userID:   "123",
			key:      "",
			expected: "aurora: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.UserID = test.userID
			config.Key = test.key

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `[{
			        "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
			        "name": "example.com"
			      }]`)
	})

	mux.HandleFunc("/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")

		reqBody, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"type":"TXT","name":"_acme-challenge","content":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":300}`, string(reqBody))

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
		      "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
		      "type": "TXT",
		      "name": "_acme-challenge",
		      "ttl":  300
		    }`)
	})

	err := provider.Present("example.com", "", "foobar")
	require.NoError(t, err, "fail to create TXT record")
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `[{
			        "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
			        "name": "example.com"
			      }]`)
	})

	mux.HandleFunc("/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
			        "id":   "ec56a4180-65aa-42ec-a945-5fd21dec0538",
			        "type": "TXT",
			        "name": "_acme-challenge",
			        "ttl":  300
			      }`)
	})

	mux.HandleFunc("/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records/ec56a4180-65aa-42ec-a945-5fd21dec0538", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{}`)
	})

	err := provider.Present("example.com", "", "foobar")
	require.NoError(t, err, "fail to create TXT record")

	err = provider.CleanUp("example.com", "", "foobar")
	require.NoError(t, err, "fail to remove TXT record")
}
