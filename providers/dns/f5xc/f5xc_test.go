package f5xc

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvToken, EnvTenantName, EnvGroupName).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvToken:      "secret",
				EnvTenantName: "shortname",
				EnvGroupName:  "group",
			},
		},
		{
			desc: "missing API token",
			envVars: map[string]string{
				EnvToken:      "",
				EnvTenantName: "shortname",
				EnvGroupName:  "group",
			},
			expected: "f5xc: some credentials information are missing: F5XC_API_TOKEN",
		},
		{
			desc: "missing tenant name",
			envVars: map[string]string{
				EnvToken:      "secret",
				EnvTenantName: "",
				EnvGroupName:  "group",
			},
			expected: "f5xc: some credentials information are missing: F5XC_TENANT_NAME",
		},
		{
			desc: "missing group name",
			envVars: map[string]string{
				EnvToken:      "secret",
				EnvTenantName: "shortname",
				EnvGroupName:  "",
			},
			expected: "f5xc: some credentials information are missing: F5XC_GROUP_NAME",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "f5xc: some credentials information are missing: F5XC_API_TOKEN,F5XC_TENANT_NAME,F5XC_GROUP_NAME",
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
		apiToken   string
		tenantName string
		groupName  string
		expected   string
	}{
		{
			desc:       "success",
			apiToken:   "secret",
			tenantName: "shortname",
			groupName:  "group",
		},
		{
			desc:       "missing API token",
			tenantName: "shortname",
			groupName:  "group",
			expected:   "f5xc: credentials missing",
		},
		{
			desc:      "missing tenant name",
			apiToken:  "secret",
			groupName: "group",
			expected:  "f5xc: missing tenant name",
		},
		{
			desc:       "missing group name",
			apiToken:   "secret",
			tenantName: "shortname",
			expected:   "f5xc: missing group name",
		},
		{
			desc:     "missing credentials",
			expected: "f5xc: missing group name",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIToken = test.apiToken
			config.TenantName = test.tenantName
			config.GroupName = test.groupName

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
