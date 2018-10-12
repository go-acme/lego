package netcup

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/acme"
)

var (
	liveTest              bool
	envTestCustomerNumber string
	envTestAPIKey         string
	envTestAPIPassword    string
	envTestDomain         string
)

func init() {
	envTestCustomerNumber = os.Getenv("NETCUP_CUSTOMER_NUMBER")
	envTestAPIKey = os.Getenv("NETCUP_API_KEY")
	envTestAPIPassword = os.Getenv("NETCUP_API_PASSWORD")
	envTestDomain = os.Getenv("NETCUP_DOMAIN")

	if len(envTestCustomerNumber) > 0 && len(envTestAPIKey) > 0 && len(envTestAPIPassword) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("NETCUP_CUSTOMER_NUMBER", envTestCustomerNumber)
	os.Setenv("NETCUP_API_KEY", envTestAPIKey)
	os.Setenv("NETCUP_API_PASSWORD", envTestAPIPassword)
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
				"NETCUP_CUSTOMER_NUMBER": "A",
				"NETCUP_API_KEY":         "B",
				"NETCUP_API_PASSWORD":    "C",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "",
				"NETCUP_API_KEY":         "",
				"NETCUP_API_PASSWORD":    "",
			},
			expected: "netcup: some credentials information are missing: NETCUP_CUSTOMER_NUMBER,NETCUP_API_KEY,NETCUP_API_PASSWORD",
		},
		{
			desc: "missing customer number",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "",
				"NETCUP_API_KEY":         "B",
				"NETCUP_API_PASSWORD":    "C",
			},
			expected: "netcup: some credentials information are missing: NETCUP_CUSTOMER_NUMBER",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "A",
				"NETCUP_API_KEY":         "",
				"NETCUP_API_PASSWORD":    "C",
			},
			expected: "netcup: some credentials information are missing: NETCUP_API_KEY",
		},
		{
			desc: "missing api password",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "A",
				"NETCUP_API_KEY":         "B",
				"NETCUP_API_PASSWORD":    "",
			},
			expected: "netcup: some credentials information are missing: NETCUP_API_PASSWORD",
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
		customer string
		key      string
		password string
		expected string
	}{
		{
			desc:     "success",
			customer: "A",
			key:      "B",
			password: "C",
		},
		{
			desc:     "missing credentials",
			expected: "netcup: netcup credentials missing",
		},
		{
			desc:     "missing customer",
			customer: "",
			key:      "B",
			password: "C",
			expected: "netcup: netcup credentials missing",
		},
		{
			desc:     "missing key",
			customer: "A",
			key:      "",
			password: "C",
			expected: "netcup: netcup credentials missing",
		},
		{
			desc:     "missing password",
			customer: "A",
			key:      "B",
			password: "",
			expected: "netcup: netcup credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("NETCUP_CUSTOMER_NUMBER")
			os.Unsetenv("NETCUP_API_KEY")
			os.Unsetenv("NETCUP_API_PASSWORD")

			config := NewDefaultConfig()
			config.Customer = test.customer
			config.Key = test.key
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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

func TestLivePresentAndCleanup(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	p, err := NewDNSProvider()
	require.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(envTestDomain, "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	require.NoError(t, err, "error finding DNSZone")

	zone = acme.UnFqdn(zone)

	testCases := []string{
		zone,
		"sub." + zone,
		"*." + zone,
		"*.sub." + zone,
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("domain(%s)", tc), func(t *testing.T) {
			err = p.Present(tc, "987d", "123d==")
			require.NoError(t, err)

			err = p.CleanUp(tc, "987d", "123d==")
			require.NoError(t, err, "Did not clean up! Please remove record yourself.")
		})
	}
}
