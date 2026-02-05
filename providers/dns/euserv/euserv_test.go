package euserv

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvEmail,
	EnvPassword,
	EnvOrderID,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvEmail:    "email@example.com",
				EnvPassword: "secret",
				EnvOrderID:  "orderA",
			},
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				EnvEmail:    "",
				EnvPassword: "secret",
				EnvOrderID:  "orderA",
			},
			expected: "euserv: some credentials information are missing: EUSERV_EMAIL",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvEmail:    "email@example.com",
				EnvPassword: "",
				EnvOrderID:  "orderA",
			},
			expected: "euserv: some credentials information are missing: EUSERV_PASSWORD",
		},
		{
			desc: "missing order ID",
			envVars: map[string]string{
				EnvEmail:    "email@example.com",
				EnvPassword: "secret",
				EnvOrderID:  "",
			},
			expected: "euserv: some credentials information are missing: EUSERV_ORDER_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "euserv: some credentials information are missing: EUSERV_EMAIL,EUSERV_PASSWORD,EUSERV_ORDER_ID",
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
		desc     string
		email    string
		password string
		orderID  string
		expected string
	}{
		{
			desc:     "success",
			email:    "email@example.com",
			password: "secret",
			orderID:  "orderA",
		},
		{
			desc:     "missing email",
			email:    "",
			password: "secret",
			orderID:  "orderA",
			expected: "euserv: credentials missing",
		},
		{
			desc:     "missing password",
			email:    "email@example.com",
			password: "",
			orderID:  "orderA",
			expected: "euserv: credentials missing",
		},
		{
			desc:     "missing order ID",
			email:    "email@example.com",
			password: "secret",
			orderID:  "",
			expected: "euserv: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "euserv: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Email = test.email
			config.Password = test.password
			config.OrderID = test.orderID

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
