package conohav3

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvTenantID,
	EnvAPIUserID,
	EnvAPIPassword).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "complete credentials, but login failed",
			envVars: map[string]string{
				EnvTenantID:    "tenant_id",
				EnvAPIUserID:   "api_user_id",
				EnvAPIPassword: "api_password",
			},
			expected: `conohav3: failed to log in: unexpected status code: [status code: 400] body: {"code": 400, "error": "user does not exist"}`,
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvTenantID:    "",
				EnvAPIUserID:   "",
				EnvAPIPassword: "",
			},
			expected: "conohav3: some credentials information are missing: CONOHAV3_TENANT_ID,CONOHAV3_API_USER_ID,CONOHAV3_API_PASSWORD",
		},
		{
			desc: "missing tenant id",
			envVars: map[string]string{
				EnvTenantID:    "",
				EnvAPIUserID:   "api_user_id",
				EnvAPIPassword: "api_password",
			},
			expected: "conohav3: some credentials information are missing: CONOHAV3_TENANT_ID",
		},
		{
			desc: "missing api user id",
			envVars: map[string]string{
				EnvTenantID:    "tenant_id",
				EnvAPIUserID:   "",
				EnvAPIPassword: "api_password",
			},
			expected: "conohav3: some credentials information are missing: CONOHAV3_API_USER_ID",
		},
		{
			desc: "missing api password",
			envVars: map[string]string{
				EnvTenantID:    "tenant_id",
				EnvAPIUserID:   "api_user_id",
				EnvAPIPassword: "",
			},
			expected: "conohav3: some credentials information are missing: CONOHAV3_API_PASSWORD",
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
		expected string
		tenant   string
		userid   string
		password string
	}{
		{
			desc:     "complete credentials, but login failed",
			expected: `conohav3: failed to log in: unexpected status code: [status code: 400] body: {"code": 400, "error": "user does not exist"}`,
			tenant:   "tenant_id",
			userid:   "api_user_id",
			password: "api_password",
		},
		{
			desc:     "missing credentials",
			expected: "conohav3: some credentials information are missing",
		},
		{
			desc:     "missing tenant id",
			expected: "conohav3: some credentials information are missing",
			userid:   "api_user_id",
			password: "api_password",
		},
		{
			desc:     "missing api user id",
			expected: "conohav3: some credentials information are missing",
			tenant:   "tenant_id",
			password: "api_password",
		},
		{
			desc:     "missing api password",
			expected: "conohav3: some credentials information are missing",
			tenant:   "tenant_id",
			userid:   "api_user_id",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.TenantID = test.tenant
			config.UserID = test.userid
			config.Password = test.password

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
