package dyn

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest            bool
	envTestCustomerName string
	envTestUserName     string
	envTestPassword     string
	envTestDomain       string
)

func init() {
	envTestCustomerName = os.Getenv("DYN_CUSTOMER_NAME")
	envTestUserName = os.Getenv("DYN_USER_NAME")
	envTestPassword = os.Getenv("DYN_PASSWORD")
	envTestDomain = os.Getenv("DYN_DOMAIN")

	if len(envTestCustomerName) > 0 && len(envTestUserName) > 0 && len(envTestPassword) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("DYN_CUSTOMER_NAME", envTestCustomerName)
	os.Setenv("DYN_USER_NAME", envTestUserName)
	os.Setenv("DYN_PASSWORD", envTestPassword)
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
				"DYN_CUSTOMER_NAME": "A",
				"DYN_USER_NAME":     "B",
				"DYN_PASSWORD":      "C",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"DYN_CUSTOMER_NAME": "",
				"DYN_USER_NAME":     "",
				"DYN_PASSWORD":      "",
			},
			expected: "dyn: some credentials information are missing: DYN_CUSTOMER_NAME,DYN_USER_NAME,DYN_PASSWORD",
		},
		{
			desc: "missing customer name",
			envVars: map[string]string{
				"DYN_CUSTOMER_NAME": "",
				"DYN_USER_NAME":     "B",
				"DYN_PASSWORD":      "C",
			},
			expected: "dyn: some credentials information are missing: DYN_CUSTOMER_NAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				"DYN_CUSTOMER_NAME": "A",
				"DYN_USER_NAME":     "",
				"DYN_PASSWORD":      "C",
			},
			expected: "dyn: some credentials information are missing: DYN_USER_NAME",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				"DYN_CUSTOMER_NAME": "A",
				"DYN_USER_NAME":     "B",
				"DYN_PASSWORD":      "",
			},
			expected: "dyn: some credentials information are missing: DYN_PASSWORD",
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
			defer restoreEnv()
			os.Unsetenv("DNSPOD_API_KEY")

			config := NewDefaultConfig()
			config.CustomerName = test.customerName
			config.Password = test.password
			config.UserName = test.userName

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
