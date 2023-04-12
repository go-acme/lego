package nicru

import (
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
	"testing"
)

const defaultDomainName = "example.com"
const envDomain = envNamespace + "DOMAIN"

const (
	fakeServiceId   = "2519234972459cdfa23423adf143324f"
	fakeSecret      = "oo5ahrie0aiPho3Vee4siupoPhahdahCh1thiesohru"
	fakeServiceName = "DS1234567890"
	fakeUsername    = "1234567/NIC-D"
	fakePassword    = "einge8Goo2eBaiXievuj"
)

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword, EnvServiceId, EnvSecret, EnvServiceName).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvServiceId:   fakeServiceId,
				EnvSecret:      fakeSecret,
				EnvServiceName: fakeServiceName,
				EnvUsername:    fakeUsername,
				EnvPassword:    fakePassword,
			},
		},
		{
			desc: "missing serviceId",
			envVars: map[string]string{
				EnvSecret:      fakeSecret,
				EnvServiceName: fakeServiceName,
				EnvUsername:    fakeUsername,
				EnvPassword:    fakePassword,
			},
			expected: "nicru: some credentials information are missing: NIC_RU_SERVICE_ID",
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				EnvServiceId:   fakeServiceId,
				EnvServiceName: fakeServiceName,
				EnvUsername:    fakeUsername,
				EnvPassword:    fakePassword,
			},
			expected: "nicru: some credentials information are missing: NIC_RU_SECRET",
		},
		{
			desc: "missing service name",
			envVars: map[string]string{
				EnvServiceId: fakeServiceId,
				EnvSecret:    fakeSecret,
				EnvUsername:  fakeUsername,
				EnvPassword:  fakePassword,
			},
			expected: "nicru: some credentials information are missing: NIC_RU_SERVICE_NAME",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvServiceId:   fakeServiceId,
				EnvSecret:      fakeSecret,
				EnvServiceName: fakeServiceName,
				EnvPassword:    fakePassword,
			},
			expected: "nicru: some credentials information are missing: NIC_RU_USER",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvServiceId:   fakeServiceId,
				EnvSecret:      fakeSecret,
				EnvServiceName: fakeServiceName,
				EnvUsername:    fakeUsername,
			},
			expected: "nicru: some credentials information are missing: NIC_RU_PASSWORD",
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
				ServiceId:          fakeServiceId,
				Secret:             fakeSecret,
				ServiceName:        fakeServiceName,
				Username:           fakeUsername,
				Password:           fakePassword,
				TTL:                defaultTTL,
				PropagationTimeout: defaultPropagationTimeout,
				PollingInterval:    defaultPollingInterval,
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "nicru: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing service name",
			config: &Config{
				Username:           fakeUsername,
				Password:           fakePassword,
				TTL:                defaultTTL,
				PropagationTimeout: defaultPropagationTimeout,
				PollingInterval:    defaultPollingInterval,
			},
			expected: "nicru: unable to build RU CENTER client: service name is missing in credentials information",
		},
		{
			desc: "missing username",
			config: &Config{
				ServiceName:        fakeServiceName,
				ServiceId:          fakeServiceId,
				Password:           fakePassword,
				TTL:                defaultTTL,
				PropagationTimeout: defaultPropagationTimeout,
				PollingInterval:    defaultPollingInterval,
			},
			expected: "nicru: unable to build RU CENTER client: username is missing in credentials information",
		},
		{
			desc: "missing password",
			config: &Config{
				ServiceName:        fakeServiceName,
				ServiceId:          fakeServiceId,
				Secret:             fakeSecret,
				Username:           fakeUsername,
				TTL:                defaultTTL,
				PropagationTimeout: defaultPropagationTimeout,
				PollingInterval:    defaultPollingInterval,
			},
			expected: "nicru: unable to build RU CENTER client: password is missing in credentials information",
		},
		{
			desc: "missing secret",
			config: &Config{
				ServiceId:          fakeServiceId,
				ServiceName:        fakeServiceName,
				Username:           fakeUsername,
				Password:           fakePassword,
				PropagationTimeout: defaultPropagationTimeout,
				PollingInterval:    defaultPollingInterval,
			},
			expected: "nicru: unable to build RU CENTER client: secret is missing in credentials information",
		},
		{
			desc: "missing serviceId",
			config: &Config{
				ServiceName: fakeServiceName,
				Secret:      fakeSecret,
				Username:    fakeUsername,
				Password:    fakePassword,
				Domain:      defaultDomainName,
			},
			expected: "nicru: unable to build RU CENTER client: serviceId is missing in credentials information",
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
