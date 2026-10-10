package selfhostdev2

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvTokens, EnvRecordsMapping).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvTokens:         "example.com:42.base64",
				EnvRecordsMapping: "example.com:123",
			},
		},
		{
			desc: "missing tokens",
			envVars: map[string]string{
				EnvTokens:         "",
				EnvRecordsMapping: "example.com:123",
			},
			expected: "selfhostdev2: some credentials information are missing: SELFHOSTDEV2_API_KEYS",
		},
		{
			desc: "missing records mapping",
			envVars: map[string]string{
				EnvTokens: "example.com:42.base64",
			},
			expected: "selfhostdev2: some credentials information are missing: SELFHOSTDEV2_RECORDS_MAPPING",
		},
		{
			desc: "invalid records mapping",
			envVars: map[string]string{
				EnvTokens:         "example.com:42.base64",
				EnvRecordsMapping: "example.com",
			},
			expected: `selfhostdev2: malformed records mapping: missing ":": example.com`,
		},
		{
			desc:     "missing information",
			envVars:  map[string]string{},
			expected: "selfhostdev2: some credentials information are missing: SELFHOSTDEV2_API_KEYS,SELFHOSTDEV2_RECORDS_MAPPING",
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
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		credentials   map[string]string
		recordMapping map[string]*Seq
		expected      string
	}{
		{
			desc: "success",
			credentials: map[string]string{
				"example.com": "42.base64",
			},
			recordMapping: map[string]*Seq{
				"example.com": NewSeq(123),
			},
		},
		{
			desc: "missing tokens",
			recordMapping: map[string]*Seq{
				"example.com": NewSeq(123),
			},
			expected: "selfhostdev2: credentials missing",
		},
		{
			desc: "missing sequence",
			credentials: map[string]string{
				"example.com": "42.base64",
			},
			recordMapping: map[string]*Seq{
				"example.com": nil,
			},
			expected: `selfhostdev2: missing record ID for "example.com"`,
		},
		{
			desc: "empty sequence",
			credentials: map[string]string{
				"example.com": "42.base64",
			},
			recordMapping: map[string]*Seq{
				"example.com": NewSeq(),
			},
			expected: `selfhostdev2: missing record ID for "example.com"`,
		},
		{
			desc: "missing records mapping",
			credentials: map[string]string{
				"example.com": "42.base64",
			},
			expected: "selfhostdev2: missing record mapping",
		},
		{
			desc: "empty records mapping",
			credentials: map[string]string{
				"example.com": "42.base64",
			},
			recordMapping: map[string]*Seq{},
			expected:      "selfhostdev2: missing record mapping",
		},
		{
			desc:     "missing information",
			expected: "selfhostdev2: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Credentials = test.credentials
			config.RecordsMapping = test.recordMapping

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Credentials = map[string]string{
				"example.com": "42.base64",
				"example.org": "43.base64",
			}
			config.RecordsMapping = map[string]*Seq{
				"example.com": NewSeq(123),
			}
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
		Route("POST /",
			servermock.ResponseFromInternal("present.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /",
			servermock.ResponseFromInternal("cleanup.json"),
		).
		Build(t)

	provider.recordIDs["abc"] = 123456

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
