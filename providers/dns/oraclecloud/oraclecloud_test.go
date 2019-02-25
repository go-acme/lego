package oraclecloud

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/platform/tester"
)

var envTest = tester.NewEnvTest(
	"OCI_PRIVKEY_BASE64",
	"OCI_PRIVKEY_PASS",
	"OCI_TENANCY_OCID",
	"OCI_USER_OCID",
	"OCI_PUBKEY_FINGERPRINT",
	"OCI_REGION",
	"OCI_COMPARTMENT_OCID").
	WithDomain("ORACLECLOUD_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"OCI_PRIVKEY_BASE64":     "secret",
				"OCI_PRIVKEY_PASS":       "secret",
				"OCI_TENANCY_OCID":       "ocid1.tenancy.oc1..secret",
				"OCI_USER_OCID":          "ocid1.user.oc1..secret",
				"OCI_PUBKEY_FINGERPRINT": "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				"OCI_REGION":             "us-phoenix-1",
				"OCI_COMPARTMENT_OCID":   "123",
			},
		},
		{
			desc: "missing CompartmentID",
			envVars: map[string]string{
				"OCI_PRIVKEY_BASE64":     "secret",
				"OCI_PRIVKEY_PASS":       "secret",
				"OCI_TENANCY_OCID":       "ocid1.tenancy.oc1..secret",
				"OCI_USER_OCID":          "ocid1.user.oc1..secret",
				"OCI_PUBKEY_FINGERPRINT": "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				"OCI_REGION":             "us-phoenix-1",
				"OCI_COMPARTMENT_OCID":   "",
			},
			expected: "oraclecloud: can not read CompartmentID from environment variable OCI_COMPARTMENT_OCID",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

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
	// validate to golangci-lint
	config := &Config{}
	config.TTL = 60
	config = nil

	_, err := NewDNSProviderConfig(config)
	require.EqualError(t, err, "oraclecloud: the configuration of the DNS provider is nil")
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
