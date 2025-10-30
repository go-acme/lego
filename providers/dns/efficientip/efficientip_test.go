package efficientip

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvPassword,
	EnvHostname,
	EnvDNSName,
	EnvViewName,
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
				EnvUsername: "user",
				EnvPassword: "secret",
				EnvHostname: "example.com",
				EnvDNSName:  "dns.smart",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "secret",
				EnvHostname: "example.com",
				EnvDNSName:  "dns.smart",
			},
			expected: "efficientip: some credentials information are missing: EFFICIENTIP_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "",
				EnvHostname: "example.com",
				EnvDNSName:  "dns.smart",
			},
			expected: "efficientip: some credentials information are missing: EFFICIENTIP_PASSWORD",
		},
		{
			desc: "missing hostname",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
				EnvHostname: "",
				EnvDNSName:  "dns.smart",
			},
			expected: "efficientip: some credentials information are missing: EFFICIENTIP_HOSTNAME",
		},
		{
			desc: "missing DNS name",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
				EnvHostname: "example.com",
				EnvDNSName:  "",
			},
			expected: "efficientip: some credentials information are missing: EFFICIENTIP_DNS_NAME",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "efficientip: some credentials information are missing: EFFICIENTIP_USERNAME,EFFICIENTIP_PASSWORD,EFFICIENTIP_HOSTNAME,EFFICIENTIP_DNS_NAME",
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
		username string
		password string
		hostname string
		dnsName  string
		expected string
	}{
		{
			desc:     "success",
			username: "user",
			password: "secret",
			hostname: "example.com",
			dnsName:  "dns.smart",
		},
		{
			desc:     "missing username",
			password: "secret",
			hostname: "example.com",
			dnsName:  "dns.smart",
			expected: "efficientip: missing username",
		},
		{
			desc:     "missing password",
			username: "user",
			hostname: "example.com",
			dnsName:  "dns.smart",
			expected: "efficientip: missing password",
		},
		{
			desc:     "missing hostname",
			username: "user",
			password: "secret",
			dnsName:  "dns.smart",
			expected: "efficientip: missing hostname",
		},
		{
			desc:     "missing dnsName",
			username: "user",
			password: "secret",
			hostname: "example.com",
			expected: "efficientip: missing dnsname",
		},
		{
			desc:     "missing all",
			expected: "efficientip: missing username",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()

			config.Username = test.username
			config.Password = test.password
			config.Hostname = test.hostname
			config.DNSName = test.dnsName

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
