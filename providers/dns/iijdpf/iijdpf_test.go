package iijdpf

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "TESTDOMAIN"

var envTest = tester.NewEnvTest(EnvAPIToken, EnvServiceCode).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIToken:    "A",
				EnvServiceCode: "dpmXXXXXX",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIToken: "A",
			},
			expected: "iijdpf: some credentials information are missing: IIJ_DPF_DPM_SERVICE_CODE",
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvServiceCode: "dpmXXXXXX",
			},
			expected: "iijdpf: some credentials information are missing: IIJ_DPF_API_TOKEN",
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
		desc        string
		token       string
		servicecode string
		expected    string
	}{
		{
			desc:        "success",
			token:       "A",
			servicecode: "dpm00000",
		},
		{
			desc:        "missing credentials",
			servicecode: "dpm00000",
			expected:    "iijdpf: API token missing",
		},
		{
			desc:     "missing credentials",
			token:    "A",
			expected: "iijdpf: Servicecode missing",
		},
		{
			desc:     "missing credentials",
			expected: "iijdpf: API token missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Token = test.token
			config.ServiceCode = test.servicecode

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
