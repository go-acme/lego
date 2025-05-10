package oraclecloud

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	envPrivKey,
	EnvPrivKeyFile,
	EnvPrivKeyPass,
	EnvTenancyOCID,
	EnvUserOCID,
	EnvPubKeyFingerprint,
	EnvRegion,
	EnvCompartmentOCID).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvPrivKeyFile,
		EnvPrivKeyPass,
		EnvTenancyOCID,
		EnvUserOCID,
		EnvPubKeyFingerprint,
		EnvRegion,
		EnvCompartmentOCID)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				envPrivKey:           mustGeneratePrivateKey("secret1"),
				EnvPrivKeyPass:       "secret1",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "123",
			},
		},
		{
			desc: "success file",
			envVars: map[string]string{
				EnvPrivKeyFile:       mustGeneratePrivateKeyFile("secret1"),
				EnvPrivKeyPass:       "secret1",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "123",
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
				envPrivKey:           mustGeneratePrivateKey("secret"),
				EnvPrivKeyPass:       "secret",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_COMPARTMENT_OCID",
		},
		{
			desc: "missing OCI_PRIVKEY",
			envVars: map[string]string{
				envPrivKey:           "",
				EnvPrivKeyPass:       "secret",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_PRIVKEY",
		},
		{
			desc: "missing OCI_PRIVKEY_PASS",
			envVars: map[string]string{
				envPrivKey:           mustGeneratePrivateKey("secret"),
				EnvPrivKeyPass:       "",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "123",
			},
			expected: "oraclecloud: private key is encrypted but no password was provided",
		},
		{
			desc: "missing OCI_TENANCY_OCID",
			envVars: map[string]string{
				envPrivKey:           mustGeneratePrivateKey("secret"),
				EnvPrivKeyPass:       "secret",
				EnvTenancyOCID:       "",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_TENANCY_OCID",
		},
		{
			desc: "missing OCI_USER_OCID",
			envVars: map[string]string{
				envPrivKey:           mustGeneratePrivateKey("secret"),
				EnvPrivKeyPass:       "secret",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_USER_OCID",
		},
		{
			desc: "missing OCI_PUBKEY_FINGERPRINT",
			envVars: map[string]string{
				envPrivKey:           mustGeneratePrivateKey("secret"),
				EnvPrivKeyPass:       "secret",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "",
				EnvRegion:            "us-phoenix-1",
				EnvCompartmentOCID:   "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_PUBKEY_FINGERPRINT",
		},
		{
			desc: "missing OCI_REGION",
			envVars: map[string]string{
				envPrivKey:           mustGeneratePrivateKey("secret"),
				EnvPrivKeyPass:       "secret",
				EnvTenancyOCID:       "ocid1.tenancy.oc1..secret",
				EnvUserOCID:          "ocid1.user.oc1..secret",
				EnvPubKeyFingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
				EnvRegion:            "",
				EnvCompartmentOCID:   "123",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_REGION",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer func() {
				privKeyFile := os.Getenv(EnvPrivKeyFile)
				if privKeyFile != "" {
					_ = os.Remove(privKeyFile)
				}
				envTest.RestoreEnv()
			}()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	envTest.ClearEnv()
	defer envTest.RestoreEnv()

	testCases := []struct {
		desc          string
		privateKey    *rsa.PrivateKey
		keyID         string
		compartmentID string
		region        string
		expected      string
	}{
		{
			desc:          "success",
			privateKey:    mustGenerateKey(),
			keyID:         "keyID",
			compartmentID: "123",
			region:        "us-phoenix-1",
		},
		{
			desc:     "missing privateKey",
			keyID:    "keyID",
			region:   "us-phoenix-1",
			expected: "oraclecloud: PrivateKey is missing",
		},
		{
			desc:       "missing keyID",
			privateKey: mustGenerateKey(),
			region:     "us-phoenix-1",
			expected:   "oraclecloud: KeyID is missing",
		},
		{
			desc:       "missing region",
			privateKey: mustGenerateKey(),
			keyID:      "keyID",
			expected:   "oraclecloud: Region is missing",
		},
		{
			desc:       "missing compartmentID",
			privateKey: mustGenerateKey(),
			keyID:      "keyID",
			region:     "us-phoenix-1",
			expected:   "oraclecloud: CompartmentID is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.PrivateKey = test.privateKey
			config.KeyID = test.keyID
			config.Region = test.region
			config.CompartmentID = test.compartmentID

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

	time.Sleep(2 * time.Second)
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

func mustGenerateKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}

func mustGeneratePrivateKey(pwd string) string {
	pemBlock, err := generatePrivateKey(pwd)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(pemBlock))
}

func mustGeneratePrivateKeyFile(pwd string) string {
	pemBlock, err := generatePrivateKey(pwd)
	if err != nil {
		panic(err)
	}

	// Create temporary file for the key
	tmpfile, err := os.CreateTemp("", "lego_test_")
	if err != nil {
		panic(err)
	}

	err = pem.Encode(tmpfile, pemBlock)
	if err != nil {
		panic(err)
	}

	err = tmpfile.Close()
	if err != nil {
		panic(err)
	}

	return tmpfile.Name()
}

func generatePrivateKey(pwd string) (*pem.Block, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader,
		"RSA PRIVATE KEY",
		x509.MarshalPKCS1PrivateKey(key),
		[]byte(pwd), x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}

	return encryptedBlock, nil
}
