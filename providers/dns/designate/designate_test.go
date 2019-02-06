package designate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/platform/tester"
)

var envTest = tester.NewEnvTest(
	"OS_AUTH_URL",
	"OS_USERNAME",
	"OS_PASSWORD",
	"OS_TENANT_NAME",
	"OS_REGION_NAME").
	WithDomain("DESIGNATE_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	server := getServer()
	defer server.Close()

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"OS_AUTH_URL":    server.URL + "/v2.0/",
				"OS_USERNAME":    "B",
				"OS_PASSWORD":    "C",
				"OS_REGION_NAME": "D",
				"OS_TENANT_NAME": "E",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"OS_AUTH_URL":    "",
				"OS_USERNAME":    "",
				"OS_PASSWORD":    "",
				"OS_REGION_NAME": "",
				"OS_TENANT_NAME": "",
			},
			expected: "designate: some credentials information are missing: OS_AUTH_URL,OS_USERNAME,OS_PASSWORD,OS_TENANT_NAME,OS_REGION_NAME",
		},
		{
			desc: "missing auth url",
			envVars: map[string]string{
				"OS_AUTH_URL":    "",
				"OS_USERNAME":    "B",
				"OS_PASSWORD":    "C",
				"OS_REGION_NAME": "D",
				"OS_TENANT_NAME": "E",
			},
			expected: "designate: some credentials information are missing: OS_AUTH_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				"OS_AUTH_URL":    server.URL + "/v2.0/",
				"OS_USERNAME":    "",
				"OS_PASSWORD":    "C",
				"OS_REGION_NAME": "D",
				"OS_TENANT_NAME": "E",
			},
			expected: "designate: some credentials information are missing: OS_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				"OS_AUTH_URL":    server.URL + "/v2.0/",
				"OS_USERNAME":    "B",
				"OS_PASSWORD":    "",
				"OS_REGION_NAME": "D",
				"OS_TENANT_NAME": "E",
			},
			expected: "designate: some credentials information are missing: OS_PASSWORD",
		},
		{
			desc: "missing region name",
			envVars: map[string]string{
				"OS_AUTH_URL":    server.URL + "/v2.0/",
				"OS_USERNAME":    "B",
				"OS_PASSWORD":    "C",
				"OS_REGION_NAME": "",
				"OS_TENANT_NAME": "E",
			},
			expected: "designate: some credentials information are missing: OS_REGION_NAME",
		},
		{
			desc: "missing tenant name",
			envVars: map[string]string{
				"OS_AUTH_URL":    server.URL + "/v2.0/",
				"OS_USERNAME":    "B",
				"OS_PASSWORD":    "C",
				"OS_REGION_NAME": "D",
				"OS_TENANT_NAME": "",
			},
			expected: "designate: some credentials information are missing: OS_TENANT_NAME",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
	server := getServer()
	defer server.Close()

	testCases := []struct {
		desc       string
		tenantName string
		password   string
		userName   string
		authURL    string
		expected   string
	}{
		{
			desc:       "success",
			tenantName: "A",
			password:   "B",
			userName:   "C",
			authURL:    server.URL + "/v2.0/",
		},
		{
			desc:       "wrong auth url",
			tenantName: "A",
			password:   "B",
			userName:   "C",
			authURL:    server.URL,
			expected:   "designate: failed to authenticate: No supported version available from endpoint " + server.URL + "/",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.opts.TenantName = test.tenantName
			config.opts.Password = test.password
			config.opts.Username = test.userName
			config.opts.IdentityEndpoint = test.authURL

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func getServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
	"access": {
		"token": {
			"id": "a",
			"expires": "9015-06-05T16:24:57.637Z"
		},
		"user": {
			"name": "a",
			"roles": [ ],
			"role_links": [ ] 
		},
		"serviceCatalog": [
			{
				"endpoints": [
					{
						"adminURL": "http://23.253.72.207:9696/",
						"region": "D",
						"internalURL": "http://23.253.72.207:9696/",
						"id": "97c526db8d7a4c88bbb8d68db1bdcdb8",
						"publicURL": "http://23.253.72.207:9696/"
					}
				],
				"endpoints_links": [ ],
				"type": "dns",
				"name": "designate"
			}
		]
	}
}`))
		w.WriteHeader(200)
	})
	return httptest.NewServer(mux)
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
