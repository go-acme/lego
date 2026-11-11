//go:build !go1.27

package certcrypto

import (
	"crypto/ecdsa"
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
