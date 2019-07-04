package oraclecloud

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	ociPrivkey,
	ociPrivkey+"_FILE",
	ociPrivkeyPass,
	ociTenancyOCID,
	ociUserOCID,
	ociPubkeyFingerprint,
	ociRegion,
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
				ociPrivkey:             mustGeneratePrivateKey("secret1"),
				ociPrivkeyPass:         "secret1",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "123",
			},
		},
		{
			desc: "success file",
			envVars: map[string]string{
				ociPrivkey + "_FILE":   mustGeneratePrivateKeyFile("secret1"),
				ociPrivkeyPass:         "secret1",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "123",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "oraclecloud: some credentials information are missing: OCI_PRIVKEY,OCI_TENANCY_OCID,OCI_USER_OCID,OCI_PUBKEY_FINGERPRINT,OCI_REGION,OCI_COMPARTMENT_OCID",
		},
		{
			desc: "missing CompartmentID",
			envVars: map[string]string{
				ociPrivkey:             mustGeneratePrivateKey("secret"),
				ociPrivkeyPass:         "secret",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_COMPARTMENT_OCID",
		},
		{
			desc: "missing OCI_PRIVKEY",
			envVars: map[string]string{
				ociPrivkey:             "",
				ociPrivkeyPass:         "secret",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_PRIVKEY",
		},
		{
			desc: "missing OCI_PRIVKEY_PASS",
			envVars: map[string]string{
				ociPrivkey:             mustGeneratePrivateKey("secret"),
				ociPrivkeyPass:         "",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "123",
			},
			expected: "oraclecloud: can not create client, bad configuration: x509: decryption password incorrect",
		},
		{
			desc: "missing OCI_TENANCY_OCID",
			envVars: map[string]string{
				ociPrivkey:             mustGeneratePrivateKey("secret"),
				ociPrivkeyPass:         "secret",
				ociTenancyOCID:         "",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_TENANCY_OCID",
		},
		{
			desc: "missing OCI_USER_OCID",
			envVars: map[string]string{
				ociPrivkey:             mustGeneratePrivateKey("secret"),
				ociPrivkeyPass:         "secret",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_USER_OCID",
		},
		{
			desc: "missing OCI_PUBKEY_FINGERPRINT",
			envVars: map[string]string{
				ociPrivkey:             mustGeneratePrivateKey("secret"),
				ociPrivkeyPass:         "secret",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "",
				ociRegion:              "us-phoenix-1",
				"OCI_COMPARTMENT_OCID": "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_PUBKEY_FINGERPRINT",
		},
		{
			desc: "missing OCI_REGION",
			envVars: map[string]string{
				ociPrivkey:             mustGeneratePrivateKey("secret"),
				ociPrivkeyPass:         "secret",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "",
				"OCI_COMPARTMENT_OCID": "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_REGION",
		},
		{
			desc: "missing OCI_REGION",
			envVars: map[string]string{
				ociPrivkey:             mustGeneratePrivateKey("secret"),
				ociPrivkeyPass:         "secret",
				ociTenancyOCID:         "ocid1.tenancy.oc1..secret",
				ociUserOCID:            "ocid1.user.oc1..secret",
				ociPubkeyFingerprint:   "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				ociRegion:              "",
				"OCI_COMPARTMENT_OCID": "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_REGION",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer func() {
				privKeyFile := os.Getenv(ociPrivkey + "_FILE")
				if privKeyFile != "" {
					_ = os.Remove(privKeyFile)
				}
				envTest.RestoreEnv()
			}()
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
	testCases := []struct {
		desc                  string
		compartmentID         string
		configurationProvider common.ConfigurationProvider
		expected              string
	}{
		{
			desc:                  "invalid configuration",
			configurationProvider: &configProvider{},
			compartmentID:         "123",
			expected:              "oraclecloud: can not create client, bad configuration: x509: decryption password incorrect",
		},
		{
			desc:          "OCIConfigProvider is missing",
			compartmentID: "123",
			expected:      "oraclecloud: OCIConfigProvider is missing",
		},
		{
			desc:     "missing CompartmentID",
			expected: "oraclecloud: CompartmentID is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.CompartmentID = test.compartmentID
			config.OCIConfigProvider = test.configurationProvider

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

func mustGeneratePrivateKey(pwd string) string {
	block, err := generatePrivateKey(pwd)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(block))
}

func mustGeneratePrivateKeyFile(pwd string) string {
	block, err := generatePrivateKey(pwd)
	if err != nil {
		panic(err)
	}

	file, err := ioutil.TempFile("", "lego_oci_*.pem")
	if err != nil {
		panic(err)
	}

	err = pem.Encode(file, block)
	if err != nil {
		panic(err)
	}

	return file.Name()
}

func generatePrivateKey(pwd string) (*pem.Block, error) {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if pwd != "" {
		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(pwd), x509.PEMCipherAES256)
		if err != nil {
			return nil, err
		}
	}

	return block, nil
}
