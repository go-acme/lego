package hyperone

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvPassportLocation, EnvAPIUrl, EnvLocationID).
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
				EnvPassportLocation: "./internal/fixtures/validPassport.json",
				EnvAPIUrl:           "",
				EnvLocationID:       "",
			},
		},
		{
			desc: "invalid passport",
			envVars: map[string]string{
				EnvPassportLocation: "./internal/fixtures/invalidPassport.json",
				EnvAPIUrl:           "",
				EnvLocationID:       "",
			},
			expected: "hyperone: passport file validation failed: private key is missing",
		},
		{
			desc: "non existing passport",
			envVars: map[string]string{
				EnvPassportLocation: "./internal/fixtures/non-existing.json",
				EnvAPIUrl:           "",
				EnvLocationID:       "",
			},
			expected: "hyperone: failed to open passport file: open ./internal/fixtures/non-existing.json:",
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
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc             string
		passportLocation string
		apiEndpoint      string
		locationID       string
		expected         string
	}{
		{
			desc:             "success",
			passportLocation: "./internal/fixtures/validPassport.json",
			apiEndpoint:      "",
			locationID:       "",
		},
		{
			desc:             "invalid passport",
			passportLocation: "./internal/fixtures/invalidPassport.json",
			apiEndpoint:      "",
			locationID:       "",
			expected:         "hyperone: passport file validation failed: private key is missing",
		},
		{
			desc:             "non existing passport",
			passportLocation: "./internal/fixtures/non-existing.json",
			apiEndpoint:      "",
			locationID:       "",
			expected:         "hyperone: failed to open passport file: open ./internal/fixtures/non-existing.json:",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.PassportLocation = test.passportLocation
			config.APIEndpoint = test.apiEndpoint
			config.LocationID = test.locationID

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
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
