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

// FIXME
func TestFoo(t *testing.T) {
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
