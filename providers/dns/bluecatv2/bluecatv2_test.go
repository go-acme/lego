package bluecatv2

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/bluecatv2/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvServerURL,
	EnvUsername,
	EnvPassword,
	EnvConfigName,
	EnvViewName,
	EnvSkipDeploy,
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
				EnvServerURL:  "https://example.com/",
				EnvUsername:   "userA",
				EnvPassword:   "secret",
				EnvConfigName: "myConfig",
				EnvViewName:   "myView",
			},
		},
		{
			desc: "missing server URL",
			envVars: map[string]string{
				EnvServerURL:  "",
				EnvUsername:   "userA",
				EnvPassword:   "secret",
				EnvConfigName: "myConfig",
				EnvViewName:   "myView",
			},
			expected: "bluecatv2: some credentials information are missing: BLUECATV2_SERVER_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvServerURL:  "https://example.com/",
				EnvUsername:   "",
				EnvPassword:   "secret",
				EnvConfigName: "myConfig",
				EnvViewName:   "myView",
			},
			expected: "bluecatv2: some credentials information are missing: BLUECATV2_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvServerURL:  "https://example.com/",
				EnvUsername:   "userA",
				EnvPassword:   "",
				EnvConfigName: "myConfig",
				EnvViewName:   "myView",
			},
			expected: "bluecatv2: some credentials information are missing: BLUECATV2_PASSWORD",
		},
		{
			desc: "missing configuration name",
			envVars: map[string]string{
				EnvServerURL:  "https://example.com/",
				EnvUsername:   "userA",
				EnvPassword:   "secret",
				EnvConfigName: "",
				EnvViewName:   "myView",
			},
			expected: "bluecatv2: some credentials information are missing: BLUECATV2_CONFIG_NAME",
		},
		{
			desc: "missing view name",
			envVars: map[string]string{
				EnvServerURL:  "https://example.com/",
				EnvUsername:   "userA",
				EnvPassword:   "secret",
				EnvConfigName: "myConfig",
				EnvViewName:   "",
			},
			expected: "bluecatv2: some credentials information are missing: BLUECATV2_VIEW_NAME",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "bluecatv2: some credentials information are missing: BLUECATV2_SERVER_URL,BLUECATV2_USERNAME,BLUECATV2_PASSWORD,BLUECATV2_CONFIG_NAME,BLUECATV2_VIEW_NAME",
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
		desc       string
		serverURL  string
		username   string
		password   string
		configName string
		viewName   string
		expected   string
	}{
		{
			desc:       "success",
			serverURL:  "https://example.com/",
			username:   "userA",
			password:   "secret",
			configName: "myConfig",
			viewName:   "myView",
		},
		{
			desc:       "missing server URL",
			username:   "userA",
			password:   "secret",
			configName: "myConfig",
			viewName:   "myView",
			expected:   "bluecatv2: missing server URL",
		},
		{
			desc:       "missing username",
			serverURL:  "https://example.com/",
			password:   "secret",
			configName: "myConfig",
			viewName:   "myView",
			expected:   "bluecatv2: credentials missing",
		},
		{
			desc:       "missing password",
			serverURL:  "https://example.com/",
			username:   "userA",
			configName: "myConfig",
			viewName:   "myView",
			expected:   "bluecatv2: credentials missing",
		},
		{
			desc:      "missing configuration name",
			serverURL: "https://example.com/",
			username:  "userA",
			password:  "secret",
			viewName:  "myView",
			expected:  "bluecatv2: missing configuration name",
		},
		{
			desc:       "missing view name",
			serverURL:  "https://example.com/",
			username:   "userA",
			password:   "secret",
			configName: "myConfig",
			expected:   "bluecatv2: missing view name",
		},
		{
			desc:     "missing credentials",
			expected: "bluecatv2: missing server URL",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ServerURL = test.serverURL
			config.Username = test.username
			config.Password = test.password
			config.ConfigName = test.configName
			config.ViewName = test.viewName

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

			config.ServerURL = server.URL
			config.Username = "userA"
			config.Password = "secret"
			config.ConfigName = "myConfiguration"
			config.ViewName = "myView"

			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api/v2/sessions",
			servermock.ResponseFromInternal("postSession.json"),
			servermock.CheckRequestJSONBodyFromInternal("postSession-request.json"),
		).
		Route("GET /api/v2/configurations",
			servermock.ResponseFromInternal("configurations.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "name:eq('myConfiguration')"),
		).
		Route("GET /api/v2/configurations/12345/views",
			servermock.ResponseFromInternal("views.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "name:eq('myView')"),
		).
		Route("GET /api/v2/zones",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				filter := req.URL.Query().Get("filter")

				if strings.Contains(filter, internal.Eq("absoluteName", "example.com").String()) {
					servermock.ResponseFromInternal("zones.json").ServeHTTP(rw, req)

					return
				}

				servermock.ResponseFromInternal("error.json").
					WithStatusCode(http.StatusNotFound).ServeHTTP(rw, req)
			}),
		).
		Route("POST /api/v2/zones/12345/resourceRecords",
			servermock.ResponseFromInternal("postZoneResourceRecord.json"),
			servermock.CheckRequestJSONBodyFromInternal("postZoneResourceRecord-request.json"),
		).
		Route("POST /api/v2/zones/12345/deployments",
			servermock.ResponseFromInternal("postZoneDeployment.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("postZoneDeployment-request.json"),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_skipDeploy(t *testing.T) {
	defer envTest.RestoreEnv()

	envTest.ClearEnv()

	envTest.Apply(map[string]string{
		EnvSkipDeploy: "true",
	})

	provider := mockBuilder().
		Route("POST /api/v2/sessions",
			servermock.ResponseFromInternal("postSession.json"),
			servermock.CheckRequestJSONBodyFromInternal("postSession-request.json"),
		).
		Route("GET /api/v2/configurations",
			servermock.ResponseFromInternal("configurations.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "name:eq('myConfiguration')"),
		).
		Route("GET /api/v2/configurations/12345/views",
			servermock.ResponseFromInternal("views.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "name:eq('myView')"),
		).
		Route("GET /api/v2/zones",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				filter := req.URL.Query().Get("filter")

				if strings.Contains(filter, internal.Eq("absoluteName", "example.com").String()) {
					servermock.ResponseFromInternal("zones.json").ServeHTTP(rw, req)

					return
				}

				servermock.ResponseFromInternal("error.json").
					WithStatusCode(http.StatusNotFound).ServeHTTP(rw, req)
			}),
		).
		Route("POST /api/v2/zones/12345/resourceRecords",
			servermock.ResponseFromInternal("postZoneResourceRecord.json"),
			servermock.CheckRequestJSONBodyFromInternal("postZoneResourceRecord-request.json"),
		).
		Route("POST /api/v2/zones/456789/deployments",
			servermock.Noop().
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api/v2/sessions",
			servermock.ResponseFromInternal("postSession.json"),
			servermock.CheckRequestJSONBodyFromInternal("postSession-request.json"),
		).
		Route("DELETE /api/v2/resourceRecords/12345",
			servermock.ResponseFromInternal("deleteResourceRecord.json"),
		).
		Route("POST /api/v2/zones/456789/deployments",
			servermock.ResponseFromInternal("postZoneDeployment.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("postZoneDeployment-request.json"),
		).
		Build(t)

	provider.zoneIDs["abc"] = 456789
	provider.recordIDs["abc"] = 12345

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_skipDeploy(t *testing.T) {
	defer envTest.RestoreEnv()

	envTest.ClearEnv()

	envTest.Apply(map[string]string{
		EnvSkipDeploy: "true",
	})

	provider := mockBuilder().
		Route("POST /api/v2/sessions",
			servermock.ResponseFromInternal("postSession.json"),
			servermock.CheckRequestJSONBodyFromInternal("postSession-request.json"),
		).
		Route("DELETE /api/v2/resourceRecords/12345",
			servermock.ResponseFromInternal("deleteResourceRecord.json"),
		).
		Route("POST /api/v2/zones/456789/deployments",
			servermock.Noop().
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	provider.zoneIDs["abc"] = 456789
	provider.recordIDs["abc"] = 12345

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
