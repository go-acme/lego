package bluecat

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest          bool
	envTestServer     string
	envTestUserName   string
	envTestPassword   string
	envTestConfigName string
	envTestDNSView    string
	envTestDomain     string
)

func init() {
	envTestServer = os.Getenv("BLUECAT_SERVER_URL")
	envTestUserName = os.Getenv("BLUECAT_USER_NAME")
	envTestPassword = os.Getenv("BLUECAT_PASSWORD")
	envTestDomain = os.Getenv("BLUECAT_DOMAIN")
	envTestConfigName = os.Getenv("BLUECAT_CONFIG_NAME")
	envTestDNSView = os.Getenv("BLUECAT_DNS_VIEW")

	if len(envTestServer) > 0 &&
		len(envTestDomain) > 0 &&
		len(envTestUserName) > 0 &&
		len(envTestPassword) > 0 &&
		len(envTestConfigName) > 0 &&
		len(envTestDNSView) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("BLUECAT_SERVER_URL", envTestServer)
	os.Setenv("BLUECAT_USER_NAME", envTestUserName)
	os.Setenv("BLUECAT_PASSWORD", envTestPassword)
	os.Setenv("BLUECAT_CONFIG_NAME", envTestConfigName)
	os.Setenv("BLUECAT_DNS_VIEW", envTestDNSView)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"BLUECAT_SERVER_URL":  "http://localhost",
				"BLUECAT_USER_NAME":   "A",
				"BLUECAT_PASSWORD":    "B",
				"BLUECAT_CONFIG_NAME": "C",
				"BLUECAT_DNS_VIEW":    "D",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"BLUECAT_SERVER_URL":  "",
				"BLUECAT_USER_NAME":   "",
				"BLUECAT_PASSWORD":    "",
				"BLUECAT_CONFIG_NAME": "",
				"BLUECAT_DNS_VIEW":    "",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_SERVER_URL,BLUECAT_USER_NAME,BLUECAT_PASSWORD,BLUECAT_CONFIG_NAME,BLUECAT_DNS_VIEW",
		},
		{
			desc: "missing server url",
			envVars: map[string]string{
				"BLUECAT_SERVER_URL":  "",
				"BLUECAT_USER_NAME":   "A",
				"BLUECAT_PASSWORD":    "B",
				"BLUECAT_CONFIG_NAME": "C",
				"BLUECAT_DNS_VIEW":    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_SERVER_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				"BLUECAT_SERVER_URL":  "http://localhost",
				"BLUECAT_USER_NAME":   "",
				"BLUECAT_PASSWORD":    "B",
				"BLUECAT_CONFIG_NAME": "C",
				"BLUECAT_DNS_VIEW":    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_USER_NAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				"BLUECAT_SERVER_URL":  "http://localhost",
				"BLUECAT_USER_NAME":   "A",
				"BLUECAT_PASSWORD":    "",
				"BLUECAT_CONFIG_NAME": "C",
				"BLUECAT_DNS_VIEW":    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_PASSWORD",
		},
		{
			desc: "missing config name",
			envVars: map[string]string{
				"BLUECAT_SERVER_URL":  "http://localhost",
				"BLUECAT_USER_NAME":   "A",
				"BLUECAT_PASSWORD":    "B",
				"BLUECAT_CONFIG_NAME": "",
				"BLUECAT_DNS_VIEW":    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_CONFIG_NAME",
		},
		{
			desc: "missing DNS view",
			envVars: map[string]string{
				"BLUECAT_SERVER_URL":  "http://localhost",
				"BLUECAT_USER_NAME":   "A",
				"BLUECAT_PASSWORD":    "B",
				"BLUECAT_CONFIG_NAME": "C",
				"BLUECAT_DNS_VIEW":    "",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_DNS_VIEW",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			for key, value := range test.envVars {
				if len(value) == 0 {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

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
	testCases := []struct {
		desc       string
		baseURL    string
		userName   string
		password   string
		configName string
		dnsView    string
		expected   string
	}{
		{
			desc:       "success",
			baseURL:    "http://localhost",
			userName:   "A",
			password:   "B",
			configName: "C",
			dnsView:    "D",
		},
		{
			desc:     "missing credentials",
			expected: "bluecat: credentials missing",
		},
		{
			desc:       "missing base URL",
			baseURL:    "",
			userName:   "A",
			password:   "B",
			configName: "C",
			dnsView:    "D",
			expected:   "bluecat: credentials missing",
		},
		{
			desc:       "missing username",
			baseURL:    "http://localhost",
			userName:   "",
			password:   "B",
			configName: "C",
			dnsView:    "D",
			expected:   "bluecat: credentials missing",
		},
		{
			desc:       "missing password",
			baseURL:    "http://localhost",
			userName:   "A",
			password:   "",
			configName: "C",
			dnsView:    "D",
			expected:   "bluecat: credentials missing",
		},
		{
			desc:       "missing config name",
			baseURL:    "http://localhost",
			userName:   "A",
			password:   "B",
			configName: "",
			dnsView:    "D",
			expected:   "bluecat: credentials missing",
		},
		{
			desc:       "missing DNS view",
			baseURL:    "http://localhost",
			userName:   "A",
			password:   "B",
			configName: "C",
			dnsView:    "",
			expected:   "bluecat: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("BLUECAT_SERVER_URL")
			os.Unsetenv("BLUECAT_USER_NAME")
			os.Unsetenv("BLUECAT_PASSWORD")
			os.Unsetenv("BLUECAT_CONFIG_NAME")
			os.Unsetenv("BLUECAT_DNS_VIEW")

			config := NewDefaultConfig()
			config.BaseURL = test.baseURL
			config.UserName = test.userName
			config.Password = test.password
			config.ConfigName = test.configName
			config.DNSView = test.dnsView

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

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(time.Second * 1)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
