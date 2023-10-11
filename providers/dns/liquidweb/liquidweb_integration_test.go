package liquidweb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liquidweb/liquidweb-go/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	config := NewDefaultConfig()
	config.Username = "blars"
	config.Password = "tacoman"
	config.BaseURL = server.URL
	config.Zone = "tacoman.com"

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	return provider, mux
}

// func TestDNSProvider_Present(t *testing.T) {
// 	provider, mux := setupTest(t)
//
// 	mux.HandleFunc("/v1/Network/DNS/Record/create", func(w http.ResponseWriter, r *http.Request) {
// 		assert.Equal(t, http.MethodPost, r.Method)
//
// 		username, password, ok := r.BasicAuth()
// 		assert.Equal(t, "blars", username)
// 		assert.Equal(t, "tacoman", password)
// 		assert.True(t, ok)
//
// 		reqBody, err := io.ReadAll(r.Body)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 		expectedReqBody := `
// 			{
// 				"params": {
// 					"name": "_acme-challenge.tacoman.com",
// 					"rdata": "\"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU\"",
// 					"ttl": 300,
// 					"type": "TXT",
// 					"zone": "tacoman.com"
// 				}
// 			}`
// 		assert.JSONEq(t, expectedReqBody, string(reqBody))
//
// 		w.WriteHeader(http.StatusOK)
// 		_, err = fmt.Fprintf(w, `{
// 			"type": "TXT",
// 			"name": "_acme-challenge.tacoman.com",
// 			"rdata": "\"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU\"",
// 			"ttl": 300,
// 			"id": 1234567,
// 			"prio": null
// 		}`)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	})
//
// 	err := provider.Present("tacoman.com", "", "")
// 	require.NoError(t, err)
// }
//
// func TestDNSProvider_CleanUp(t *testing.T) {
// 	provider, mux := setupTest(t)
//
// 	mux.HandleFunc("/v1/Network/DNS/Record/delete", func(w http.ResponseWriter, r *http.Request) {
// 		assert.Equal(t, http.MethodPost, r.Method)
//
// 		username, password, ok := r.BasicAuth()
// 		assert.Equal(t, "blars", username)
// 		assert.Equal(t, "tacoman", password)
// 		assert.True(t, ok)
//
// 		_, err := fmt.Fprintf(w, `{"deleted": "123"}`)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	})
//
// 	provider.recordIDs["123"] = 1234567
//
// 	err := provider.CleanUp("tacoman.com.", "123", "")
// 	require.NoError(t, err, "fail to remove TXT record")
// }

func TestDNSProvider_Present(t *testing.T) {
	envTest.Apply(map[string]string{
		EnvUsername: "blars",
		EnvPassword: "tacoman",
		EnvURL:      mockAPIServer(t),
		EnvZone:     "tacoman.com", // this needs to be removed from test?
	})

	defer envTest.ClearEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present("tacoman.com", "", "")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	envTest.Apply(map[string]string{
		EnvUsername: "blars",
		EnvPassword: "tacoman",
		EnvURL: mockAPIServer(t, network.DNSRecord{
			Name:   "_acme-challenge.tacoman.com",
			RData:  "123d==",
			Type:   "TXT",
			TTL:    300,
			ID:     1234567,
			ZoneID: 42,
		}),
		EnvZone: "tacoman.com", // this needs to be removed from test?
	})

	defer envTest.ClearEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	provider.recordIDs["123d=="] = 1234567

	err = provider.CleanUp("tacoman.com.", "123d==", "")
	require.NoError(t, err, "fail to remove TXT record")
}

// FIXME
func TestIntegration(t *testing.T) {
	testCases := []struct {
		desc          string
		envVars       map[string]string
		initRecs      []network.DNSRecord
		domain        string
		token         string
		keyauth       string
		present       bool
		cleanup       bool
		expPresentErr string
		expCleanupErr string
	}{
		{
			desc: "expected successful",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvPassword: "tacoman",
			},
			domain:  "tacoman.com",
			token:   "123",
			keyauth: "456",
			present: true,
			cleanup: true,
		},
		{
			desc: "other successful",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvPassword: "tacoman",
			},
			domain:  "banana.com",
			token:   "123",
			keyauth: "456",
			present: true,
			cleanup: true,
		},
		{
			desc: "zone not on account",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvPassword: "tacoman",
			},
			domain:        "huckleberry.com",
			token:         "123",
			keyauth:       "456",
			present:       true,
			cleanup:       false,
			expPresentErr: "no valid zone in account for certificate _acme-challenge.huckleberry.com",
		},
		{
			desc: "ssl for domain",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvPassword: "tacoman",
			},
			domain:        "sundae.cherry.com",
			token:         "5847953",
			keyauth:       "34872934",
			present:       true,
			cleanup:       true,
			expPresentErr: "",
		},
		{
			desc: "complicated domain",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvPassword: "tacoman",
			},
			domain:        "always.money.stand.banana.com",
			token:         "5847953",
			keyauth:       "there is always money in the banana stand",
			present:       true,
			cleanup:       true,
			expPresentErr: "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			test.envVars[EnvURL] = mockAPIServer(t, test.initRecs...)
			envTest.ClearEnv()
			envTest.Apply(test.envVars)

			provider, err := NewDNSProvider()
			require.NoError(t, err)

			if test.present {
				err = provider.Present(test.domain, test.token, test.keyauth)
				if test.expPresentErr == "" {
					assert.NoError(t, err)
				} else {
					assert.Equal(t, test.expPresentErr, err.Error())
				}
			}

			if test.cleanup {
				err = provider.CleanUp(test.domain, test.token, test.keyauth)
				if test.expCleanupErr == "" {
					assert.NoError(t, err)
				} else {
					assert.Equal(t, test.expCleanupErr, err.Error())
				}
			}
		})
	}
}
