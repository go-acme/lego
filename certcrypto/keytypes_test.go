package certcrypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPrivateKeyType_error(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	require.NoError(t, err)

	_, err = GetPrivateKeyType(key)
	require.EqualError(t, err, "unsupported ECDSA curve: 224")
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
