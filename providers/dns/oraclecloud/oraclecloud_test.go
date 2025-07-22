package oraclecloud

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/nrdcg/oci-go-sdk/common/v1065"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

// Used by Instance Principal authentication.
const (
	envMetadataBaseURL        = "OCI_METADATA_BASE_URL"
	envSDKAuthClientRegionURL = "OCI_SDK_AUTH_CLIENT_REGION_URL"
)

var envTest = tester.NewEnvTest(
	envPrivKey,
	EnvAuthType,
	envMetadataBaseURL,
	envSDKAuthClientRegionURL,
	EnvPrivKeyFile,
	EnvPrivKeyPass,
	EnvTenancyOCID,
	EnvUserOCID,
	EnvPubKeyFingerprint,
	EnvRegion,
	EnvCompartmentOCID).
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
			expected: "oraclecloud: can not create client, bad configuration: ",
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

func TestNewDNSProvider_instance_principal(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAuthType:        "instance_principal",
				EnvCompartmentOCID: "123",
			},
		},
		{
			desc: "missing CompartmentID",
			envVars: map[string]string{
				EnvAuthType: "instance_principal",
			},
			expected: "oraclecloud: some credentials information are missing: OCI_COMPARTMENT_OCID",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer func() {
				envTest.RestoreEnv()
			}()

			envTest.ClearEnv()

			serverURL := servermock.NewBuilder(
				func(server *httptest.Server) (string, error) {
					return server.URL, nil
				}).
				Route("GET /instance/region", servermock.RawStringResponse("oc1")).
				// To generate fake certificates:
				//     go run `go env GOROOT`/src/crypto/tls/generate_cert.go --host example.org --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
				Route("GET /identity/cert.pem", servermock.ResponseFromFixture("cert.pem")).
				Route("GET /identity/key.pem", servermock.ResponseFromFixture("key.pem")).
				Route("GET /identity/intermediate.pem", servermock.ResponseFromFixture("cert.pem")).
				// https://github.com/oracle/oci-go-sdk/blob/413a2f277f95c5eb76e26a0e0833c396a518bf50/common/auth/jwt_test.go#L12
				Route("POST /v1/x509", servermock.RawStringResponse(`{"token":"eyJhbGciOiJSUzI1NiIsImtpZCI6ImFzdyIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJvcGMub3JhY2xlLmNvbSIsImV4cCI6MTUxMTgzODc5MywiaWF0IjoxNTExODE3MTkzLCJpc3MiOiJhdXRoU2VydmljZS5vcmFjbGUuY29tIiwib3BjLWNlcnR0eXBlIjoiaW5zdGFuY2UiLCJvcGMtY29tcGFydG1lbnQiOiJvY2lkMS5jb21wYXJ0bWVudC5vYzEuLmJsdWhibHVoYmx1aCIsIm9wYy1pbnN0YW5jZSI6Im9jaWQxLmluc3RhbmNlLm9jMS5waHguYmx1aGJsdWhibHVoIiwib3BjLXRlbmFudCI6Im9jaWR2MTp0ZW5hbmN5Om9jMTpwaHg6MTIzNDU2Nzg5MDpibHVoYmx1aGJsdWgiLCJwdHlwZSI6Imluc3RhbmNlIiwic3ViIjoib2NpZDEuaW5zdGFuY2Uub2MxLnBoeC5ibHVoYmx1aGJsdWgiLCJ0ZW5hbnQiOiJvY2lkdjE6dGVuYW5jeTpvYzE6cGh4OjEyMzQ1Njc4OTA6Ymx1aGJsdWhibHVoIiwidHR5cGUiOiJ4NTA5In0.zen7q2yJSpMjzH4ym_H7VEwZA0-vTT4Wcild-HRfLxX6A1ej4tlpACa7A24j5JoZYI4mHooZVJ8e7ZezFenK0zZx5j8RbIjsqJKwroYXExOiBXLCUwMWOLXIndEsUzzGLqnPfKHXd80vrhMLmtkVTCJqBMzvPUSYkH_ciWgmjP9m0YETdQ9ifghkADhZGt9IlnOswg0s3Bx9ASwxFZEtom0BmU9GwEuITTTZfKvndk785BlNeZMOjhovaD97-LYpv5B_PiWEz8zialK5zxjijLCw06zyA8CQRQqmVCagNUPilfz_BcPyImzvFDuzQcPyDkTcsB7weX35tafHmA_Ul"}`)).
				Build(t)

			envVars := map[string]string{
				envMetadataBaseURL:        serverURL,
				envSDKAuthClientRegionURL: serverURL,
			}

			for k, v := range test.envVars {
				envVars[k] = v
			}

			envTest.Apply(envVars)

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
		desc                  string
		compartmentID         string
		configurationProvider common.ConfigurationProvider
		expected              string
	}{
		{
			desc:                  "configuration provider error",
			configurationProvider: mockConfigurationProvider("wrong-secret"),
			compartmentID:         "123",
			expected:              "oraclecloud: can not create client, bad configuration: x509: decryption password incorrect",
		},
		{
			desc:          "OCIConfigProvider is missing",
			compartmentID: "123",
			expected:      "oraclecloud: OCIConfigProvider is missing",
		},
		{
			desc:                  "missing CompartmentID",
			configurationProvider: mockConfigurationProvider("secret"),
			expected:              "oraclecloud: CompartmentID is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.CompartmentID = test.compartmentID
			config.OCIConfigProvider = test.configurationProvider

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockConfigurationProvider(keyPassphrase string) *environmentConfigurationProvider {
	envTest.Apply(map[string]string{
		envPrivKey: mustGeneratePrivateKey("secret"),
	})

	return &environmentConfigurationProvider{
		values: map[string]string{
			EnvCompartmentOCID:   "test",
			EnvPrivKeyPass:       "test",
			EnvTenancyOCID:       "test",
			EnvUserOCID:          "test",
			EnvPubKeyFingerprint: "test",
			EnvRegion:            "test",
		},
		privateKeyPassphrase: keyPassphrase,
	}
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

	file, err := os.CreateTemp("", "lego_oci_*.pem")
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
	key, err := rsa.GenerateKey(rand.Reader, 1024)
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
