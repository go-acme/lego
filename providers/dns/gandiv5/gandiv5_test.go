package gandiv5

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAPIKey, EnvPersonalAccessToken)

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
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "gandiv5: credentials information are missing",
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
				require.NotNil(t, p.inProgressFQDNs)
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
			expected: "gandiv5: credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.inProgressFQDNs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

// TestDNSProvider runs Present and CleanUp against a fake Gandi RPC
// Server, whose responses are predetermined for particular requests.
func TestDNSProvider(t *testing.T) {
	provider := servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.PersonalAccessToken = "123412341234123412341234"
			config.BaseURL = server.URL
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer 123412341234123412341234"),
	).
		Route("GET /domains/example.com/records/_acme-challenge.abc.def/TXT",
			servermock.RawStringResponse(`{"rrset_ttl":300,"rrset_values":[],"rrset_name":"_acme-challenge.abc.def","rrset_type":"TXT"}`)).
		Route("PUT /domains/example.com/records/_acme-challenge.abc.def/TXT",
			servermock.RawStringResponse(`{"message": "Zone Record Created"}`),
			servermock.CheckRequestJSONBody(`{"rrset_ttl":300,"rrset_values":["ezRpBPY8wH8djMLYjX2uCKPwiKDkFZ1SFMJ6ZXGlHrQ"]}`)).
		Route("DELETE /domains/example.com/records/_acme-challenge.abc.def/TXT", nil).
		Build(t)

	fakeKeyAuth := "XXXX"

	// define function to override findZoneByFqdn with
	fakeFindZoneByFqdn := func(fqdn string) (string, error) {
		return "example.com.", nil
	}

	// override findZoneByFqdn function
	savedFindZoneByFqdn := provider.findZoneByFqdn
	defer func() {
		provider.findZoneByFqdn = savedFindZoneByFqdn
	}()
	provider.findZoneByFqdn = fakeFindZoneByFqdn

	// run Present
	err := provider.Present("abc.def.example.com", "", fakeKeyAuth)
	require.NoError(t, err)

	// run CleanUp
	err = provider.CleanUp("abc.def.example.com", "", fakeKeyAuth)
	require.NoError(t, err)
}
