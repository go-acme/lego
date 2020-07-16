package designate

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAuthURL,
	EnvUsername,
	EnvPassword,
	EnvTenantName,
	EnvRegionName,
	EnvProjectID).
	WithDomain(envDomain)

func TestNewDNSProviderFromEnv(t *testing.T) {
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
				EnvAuthURL:    server.URL + "/v2.0/",
				EnvUsername:   "B",
				EnvPassword:   "C",
				EnvRegionName: "D",
				EnvTenantName: "E",
				EnvProjectID:  "F",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAuthURL:    "",
				EnvUsername:   "",
				EnvPassword:   "",
				EnvRegionName: "",
				EnvTenantName: "",
			},
			expected: "designate: some credentials information are missing: OS_AUTH_URL,OS_USERNAME,OS_PASSWORD,OS_TENANT_NAME,OS_REGION_NAME",
		},
		{
			desc: "missing auth url",
			envVars: map[string]string{
				EnvAuthURL:    "",
				EnvUsername:   "B",
				EnvPassword:   "C",
				EnvRegionName: "D",
				EnvTenantName: "E",
			},
			expected: "designate: some credentials information are missing: OS_AUTH_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAuthURL:    server.URL + "/v2.0/",
				EnvUsername:   "",
				EnvPassword:   "C",
				EnvRegionName: "D",
				EnvTenantName: "E",
			},
			expected: "designate: some credentials information are missing: OS_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvAuthURL:    server.URL + "/v2.0/",
				EnvUsername:   "B",
				EnvPassword:   "",
				EnvRegionName: "D",
				EnvTenantName: "E",
			},
			expected: "designate: some credentials information are missing: OS_PASSWORD",
		},
		{
			desc: "missing region name",
			envVars: map[string]string{
				EnvAuthURL:    server.URL + "/v2.0/",
				EnvUsername:   "B",
				EnvPassword:   "C",
				EnvRegionName: "",
				EnvTenantName: "E",
			},
			expected: "designate: some credentials information are missing: OS_REGION_NAME",
		},
		{
			desc: "missing tenant name",
			envVars: map[string]string{
				EnvAuthURL:    server.URL + "/v2.0/",
				EnvUsername:   "B",
				EnvPassword:   "C",
				EnvRegionName: "D",
				EnvTenantName: "",
			},
			expected: "designate: some credentials information are missing: OS_TENANT_NAME",
		},
	}

	// unset the OS_CLOUD environment variable otherwise the information will be read from
	// the clouds.yaml
	var cloud string
	if vals, err := env.Get(EnvCloud); err == nil {
		cloud = vals[EnvCloud]
		_ = os.Unsetenv(EnvCloud)
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

	if cloud != "" {
		_ = os.Setenv(EnvCloud, cloud)
	}
}

// Create a temporary cloud file for testing purpose.
func createCloudsYaml(cloudName string, cloud clientconfig.Cloud) (string, error) {
	file, err := ioutil.TempFile("", "lego_test")
	if err != nil {
		return "", err
	}

	clouds := clientconfig.Clouds{
		Clouds: map[string]clientconfig.Cloud{
			cloudName: cloud,
		},
	}

	d, err := yaml.Marshal(&clouds)
	if err != nil {
		return "", err
	}

	_, err = file.Write(d)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func TestNewDNSProviderFromCloud(t *testing.T) {
	server := getServer()
	defer server.Close()
	configFileEnv := "OS_CLIENT_CONFIG_FILE"

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
					AuthURL:     server.URL + "/v2.0/",
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
					AuthURL:     server.URL + "/v2.0/",
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
					AuthURL:     server.URL + "/v2.0/",
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

			f, err := createCloudsYaml(test.osCloud, test.cloud)
			require.NoError(t, err)

			_ = os.Setenv(EnvCloud, test.osCloud)
			_ = os.Setenv(configFileEnv, f)
			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}

			_ = os.Unsetenv(EnvCloud)
			_ = os.Unsetenv(configFileEnv)
			_ = os.Remove(f)
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
