package iij

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "TESTDOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIAccessKey,
	EnvAPISecretKey,
	EnvDoServiceCode).
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
				EnvAPIAccessKey:  "A",
				EnvAPISecretKey:  "B",
				EnvDoServiceCode: "C",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIAccessKey:  "",
				EnvAPISecretKey:  "",
				EnvDoServiceCode: "",
			},
			expected: "iij: some credentials information are missing: IIJ_API_ACCESS_KEY,IIJ_API_SECRET_KEY,IIJ_DO_SERVICE_CODE",
		},
		{
			desc: "missing api access key",
			envVars: map[string]string{
				EnvAPIAccessKey:  "",
				EnvAPISecretKey:  "B",
				EnvDoServiceCode: "C",
			},
			expected: "iij: some credentials information are missing: IIJ_API_ACCESS_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAPIAccessKey:  "A",
				EnvAPISecretKey:  "",
				EnvDoServiceCode: "C",
			},
			expected: "iij: some credentials information are missing: IIJ_API_SECRET_KEY",
		},
		{
			desc: "missing do service code",
			envVars: map[string]string{
				EnvAPIAccessKey:  "A",
				EnvAPISecretKey:  "B",
				EnvDoServiceCode: "",
			},
			expected: "iij: some credentials information are missing: IIJ_DO_SERVICE_CODE",
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
				require.NotNil(t, p.api)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		accessKey     string
		secretKey     string
		doServiceCode string
		expected      string
	}{
		{
			desc:          "success",
			accessKey:     "A",
			secretKey:     "B",
			doServiceCode: "C",
		},
		{
			desc:     "missing credentials",
			expected: "iij: credentials missing",
		},
		{
			desc:          "missing access key",
			accessKey:     "",
			secretKey:     "B",
			doServiceCode: "C",
			expected:      "iij: credentials missing",
		},
		{
			desc:          "missing secret key",
			accessKey:     "A",
			secretKey:     "",
			doServiceCode: "C",
			expected:      "iij: credentials missing",
		},
		{
			desc:          "missing do service code",
			accessKey:     "A",
			secretKey:     "B",
			doServiceCode: "",
			expected:      "iij: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessKey = test.accessKey
			config.SecretKey = test.secretKey
			config.DoServiceCode = test.doServiceCode

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.api)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestSplitDomain(t *testing.T) {
	testCases := []struct {
		desc          string
		domain        string
		zones         []string
		expectedOwner string
		expectedZone  string
	}{
		{
			desc:          "domain equals zone",
			domain:        "example.com",
			zones:         []string{"example.com"},
			expectedOwner: "",
			expectedZone:  "example.com",
		},
		{
			desc:          "with a subdomain",
			domain:        "_acme-challenge.my.example.com",
			zones:         []string{"example.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "example.com",
		},
		{
			desc:          "with a subdomain in a zone",
			domain:        "_acme-challenge.my.sub.example.com",
			zones:         []string{"sub.example.com", "example.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "sub.example.com",
		},
		{
			desc:          "with a sub-subdomain",
			domain:        "_acme-challenge.my.sub.example.com",
			zones:         []string{"domain1.com", "example.com"},
			expectedOwner: "_acme-challenge.my.sub",
			expectedZone:  "example.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			owner, zone, err := splitDomain(test.domain, test.zones)
			require.NoError(t, err)

			assert.Equal(t, test.expectedOwner, owner)
			assert.Equal(t, test.expectedZone, zone)
		})
	}
}

func TestSplitDomain_error(t *testing.T) {
	testCases := []struct {
		desc          string
		domain        string
		zones         []string
		expectedOwner string
		expectedZone  string
	}{
		{
			desc:   "no zone",
			domain: "example.com",
			zones:  nil,
		},
		{
			desc:   "domain does not contain zone",
			domain: "example.com",
			zones:  []string{"example.org"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, _, err := splitDomain(test.domain, test.zones)
			require.Error(t, err)
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

func mockBuilder() *servermock2.Builder[*DNSProvider] {
	return servermock2.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.AccessKey = "user"
			config.SecretKey = "secret"
			config.DoServiceCode = "123"

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.api.Endpoint = server.URL

			return p, nil
		},
		servermock2.CheckHeader().
			WithRegexp("Authorization", "IIJAPI user:.+").
			WithRegexp("X-Iijapi-Expire", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.+`).
			With("X-Iijapi-Signaturemethod", "HmacSHA256").
			With("X-Iijapi-Signatureversion", "2"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /r/20140601/123/zones.json",
			servermock2.ResponseFromFixture("zones.json"),
		).
		Route("POST /r/20140601/123/example.com/record.json",
			servermock2.ResponseFromFixture("record.json"),
			servermock2.CheckRequestJSONBodyFromFixture("record-request.json"),
		).
		Route("PUT /r/20140601/123/commit.json",
			servermock2.ResponseFromFixture("commit.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /r/20140601/123/zones.json",
			servermock2.ResponseFromFixture("zones.json"),
		).
		Route("GET /r/20140601/123/example.com/records/DETAIL.json",
			servermock2.ResponseFromFixture("detail.json"),
		).
		Route("DELETE /r/20140601/123/example.com/record/963.json",
			servermock2.ResponseFromFixture("delete.json"),
		).
		Route("PUT /r/20140601/123/commit.json",
			servermock2.ResponseFromFixture("commit.json"),
		).
		Route("/",
			servermock2.DumpRequest(),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
