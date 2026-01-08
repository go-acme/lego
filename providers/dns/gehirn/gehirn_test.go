package gehirn

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvTokenID, EnvTokenSecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvTokenID:     "abc123",
				EnvTokenSecret: "secret",
			},
		},
		{
			desc: "missing token ID",
			envVars: map[string]string{
				EnvTokenID:     "",
				EnvTokenSecret: "secret",
			},
			expected: "gehirn: some credentials information are missing: GEHIRN_TOKEN_ID",
		},
		{
			desc: "missing token secret",
			envVars: map[string]string{
				EnvTokenID:     "abc123",
				EnvTokenSecret: "",
			},
			expected: "gehirn: some credentials information are missing: GEHIRN_TOKEN_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "gehirn: some credentials information are missing: GEHIRN_TOKEN_ID,GEHIRN_TOKEN_SECRET",
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
		desc        string
		tokenID     string
		tokenSecret string
		expected    string
	}{
		{
			desc:        "success",
			tokenID:     "abc123",
			tokenSecret: "secret",
		},
		{
			desc:        "missing token ID",
			tokenSecret: "secret",
			expected:    "gehirn: credentials missing",
		},
		{
			desc:     "missing token secret",
			tokenID:  "abc123",
			expected: "gehirn: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "gehirn: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.TokenID = test.tokenID
			config.TokenSecret = test.tokenSecret

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

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.TokenID = "abc123"
			config.TokenSecret = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromInternal("list_zones.json"),
		).
		Route("POST /zones/ZONE-ID-3/versions/VERSION-ID-3/records",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromInternal("list_zones.json"),
		).
		Route("DELETE /zones/ZONE-ID-3/versions/VERSION-ID-3/records/RECORD-ID-1",
			servermock.ResponseFromInternal("delete_record.json"),
		).
		Build(t)

	provider.recordIDs["abc"] = "RECORD-ID-1"

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
