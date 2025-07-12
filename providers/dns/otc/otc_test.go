package otc

import (
	"fmt"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/otc/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvDomainName,
	EnvUserName,
	EnvPassword,
	EnvProjectName,
	EnvIdentityEndpoint).
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
				EnvDomainName:  "example.com",
				EnvUserName:    "user",
				EnvPassword:    "secret",
				EnvProjectName: "test",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvDomainName:  "",
				EnvUserName:    "",
				EnvPassword:    "",
				EnvProjectName: "",
			},
			expected: "otc: some credentials information are missing: OTC_DOMAIN_NAME,OTC_USER_NAME,OTC_PASSWORD,OTC_PROJECT_NAME",
		},
		{
			desc: "missing domain name",
			envVars: map[string]string{
				EnvDomainName:  "",
				EnvUserName:    "user",
				EnvPassword:    "secret",
				EnvProjectName: "test",
			},
			expected: "otc: some credentials information are missing: OTC_DOMAIN_NAME",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvDomainName:  "example.com",
				EnvUserName:    "",
				EnvPassword:    "secret",
				EnvProjectName: "test",
			},
			expected: "otc: some credentials information are missing: OTC_USER_NAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvDomainName:  "example.com",
				EnvUserName:    "user",
				EnvPassword:    "",
				EnvProjectName: "test",
			},
			expected: "otc: some credentials information are missing: OTC_PASSWORD",
		},
		{
			desc: "missing project name",
			envVars: map[string]string{
				EnvDomainName:  "example.com",
				EnvUserName:    "user",
				EnvPassword:    "secret",
				EnvProjectName: "",
			},
			expected: "otc: some credentials information are missing: OTC_PROJECT_NAME",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		domainName  string
		projectName string
		username    string
		password    string
		expected    string
	}{
		{
			desc:        "success",
			domainName:  "example.com",
			projectName: "test",
			username:    "user",
			password:    "secret",
		},
		{
			desc:     "missing credentials",
			expected: "otc: credentials missing",
		},
		{
			desc:        "missing domain name",
			domainName:  "",
			projectName: "test",
			username:    "user",
			password:    "secret",
			expected:    "otc: credentials missing",
		},
		{
			desc:        "missing project name",
			domainName:  "example.com",
			projectName: "",
			username:    "user",
			password:    "secret",
			expected:    "otc: credentials missing",
		},
		{
			desc:        "missing username",
			domainName:  "example.com",
			projectName: "test",
			username:    "",
			password:    "secret",
			expected:    "otc: credentials missing",
		},
		{
			desc:        "missing password ",
			domainName:  "example.com",
			projectName: "test",
			username:    "user",
			password:    "",
			expected:    "otc: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.DomainName = test.domainName
			config.ProjectName = test.projectName
			config.UserName = test.username
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v2/zones",
			responseFromFixture("zones_GET.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.")).
		Route("/", servermock.DumpRequest()).
		Build(t)

	err := provider.Present("example.com", "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_emptyZone(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v2/zones",
			responseFromFixture("zones_GET_empty.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.")).
		Route("/", servermock.DumpRequest()).
		Build(t)

	err := provider.Present("example.com", "", "123d==")
	require.EqualError(t, err, "otc: unable to get zone: zone example.com. not found")
}

func TestDNSProvider_Cleanup(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v2/zones",
			responseFromFixture("zones_GET.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.")).
		Route("GET /v2/zones/123123/recordsets",
			responseFromFixture("zones-recordsets_GET.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge.example.com.").
				With("type", "TXT")).
		Route("DELETE /v2/zones/123123/recordsets/321321",
			responseFromFixture("zones-recordsets_DELETE.json")).
		Build(t)

	err := provider.CleanUp("example.com", "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_emptyRecordset(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v2/zones",
			responseFromFixture("zones_GET.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.")).
		Route("GET /v2/zones/123123/recordsets",
			responseFromFixture("zones-recordsets_GET_empty.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge.example.com.").
				With("type", "TXT")).
		Build(t)

	err := provider.CleanUp("example.com", "", "123d==")
	require.EqualError(t, err, "otc: unable to get record _acme-challenge.example.com. for zone example.com: record not found")
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.UserName = "user"
			config.Password = "secret"
			config.DomainName = "example.com"
			config.ProjectName = "test"
			config.IdentityEndpoint = fmt.Sprintf("%s/v3/auth/token", server.URL)

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().WithJSONHeaders(),
	).
		Route("POST /v3/auth/token", internal.IdentityHandlerMock())
}

func responseFromFixture(filename string) *servermock.ResponseFromFileHandler {
	return servermock.ResponseFromFile(path.Join("internal", "fixtures", filename))
}
