package civo

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIToken).
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
				EnvAPIToken: "00000000000000000000000000000000000000000000000000",
			},
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIToken: "",
			},
			expected: fmt.Sprintf("civo: some credentials information are missing: %s", EnvAPIToken),
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
		desc     string
		token    string
		ttl      int
		expected string
	}{
		{
			desc:  "success",
			token: "00000000000000000000000000000000000000000000000000",
			ttl:   minTTL,
		},
		{
			desc:     "missing api key",
			token:    "",
			ttl:      minTTL,
			expected: "civo: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.TTL = test.ttl
			config.Token = test.token

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

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.HTTPClient = server.Client()
			config.Token = "secret"

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("Authorization", "Bearer secret").
			WithRegexp("User-Agent", `goacme-lego/[0-9.]+ \(.+\)`),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		// https://www.civo.com/api/dns#list-domain-names
		Route("GET /dns",
			servermock.ResponseFromInternal("list_domain_names.json"),
			servermock.CheckQueryParameter().Strict().
				With("region", "LON1")).
		// https://www.civo.com/api/dns#create-a-new-dns-record
		Route("POST /dns/7088fcea-7658-43e6-97fa-273f901978fd/records",
			servermock.ResponseFromInternal("create_dns_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_dns_record-request.json")).
		Build(t)

	err := provider.Present("example.com", "abd", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		// https://www.civo.com/api/dns#list-domain-names
		Route("GET /dns",
			servermock.ResponseFromInternal("list_domain_names.json"),
			servermock.CheckQueryParameter().
				With("region", "LON1")).
		// https://www.civo.com/api/dns#list-dns-records
		Route("GET /dns/7088fcea-7658-43e6-97fa-273f901978fd/records",
			servermock.ResponseFromInternal("list_dns_records.json"),
			servermock.CheckQueryParameter().Strict().
				With("region", "LON1")).
		// https://www.civo.com/api/dns#deleting-a-dns-record
		Route("DELETE /dns/edc5dacf-a2ad-4757-41ee-c12f06259c70/records/76cc107f-fbef-4e2b-b97f-f5d34f4075d3",
			servermock.ResponseFromInternal("delete_dns_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("region", "LON1")).
		Build(t)

	err := provider.CleanUp("example.com", "abd", "123d==")
	require.NoError(t, err)
}
