package liquidweb

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/liquidweb/liquidweb-go/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = EnvPrefix + "DOMAIN"

func TestNewDNSProvider(t *testing.T) {
	envTest := tester.NewEnvTest(
		EnvPrefix+EnvURL,
		EnvPrefix+EnvUsername,
		EnvPrefix+EnvPassword,
		EnvPrefix+EnvZone).
		WithDomain(envDomain)
	defer envTest.ClearEnv()

	for _, test := range testNewDNSProvider_testdata {
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
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	envTest := tester.NewEnvTest(
		EnvPrefix+EnvURL,
		EnvPrefix+EnvUsername,
		EnvPrefix+EnvPassword,
		EnvPrefix+EnvZone).
		WithDomain(envDomain)

	envTest.Apply(map[string]string{
		EnvPrefix + EnvUsername: "blars",
		EnvPrefix + EnvPassword: "tacoman",
		EnvPrefix + EnvURL:      mockApiServer(t),
		EnvPrefix + EnvZone:     "tacoman.com", // this needs to be removed from test?
	})

	defer envTest.ClearEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present("tacoman.com", "", "")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	envTest := tester.NewEnvTest(
		EnvPrefix+EnvURL,
		EnvPrefix+EnvUsername,
		EnvPrefix+EnvPassword,
		EnvPrefix+EnvZone).
		WithDomain(envDomain)

	envTest.Apply(map[string]string{
		EnvPrefix + EnvUsername: "blars",
		EnvPrefix + EnvPassword: "tacoman",
		EnvPrefix + EnvURL: mockApiServer(t, network.DNSRecord{
			Name:   "_acme-challenge.tacoman.com",
			RData:  "123d==",
			Type:   "TXT",
			TTL:    300,
			ID:     1234567,
			ZoneID: 42,
		}),
		EnvPrefix + EnvZone: "tacoman.com", // this needs to be removed from test?
	})

	defer envTest.ClearEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	provider.recordIDs["123d=="] = 1234567

	err = provider.CleanUp("tacoman.com.", "123d==", "")
	require.NoError(t, err, "fail to remove TXT record")
}

func TestLivePresent(t *testing.T) {
	envTest := tester.NewEnvTest(
		EnvPrefix+EnvURL,
		EnvPrefix+EnvUsername,
		EnvPrefix+EnvPassword,
		EnvPrefix+EnvZone).
		WithDomain(envDomain)
	defer envTest.ClearEnv()

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
	envTest := tester.NewEnvTest(
		EnvPrefix+EnvURL,
		EnvPrefix+EnvUsername,
		EnvPrefix+EnvPassword,
		EnvPrefix+EnvZone).
		WithDomain(envDomain)

	defer envTest.ClearEnv()

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

func TestIntegration(t *testing.T) {
	envTest := tester.NewEnvTest(
		"LWAPI_USERNAME",
		"LWAPI_PASSWORD",
		"LWAPI_URL")

	for testName, td := range testIntegration_testdata {
		t.Run(testName, func(t *testing.T) {
			td := td

			td.envVars["LWAPI_URL"] = mockApiServer(t, td.initRecs...)
			envTest.ClearEnv()
			envTest.Apply(td.envVars)

			provider, err := NewDNSProvider()
			require.NoError(t, err)

			if td.present {
				err = provider.Present(td.domain, td.token, td.keyauth)
				if td.expPresentErr == "" {
					assert.NoError(t, err)
				} else {
					assert.Equal(t, td.expPresentErr, err.Error())
				}
			}

			if td.cleanup {
				err = provider.CleanUp(td.domain, td.token, td.keyauth)
				if td.expCleanupErr == "" {
					assert.NoError(t, err)
				} else {
					assert.Equal(t, td.expCleanupErr, err.Error())
				}
			}
		})
	}
}
