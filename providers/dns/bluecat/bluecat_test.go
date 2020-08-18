package bluecat

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvServerURL,
	EnvUserName,
	EnvPassword,
	EnvConfigName,
	EnvDNSView).
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
				EnvServerURL:  "http://localhost",
				EnvUserName:   "A",
				EnvPassword:   "B",
				EnvConfigName: "C",
				EnvDNSView:    "D",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvServerURL:  "",
				EnvUserName:   "",
				EnvPassword:   "",
				EnvConfigName: "",
				EnvDNSView:    "",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_SERVER_URL,BLUECAT_USER_NAME,BLUECAT_PASSWORD,BLUECAT_CONFIG_NAME,BLUECAT_DNS_VIEW",
		},
		{
			desc: "missing server url",
			envVars: map[string]string{
				EnvServerURL:  "",
				EnvUserName:   "A",
				EnvPassword:   "B",
				EnvConfigName: "C",
				EnvDNSView:    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_SERVER_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvServerURL:  "http://localhost",
				EnvUserName:   "",
				EnvPassword:   "B",
				EnvConfigName: "C",
				EnvDNSView:    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_USER_NAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvServerURL:  "http://localhost",
				EnvUserName:   "A",
				EnvPassword:   "",
				EnvConfigName: "C",
				EnvDNSView:    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_PASSWORD",
		},
		{
			desc: "missing config name",
			envVars: map[string]string{
				EnvServerURL:  "http://localhost",
				EnvUserName:   "A",
				EnvPassword:   "B",
				EnvConfigName: "",
				EnvDNSView:    "D",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_CONFIG_NAME",
		},
		{
			desc: "missing DNS view",
			envVars: map[string]string{
				EnvServerURL:  "http://localhost",
				EnvUserName:   "A",
				EnvPassword:   "B",
				EnvConfigName: "C",
				EnvDNSView:    "",
			},
			expected: "bluecat: some credentials information are missing: BLUECAT_DNS_VIEW",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider(nil)

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
			config := NewDefaultConfig(nil)
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
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
