package vinyldns

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

const (
	targetRootDomain  = "example.com"
	targetDomain      = "host." + targetRootDomain
	zoneID            = "00000000-0000-0000-0000-000000000000"
	newRecordSetID    = "11000000-0000-0000-0000-000000000000"
	newCreateChangeID = "20000000-0000-0000-0000-000000000000"
	recordID          = "30000000-0000-0000-0000-000000000000"
)

var envTest = tester.NewEnvTest(
	EnvAccessKey,
	EnvSecretKey,
	EnvHost).
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
				EnvAccessKey: "123",
				EnvSecretKey: "456",
				EnvHost:      "https://example.org",
			},
		},
		{
			desc: "missing all credentials",
			envVars: map[string]string{
				EnvHost: "https://example.org",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_ACCESS_KEY,VINYLDNS_SECRET_KEY",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				EnvSecretKey: "456",
				EnvHost:      "https://example.org",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_ACCESS_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAccessKey: "123",
				EnvHost:      "https://example.org",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_SECRET_KEY",
		},
		{
			desc: "missing host",
			envVars: map[string]string{
				EnvAccessKey: "123",
				EnvSecretKey: "456",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_HOST",
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
		desc      string
		accessKey string
		secretKey string
		host      string
		expected  string
	}{
		{
			desc:      "success",
			accessKey: "123",
			secretKey: "456",
			host:      "https://example.org",
		},
		{
			desc:     "missing all credentials",
			host:     "https://example.org",
			expected: "vinyldns: credentials are missing",
		},
		{
			desc:      "missing access key",
			secretKey: "456",
			host:      "https://example.org",
			expected:  "vinyldns: credentials are missing",
		},
		{
			desc:      "missing secret key",
			accessKey: "123",
			host:      "https://example.org",
			expected:  "vinyldns: credentials are missing",
		},
		{
			desc:      "missing host",
			accessKey: "123",
			secretKey: "456",
			expected:  "vinyldns: host is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessKey = test.accessKey
			config.SecretKey = test.secretKey
			config.Host = test.host

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

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		config := NewDefaultConfig()
		config.AccessKey = "foo"
		config.SecretKey = "bar"
		config.Host = server.URL
		config.HTTPClient = server.Client()

		return NewDNSProviderConfig(config)
	})
}

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		desc    string
		keyAuth string
		builder *servermock.Builder[*DNSProvider]
	}{
		{
			desc:    "new record",
			keyAuth: "123456d==",
			builder: mockBuilder().
				Route("GET /zones/name/"+targetRootDomain+".",
					servermock.ResponseFromFixture("zoneByName.json")).
				Route("GET /zones/"+zoneID+"/recordsets",
					servermock.ResponseFromFixture("recordSetsListAll-empty.json")).
				Route("POST /zones/"+zoneID+"/recordsets",
					servermock.ResponseFromFixture("recordSetUpdate-create.json").
						WithStatusCode(http.StatusAccepted)).
				Route("GET /zones/"+zoneID+"/recordsets/"+newRecordSetID+"/changes/"+newCreateChangeID,
					servermock.ResponseFromFixture("recordSetChange-create.json")),
		},
		{
			desc:    "existing record",
			keyAuth: "123456d==",
			builder: mockBuilder().
				Route("GET /zones/name/"+targetRootDomain+".",
					servermock.ResponseFromFixture("zoneByName.json")).
				Route("GET /zones/"+zoneID+"/recordsets",
					servermock.ResponseFromFixture("recordSetsListAll.json")),
		},
		{
			desc:    "duplicate key",
			keyAuth: "abc123!!",
			builder: mockBuilder().
				Route("GET /zones/name/"+targetRootDomain+".",
					servermock.ResponseFromFixture("zoneByName.json")).
				Route("GET /zones/"+zoneID+"/recordsets",
					servermock.ResponseFromFixture("recordSetsListAll.json")).
				Route("PUT /zones/"+zoneID+"/recordsets/"+recordID,
					servermock.ResponseFromFixture("recordSetUpdate-create.json").
						WithStatusCode(http.StatusAccepted)).
				Route("GET /zones/"+zoneID+"/recordsets/"+newRecordSetID+"/changes/"+newCreateChangeID,
					servermock.ResponseFromFixture("recordSetChange-create.json")),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			provider := test.builder.Build(t)

			err := provider.Present(targetDomain, "token"+test.keyAuth, test.keyAuth)
			require.NoError(t, err)
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /zones/name/"+targetRootDomain+".",
			servermock.ResponseFromFixture("zoneByName.json")).
		Route("GET /zones/"+zoneID+"/recordsets",
			servermock.ResponseFromFixture("recordSetsListAll.json")).
		Route("DELETE /zones/"+zoneID+"/recordsets/"+recordID,
			servermock.ResponseFromFixture("recordSetDelete.json").
				WithStatusCode(http.StatusAccepted)).
		Route("GET /zones/"+zoneID+"/recordsets/"+newRecordSetID+"/changes/"+newCreateChangeID,
			servermock.ResponseFromFixture("recordSetChange-delete.json")).
		Build(t)

	err := provider.CleanUp(targetDomain, "123456d==", "123456d==")
	require.NoError(t, err)
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
