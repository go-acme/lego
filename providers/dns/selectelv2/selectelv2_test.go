package selectelv2

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsernameOS,
	EnvPasswordOS,
	EnvDomainName,
	EnvUserDomainName,
	EnvProjectID,
	EnvAuthRegion,
	EnvAuthURL,
).
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
				EnvUsernameOS: "someName",
				EnvPasswordOS: "qwerty",
				EnvDomainName: "1",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvPasswordOS: "qwerty",
				EnvDomainName: "1",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsernameOS: "someName",
				EnvDomainName: "1",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_PASSWORD",
		},
		{
			desc: "missing account",
			envVars: map[string]string{
				EnvUsernameOS: "someName",
				EnvPasswordOS: "qwerty",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_ACCOUNT_ID",
		},
		{
			desc: "missing project",
			envVars: map[string]string{
				EnvUsernameOS: "someName",
				EnvPasswordOS: "qwerty",
				EnvDomainName: "1",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_PROJECT_ID",
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
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.baseClient)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		username  string
		password  string
		account   string
		projectID string
		expected  string
	}{
		{
			desc:      "success",
			username:  "user",
			password:  "secret",
			account:   "1",
			projectID: "111a11111aaa11aa1a11aaa11111aa1a",
		},
		{
			desc:      "missing username",
			password:  "secret",
			account:   "1",
			projectID: "111a11111aaa11aa1a11aaa11111aa1a",
			expected:  "selectelv2: missing username",
		},
		{
			desc:      "missing password",
			username:  "user",
			account:   "1",
			projectID: "111a11111aaa11aa1a11aaa11111aa1a",
			expected:  "selectelv2: missing password",
		},
		{
			desc:      "missing account",
			username:  "user",
			password:  "secret",
			projectID: "111a11111aaa11aa1a11aaa11111aa1a",
			expected:  "selectelv2: missing account ID",
		},
		{
			desc:     "missing projectID",
			username: "user",
			password: "secret",
			account:  "1",
			expected: "selectelv2: missing project ID",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.DomainName = test.account
			config.ProjectID = test.projectID

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.baseClient)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
