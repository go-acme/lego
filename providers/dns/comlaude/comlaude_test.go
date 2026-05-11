package comlaude

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvPassword,
	EnvAPIKey,
	EnvGroupID,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
				EnvAPIKey:   "key",
				EnvGroupID:  "grp1",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "secret",
				EnvAPIKey:   "key",
				EnvGroupID:  "grp1",
			},
			expected: "comlaude: some credentials information are missing: COMLAUDE_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "",
				EnvAPIKey:   "key",
				EnvGroupID:  "grp1",
			},
			expected: "comlaude: some credentials information are missing: COMLAUDE_PASSWORD",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
				EnvAPIKey:   "",
				EnvGroupID:  "grp1",
			},
			expected: "comlaude: some credentials information are missing: COMLAUDE_API_KEY",
		},
		{
			desc: "missing group ID",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
				EnvAPIKey:   "key",
				EnvGroupID:  "",
			},
			expected: "comlaude: some credentials information are missing: COMLAUDE_GROUP_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "comlaude: some credentials information are missing: COMLAUDE_USERNAME,COMLAUDE_PASSWORD,COMLAUDE_API_KEY,COMLAUDE_GROUP_ID",
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
		username string
		password string
		apiKey   string
		groupID  string
		expected string
	}{
		{
			desc:     "success",
			username: "user",
			password: "secret",
			apiKey:   "key",
			groupID:  "grp1",
		},
		{
			desc:     "missing username",
			password: "secret",
			apiKey:   "key",
			groupID:  "grp1",
			expected: "comlaude: credentials missing",
		},
		{
			desc:     "missing password",
			username: "user",
			apiKey:   "key",
			groupID:  "grp1",
			expected: "comlaude: credentials missing",
		},
		{
			desc:     "missing apiKey",
			username: "user",
			password: "secret",
			groupID:  "grp1",
			expected: "comlaude: credentials missing",
		},
		{
			desc:     "missing groupID",
			username: "user",
			password: "secret",
			apiKey:   "key",
			expected: "comlaude: group ID missing",
		},
		{
			desc:     "missing credentials",
			expected: "comlaude: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.APIKey = test.apiKey
			config.GroupID = test.groupID

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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Username = "user"
			config.Password = "secret"
			config.APIKey = "key"
			config.GroupID = "grp1"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)
			p.identifier.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api_login",
			servermock.ResponseFromInternal("api_login.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("username", "user").
				With("password", "secret").
				With("api_key", "key"),
		).
		Route("GET /groups/grp1/domains",
			servermock.ResponseFromInternal("domains_get.json"),
			servermock.CheckHeader().WithAuthorization("Bearer xxx"),
			servermock.CheckQueryParameter().Strict().
				With("filter[name]", "*.example.com"),
		).
		Route("POST /groups/grp1/zones/62b873d3-a31c-4921-a309-548810913c4f/records",
			servermock.ResponseFromInternal("record_create.json"),
			servermock.CheckHeader().WithAuthorization("Bearer xxx"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().
				With("name", "_acme-challenge").
				With("ttl", "120").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("type", "TXT"),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)

	require.Len(t, provider.zoneIDs, 1)
	require.Len(t, provider.recordIDs, 1)

	assert.Equal(t, "62b873d3-a31c-4921-a309-548810913c4f", provider.zoneIDs["abc"])
	assert.Equal(t, "8a746001-d319-4583-bfb6-ae8aacc628aa", provider.recordIDs["abc"])
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api_login",
			servermock.ResponseFromInternal("api_login.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("username", "user").
				With("password", "secret").
				With("api_key", "key"),
		).
		Route("DELETE /groups/grp1/zones/62b873d3-a31c-4921-a309-548810913c4f/records/8a746001-d319-4583-bfb6-ae8aacc628aa",
			servermock.ResponseFromInternal("record_delete.json"),
			servermock.CheckHeader().WithAuthorization("Bearer xxx"),
		).
		Build(t)

	token := "abc"

	provider.zoneIDs["abc"] = "62b873d3-a31c-4921-a309-548810913c4f"
	provider.recordIDs["abc"] = "8a746001-d319-4583-bfb6-ae8aacc628aa"

	err := provider.CleanUp("example.com", token, "123d==")
	require.NoError(t, err)
}
