package liquidweb

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/liquidweb/liquidweb-go/network"
	"github.com/stretchr/testify/require"
)

const envDomain = EnvPrefix + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvPrefix+EnvURL,
	EnvPrefix+EnvUsername,
	EnvPrefix+EnvPassword,
	EnvPrefix+EnvZone).
	WithDomain(envDomain)

func setupTest(t *testing.T, initRecs ...network.DNSRecord) *DNSProvider {
	t.Helper()

	envTest.Apply(map[string]string{
		EnvPrefix + EnvUsername: "blars",
		EnvPrefix + EnvPassword: "tacoman",
		EnvPrefix + EnvURL:      mockApiServer(t, initRecs...),
		EnvPrefix + EnvZone:     "tacoman.com", // this needs to be removed from test?
	})

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	return provider
}

func TestNewDNSProvider(t *testing.T) {
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
	provider := setupTest(t)

	err := provider.Present("tacoman.com", "", "")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := setupTest(t, network.DNSRecord{
		Name:  "tacoman.com.",
		RData: "123",
		ID:    1234567,
	})

	provider.recordIDs["123"] = 1234567

	err := provider.CleanUp("tacoman.com.", "123", "")
	require.NoError(t, err, "fail to remove TXT record")
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
