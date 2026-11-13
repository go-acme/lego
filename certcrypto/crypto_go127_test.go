//go:build go1.27

package certcrypto

import (
	"crypto/ecdsa"
	"crypto/mldsa"
	"crypto/rsa"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE(ldez) RSA4096 and RSA8192 are not tested because the key generation is very slow.
func TestGeneratePrivateKey(t *testing.T) {
	testCases := []struct {
		keyType  KeyType
		certType any
	}{
		{
			keyType:  EC256,
			certType: &ecdsa.PrivateKey{},
		},
		{
			keyType:  EC384,
			certType: &ecdsa.PrivateKey{},
		},
		{
			keyType:  RSA2048,
			certType: &rsa.PrivateKey{},
		},
		{
			keyType:  RSA3072,
			certType: &rsa.PrivateKey{},
		},
		{
			keyType:  MLDSA44,
			certType: &mldsa.PrivateKey{},
		},
		{
			keyType:  MLDSA65,
			certType: &mldsa.PrivateKey{},
		},
		{
			keyType:  MLDSA87,
			certType: &mldsa.PrivateKey{},
		},
	}

	for _, test := range testCases {
		t.Run(test.keyType.String(), func(t *testing.T) {
			t.Parallel()

			key, err := GeneratePrivateKey(test.keyType)
			require.NoError(t, err, "Error generating private key")

			assert.IsType(t, test.certType, key)

			keyType, err := GetKeyType(key)
			require.NoError(t, err)

			assert.Equal(t, test.keyType, keyType)
		})
	}
}

// NOTE(ldez) RSA4096 and RSA8192 are not tested because the key generation is very slow.
func TestPEMEncode(t *testing.T) {
	for _, keyType := range []KeyType{
		EC256, EC384,
		RSA2048, RSA3072,
		MLDSA44, MLDSA65, MLDSA87,
	} {
		t.Run(keyType.String(), func(t *testing.T) {
			t.Parallel()

			key, err := GeneratePrivateKey(keyType)
			require.NoError(t, err, "Error generating private key")

			data := PEMEncode(key)
			require.NotNil(t, data)

			p, rest := pem.Decode(data)

			assert.Equal(t, "PRIVATE KEY", p.Type)
			assert.Empty(t, rest)
			assert.Empty(t, p.Headers)
		})
	}
}

// NOTE(ldez) RSA4096 and RSA8192 are not tested because the key generation is very slow.
func TestParsePEMPrivateKey(t *testing.T) {
	testCases := []struct {
		keyType  KeyType
		certType any
	}{
		{
			keyType:  EC256,
			certType: &ecdsa.PrivateKey{},
		},
		{
			keyType:  EC384,
			certType: &ecdsa.PrivateKey{},
		},
		{
			keyType:  RSA2048,
			certType: &rsa.PrivateKey{},
		},
		{
			keyType:  RSA3072,
			certType: &rsa.PrivateKey{},
		},
		{
			keyType:  MLDSA44,
			certType: &mldsa.PrivateKey{},
		},
		{
			keyType:  MLDSA65,
			certType: &mldsa.PrivateKey{},
		},
		{
			keyType:  MLDSA87,
			certType: &mldsa.PrivateKey{},
		},
	}

	for _, test := range testCases {
		t.Run(test.keyType.String(), func(t *testing.T) {
			t.Parallel()

			privateKey, err := GeneratePrivateKey(test.keyType)
			require.NoError(t, err, "Error generating private key")

			pemPrivateKey := PEMEncode(privateKey)

			// Decoding a key should work and create an identical RSA key to the original,
			// ignoring precomputed values.
			decoded, err := ParsePEMPrivateKey(pemPrivateKey)
			require.NoError(t, err)

			require.IsType(t, test.certType, decoded)

			// Decoding a PEM block that doesn't contain a private key should error
			_, err = ParsePEMPrivateKey(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE"}))
			require.Errorf(t, err, "Expected to return an error for non-private key input")

			// Decoding a PEM block that doesn't contain a private key should error
			_, err = ParsePEMPrivateKey(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE"}))
			require.Errorf(t, err, "Expected to return an error for non-private key input")

			// Decoding a PEM block that doesn't actually contain a key should error
			_, err = ParsePEMPrivateKey(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY"}))
			require.Errorf(t, err, "Expected to return an error for empty input")

			// Decoding non-PEM input should return an error
			_, err = ParsePEMPrivateKey([]byte("This is not PEM"))
			require.Errorf(t, err, "Expected to return an error for non-PEM input")
		})
	}
}
