package vkcloud

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

const (
	fakeProjectID = "an_project_id_from_vk_cloud_ui"
	fakeUsername  = "vkclouduser@email.address"
	fakePasswd    = "vkcloudpasswd"
)

var envTest = tester.NewEnvTest(EnvProjectID, EnvUsername, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvProjectID: fakeProjectID,
				EnvUsername:  fakeUsername,
				EnvPassword:  fakePasswd,
			},
		},
		{
			desc: "missing project id",
			envVars: map[string]string{
				EnvUsername: fakeUsername,
				EnvPassword: fakePasswd,
			},
			expected: "vkcloud: some credentials information are missing: VK_CLOUD_PROJECT_ID",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvProjectID: fakeProjectID,
				EnvPassword:  fakePasswd,
			},
			expected: "vkcloud: some credentials information are missing: VK_CLOUD_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvProjectID: fakeProjectID,
				EnvUsername:  fakeUsername,
			},
			expected: "vkcloud: some credentials information are missing: VK_CLOUD_PASSWORD",
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
		desc     string
		config   *Config
		expected string
	}{
		{
			desc: "success",
			config: &Config{
				ProjectID:        fakeProjectID,
				Username:         fakeUsername,
				Password:         fakePasswd,
				DNSEndpoint:      defaultDNSEndpoint,
				IdentityEndpoint: defaultIdentityEndpoint,
				DomainName:       defaultDomainName,
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "vkcloud: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing project id",
			config: &Config{
				Username:         fakeUsername,
				Password:         fakePasswd,
				DNSEndpoint:      defaultDNSEndpoint,
				IdentityEndpoint: defaultIdentityEndpoint,
				DomainName:       defaultDomainName,
			},
			expected: "vkcloud: unable to build VK Cloud client: project id is missing in credentials information",
		},
		{
			desc: "missing username",
			config: &Config{
				ProjectID:        fakeProjectID,
				Password:         fakePasswd,
				DNSEndpoint:      defaultDNSEndpoint,
				IdentityEndpoint: defaultIdentityEndpoint,
				DomainName:       defaultDomainName,
			},
			expected: "vkcloud: unable to build VK Cloud client: username is missing in credentials information",
		},
		{
			desc: "missing password",
			config: &Config{
				ProjectID:        fakeProjectID,
				Username:         fakeUsername,
				DNSEndpoint:      defaultDNSEndpoint,
				IdentityEndpoint: defaultIdentityEndpoint,
				DomainName:       defaultDomainName,
			},
			expected: "vkcloud: unable to build VK Cloud client: password is missing in credentials information",
		},
		{
			desc: "missing dns endpoint",
			config: &Config{
				ProjectID:        fakeProjectID,
				Username:         fakeUsername,
				Password:         fakePasswd,
				IdentityEndpoint: defaultIdentityEndpoint,
				DomainName:       defaultDomainName,
			},
			expected: "vkcloud: DNS endpoint is missing in config",
		},
		{
			desc: "missing identity endpoint",
			config: &Config{
				ProjectID:   fakeProjectID,
				Username:    fakeUsername,
				Password:    fakePasswd,
				DNSEndpoint: defaultDNSEndpoint,
				DomainName:  defaultDomainName,
			},
			expected: "vkcloud: unable to build VK Cloud client: identity endpoint is missing in config",
		},
		{
			desc: "missing domain name",
			config: &Config{
				ProjectID:        fakeProjectID,
				Username:         fakeUsername,
				Password:         fakePasswd,
				DNSEndpoint:      defaultDNSEndpoint,
				IdentityEndpoint: defaultIdentityEndpoint,
			},
			expected: "vkcloud: unable to build VK Cloud client: domain name is missing in config",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
