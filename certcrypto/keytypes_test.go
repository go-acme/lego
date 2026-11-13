package certcrypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetKeyType_unsupported(t *testing.T) {
	mustGenerateKey := func(c crypto.Signer, err error) crypto.Signer {
		require.NoError(t, err)
		return c
	}

	testCases := []struct {
		keyType  string
		key      crypto.Signer
		expected string
	}{
		{
			keyType:  "P224",
			key:      mustGenerateKey(ecdsa.GenerateKey(elliptic.P224(), rand.Reader)),
			expected: "unsupported ECDSA curve: 224",
		},
		{
			keyType:  "RSA1024",
			key:      mustGenerateKey(rsa.GenerateKey(rand.Reader, 1024)),
			expected: "unsupported RSA key: 1024",
		},
	}

	for _, test := range testCases {
		t.Run(test.keyType, func(t *testing.T) {
			t.Parallel()

			_, err := GetKeyType(test.key)
			require.EqualError(t, err, test.expected)
		})
	}
}

func TestGetPrivateKeyType_unsupported(t *testing.T) {
	mustGenerateKey := func(c crypto.Signer, err error) crypto.Signer {
		require.NoError(t, err)
		return c
	}

	testCases := []struct {
		keyType  string
		key      crypto.Signer
		expected string
	}{
		{
			keyType:  "P224",
			key:      mustGenerateKey(ecdsa.GenerateKey(elliptic.P224(), rand.Reader)),
			expected: "unsupported ECDSA curve: 224",
		},
		{
			keyType:  "RSA1024",
			key:      mustGenerateKey(rsa.GenerateKey(rand.Reader, 1024)),
			expected: "unsupported RSA key: 1024",
		},
	}

	for _, test := range testCases {
		t.Run(test.keyType, func(t *testing.T) {
			t.Parallel()

			_, err := GetPrivateKeyType(test.key)
			require.EqualError(t, err, test.expected)
		})
	}
}

func TestGetCertificateKeyType(t *testing.T) {
	path := "./testdata/cert-p256.pem"

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	cert, err := ParsePEMCertificate(data)
	require.NoError(t, err)

	keyType, err := GetCertificateKeyType(cert)
	require.NoError(t, err)

	assert.Equal(t, EC256, keyType)
}

func TestGetCertificateKeyType_error(t *testing.T) {
	path := "./testdata/cert-p224.pem"

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	cert, err := ParsePEMCertificate(data)
	require.NoError(t, err)

	_, err = GetCertificateKeyType(cert)
	require.EqualError(t, err, "unsupported ECDSA curve: 224")
}

func TestToKeyType(t *testing.T) {
	for _, expected := range AllKeyTypes() {
		t.Run(expected.String(), func(t *testing.T) {
			t.Parallel()

			keyType, err := ToKeyType(strings.ToLower(expected.String()))
			require.NoError(t, err)

			assert.Equal(t, expected, keyType)
			assert.True(t, IsSupported(keyType))
			assert.Contains(t, AllKeyTypes(), keyType)
		})
	}
}
