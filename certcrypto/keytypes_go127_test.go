//go:build go1.27

package certcrypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE(ldez) RSA4096 and RSA8192 are not tested because the key generation is very slow.
func TestGetPrivateKeyType(t *testing.T) {
	mustGenerateKey := func(c crypto.Signer, err error) crypto.Signer {
		require.NoError(t, err)
		return c
	}

	testCases := []struct {
		keyType KeyType
		key     crypto.Signer
	}{
		{
			keyType: EC256,
			key:     mustGenerateKey(ecdsa.GenerateKey(elliptic.P256(), rand.Reader)),
		},
		{
			keyType: EC384,
			key:     mustGenerateKey(ecdsa.GenerateKey(elliptic.P384(), rand.Reader)),
		},
		{
			keyType: RSA2048,
			key:     mustGenerateKey(rsa.GenerateKey(rand.Reader, 2048)),
		},
		{
			keyType: RSA3072,
			key:     mustGenerateKey(rsa.GenerateKey(rand.Reader, 3072)),
		},
		{
			keyType: MLDSA44,
			key:     mustGenerateKey(mldsa.GenerateKey(mldsa.MLDSA44())),
		},
		{
			keyType: MLDSA65,
			key:     mustGenerateKey(mldsa.GenerateKey(mldsa.MLDSA65())),
		},
		{
			keyType: MLDSA87,
			key:     mustGenerateKey(mldsa.GenerateKey(mldsa.MLDSA87())),
		},
	}

	for _, test := range testCases {
		t.Run(test.keyType.String(), func(t *testing.T) {
			t.Parallel()

			keyType, err := GetPrivateKeyType(test.key)
			require.NoError(t, err)

			assert.Equal(t, test.keyType, keyType)
		})
	}
}

func TestAllKeyTypes(t *testing.T) {
	assert.Len(t, AllKeyTypes(), 9)
}
