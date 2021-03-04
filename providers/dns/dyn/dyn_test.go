package dyn

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvCustomerName,
	EnvUserName,
	EnvPassword).
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
				EnvCustomerName: "A",
				EnvUserName:     "B",
				EnvPassword:     "C",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvCustomerName: "",
				EnvUserName:     "",
				EnvPassword:     "",
			},
			expected: "dyn: some credentials information are missing: DYN_CUSTOMER_NAME,DYN_USER_NAME,DYN_PASSWORD",
		},
		{
			desc: "missing customer name",
			envVars: map[string]string{
				EnvCustomerName: "",
				EnvUserName:     "B",
				EnvPassword:     "C",
			},
			expected: "dyn: some credentials information are missing: DYN_CUSTOMER_NAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvCustomerName: "A",
				EnvUserName:     "",
				EnvPassword:     "C",
			},
			expected: "dyn: some credentials information are missing: DYN_USER_NAME",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvCustomerName: "A",
				EnvUserName:     "B",
				EnvPassword:     "",
			},
			expected: "dyn: some credentials information are missing: DYN_PASSWORD",
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
		desc         string
		customerName string
		password     string
		userName     string
		expected     string
	}{
		{
			desc:         "success",
			customerName: "A",
			password:     "B",
			userName:     "C",
		},
		{
			desc:     "missing credentials",
			expected: "dyn: credentials missing",
		},
		{
			desc:         "missing customer name",
			customerName: "",
			password:     "B",
			userName:     "C",
			expected:     "dyn: credentials missing",
		},
		{
			desc:         "missing password",
			customerName: "A",
			password:     "",
			userName:     "C",
			expected:     "dyn: credentials missing",
		},
		{
			desc:         "missing username",
			customerName: "A",
			password:     "B",
			userName:     "",
			expected:     "dyn: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.CustomerName = test.customerName
			config.Password = test.password
			config.UserName = test.userName

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
