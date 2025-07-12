package auroradns

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/nrdcg/auroradns"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAPIKey, EnvSecret)

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIKey = "asdf1234"
			config.Secret = "key"
			config.BaseURL = server.URL

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithContentType("application/json").
			WithRegexp("Authorization", `AuroraDNSv1 .+`).
			WithRegexp("X-Auroradns-Date", `[0-9TZ]+`))
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
				EnvSecret: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey: "",
				EnvSecret: "",
			},
			expected: "aurora: some credentials information are missing: AURORA_API_KEY,AURORA_SECRET",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIKey: "",
				EnvSecret: "456",
			},
			expected: "aurora: some credentials information are missing: AURORA_API_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAPIKey: "123",
				EnvSecret: "",
			},
			expected: "aurora: some credentials information are missing: AURORA_SECRET",
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
		apiKey   string
		secret   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
			secret: "456",
		},
		{
			desc:     "missing credentials",
			apiKey:   "",
			secret:   "",
			expected: "aurora: some credentials information are missing",
		},
		{
			desc:     "missing user id",
			apiKey:   "",
			secret:   "456",
			expected: "aurora: some credentials information are missing",
		},
		{
			desc:     "missing key",
			apiKey:   "123",
			secret:   "",
			expected: "aurora: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.Secret = test.secret

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
	provider := mockBuilder().
		Route("GET /zones",
			servermock.JSONEncode([]auroradns.Zone{{
				ID:   "c56a4180-65aa-42ec-a945-5fd21dec0538",
				Name: "example.com",
			}}).
				WithStatusCode(http.StatusCreated)).
		Route("POST /zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records",
			servermock.JSONEncode(auroradns.Record{
				ID:         "ec56a4180-65aa-42ec-a945-5fd21dec0538",
				RecordType: "TXT",
				Name:       "_acme-challenge",
				TTL:        300,
			}).
				WithStatusCode(http.StatusCreated)).
		Build(t)

	err := provider.Present("example.com", "", "foobar")
	require.NoError(t, err, "fail to create TXT record")
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /zones",
			servermock.JSONEncode([]auroradns.Zone{{
				ID:   "c56a4180-65aa-42ec-a945-5fd21dec0538",
				Name: "example.com",
			}}).
				WithStatusCode(http.StatusCreated)).
		Route("POST /zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records",
			servermock.JSONEncode(auroradns.Record{
				ID:         "ec56a4180-65aa-42ec-a945-5fd21dec0538",
				RecordType: "TXT",
				Name:       "_acme-challenge",
				TTL:        300,
			}).
				WithStatusCode(http.StatusCreated)).
		Route("DELETE /zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records/ec56a4180-65aa-42ec-a945-5fd21dec0538",
			servermock.RawStringResponse("{}").
				WithStatusCode(http.StatusCreated)).
		Build(t)

	err := provider.Present("example.com", "", "foobar")
	require.NoError(t, err, "fail to create TXT record")

	err = provider.CleanUp("example.com", "", "foobar")
	require.NoError(t, err, "fail to remove TXT record")
}
