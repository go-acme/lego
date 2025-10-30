package yandex360

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvOAuthToken, EnvOrgID).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvOAuthToken: "secret",
				EnvOrgID:      "123456",
			},
		},
		{
			desc: "missing org ID",
			envVars: map[string]string{
				EnvOAuthToken: "secret",
			},
			expected: "yandex360: some credentials information are missing: YANDEX360_ORG_ID",
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvOrgID: "123456",
			},
			expected: "yandex360: some credentials information are missing: YANDEX360_OAUTH_TOKEN",
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
		desc       string
		oauthToken string
		orgID      int64
		expected   string
	}{
		{
			desc:       "success",
			oauthToken: "secret",
			orgID:      123456,
		},
		{
			desc:       "missing org ID",
			oauthToken: "secret",
			expected:   "yandex360: orgID is required",
		},
		{
			desc:     "missing token",
			orgID:    123456,
			expected: "yandex360: OAuth token is required",
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.OAuthToken = test.oauthToken
			config.OrgID = test.orgID

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
