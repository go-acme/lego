package designate

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const (
	envDomain             = envNamespace + "DOMAIN"
	envOSClientConfigFile = "OS_CLIENT_CONFIG_FILE"
)

var envTest = tester.NewEnvTest(
	EnvCloud,
	EnvAuthURL,
	EnvUsername,
	EnvPassword,
	EnvUserID,
	EnvAppCredID,
	EnvAppCredName,
	EnvAppCredSecret,
	EnvTenantName,
	EnvRegionName,
	EnvProjectID,
	envOSClientConfigFile).
	WithDomain(envDomain)

func TestNewDNSProvider_fromEnv(t *testing.T) {
	serverURL := setupTestProvider(t)

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAuthURL:    serverURL + "/v2.0/",
				EnvUsername:   "B",
				EnvPassword:   "C",
				EnvRegionName: "D",
				EnvProjectID:  "E",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAuthURL:    "",
				EnvUsername:   "",
				EnvPassword:   "",
				EnvRegionName: "",
			},
			expected: "designate: Missing environment variable [OS_AUTH_URL]",
		},
		{
			desc: "missing auth url",
			envVars: map[string]string{
				EnvAuthURL:    "",
				EnvUsername:   "B",
				EnvPassword:   "C",
				EnvRegionName: "D",
			},
			expected: "designate: Missing environment variable [OS_AUTH_URL]",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAuthURL:    serverURL + "/v2.0/",
				EnvUsername:   "",
				EnvPassword:   "C",
				EnvRegionName: "D",
			},
			expected: "designate: Missing one of the following environment variables [OS_USERID, OS_USERNAME]",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvAuthURL:    serverURL + "/v2.0/",
				EnvUsername:   "B",
				EnvPassword:   "",
				EnvRegionName: "D",
			},
			expected: "designate: Missing environment variable [OS_PASSWORD]",
		},
		{
			desc: "missing application credential secret",
			envVars: map[string]string{
				EnvAuthURL:    serverURL + "/v2.0/",
				EnvRegionName: "D",
				EnvAppCredID:  "F",
			},
			expected: "designate: Missing environment variable [OS_APPLICATION_CREDENTIAL_SECRET]",
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

func TestNewDNSProvider_fromCloud(t *testing.T) {
	serverURL := setupTestProvider(t)

	testCases := []struct {
		desc     string
		osCloud  string
		cloud    clientconfig.Cloud
		expected string
	}{
		{
			desc:    "success",
			osCloud: "good_cloud",
			cloud: clientconfig.Cloud{
				AuthInfo: &clientconfig.AuthInfo{
					AuthURL:     serverURL + "/v2.0/",
					Username:    "B",
					Password:    "C",
					ProjectName: "E",
					ProjectID:   "F",
				},
				RegionName: "D",
			},
		},
		{
			desc:    "missing auth url",
			osCloud: "missing_auth_url",
			cloud: clientconfig.Cloud{
				AuthInfo: &clientconfig.AuthInfo{
					Username:    "B",
					Password:    "C",
					ProjectName: "E",
					ProjectID:   "F",
				},
				RegionName: "D",
			},
			expected: "designate: Missing input for argument [auth_url]",
		},
		{
			desc:    "missing username",
			osCloud: "missing_username",
			cloud: clientconfig.Cloud{
				AuthInfo: &clientconfig.AuthInfo{
					AuthURL:     serverURL + "/v2.0/",
					Password:    "C",
					ProjectName: "E",
					ProjectID:   "F",
				},
				RegionName: "D",
			},
			expected: "designate: failed to authenticate: Missing input for argument [Username]",
		},
		{
			desc:    "missing password",
			osCloud: "missing_auth_url",
			cloud: clientconfig.Cloud{
				AuthInfo: &clientconfig.AuthInfo{
					AuthURL:     serverURL + "/v2.0/",
					Username:    "B",
					ProjectName: "E",
					ProjectID:   "F",
				},
				RegionName: "D",
			},
			expected: "designate: failed to authenticate: Exactly one of PasswordCredentials and TokenCredentials must be provided",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(map[string]string{
				EnvCloud:              test.osCloud,
				envOSClientConfigFile: createCloudsYaml(t, test.osCloud, test.cloud),
			})

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
	serverURL := setupTestProvider(t)

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
			authURL:    serverURL + "/v2.0/",
		},
		{
			desc:       "wrong auth url",
			tenantName: "A",
			password:   "B",
			userName:   "C",
			authURL:    serverURL,
			expected:   "designate: failed to authenticate: No supported version available from endpoint " + serverURL + "/",
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

// createCloudsYaml creates a temporary cloud file for testing purpose.
func createCloudsYaml(t *testing.T, cloudName string, cloud clientconfig.Cloud) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "lego_test")
	require.NoError(t, err)

	t.Cleanup(func() { _ = file.Close() })

	clouds := clientconfig.Clouds{
		Clouds: map[string]clientconfig.Cloud{
			cloudName: cloud,
		},
	}

	err = yaml.NewEncoder(file).Encode(&clouds)
	require.NoError(t, err)

	return file.Name()
}

func setupTestProvider(t *testing.T) string {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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
		w.WriteHeader(http.StatusOK)
	})

	return server.URL
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
