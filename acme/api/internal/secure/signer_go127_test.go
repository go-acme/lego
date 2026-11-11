//go:build go1.27

package secure

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"
)

// NOTE(ldez) RSA4096 and RSA8192 are not tested because the key generation is very slow.
func TestSigner_SignContent(t *testing.T) {
	mustGenerateKey := func(c crypto.Signer, err error) crypto.Signer {
		require.NoError(t, err)
		return c
	}

	testCases := []struct {
		desc string
		key  crypto.Signer
	}{
		{
			desc: "RSA2048",
			key:  mustGenerateKey(rsa.GenerateKey(rand.Reader, 2048)),
		},
		{
			desc: "RSA3072",
			key:  mustGenerateKey(rsa.GenerateKey(rand.Reader, 3072)),
		},
		{
			desc: "EC256",
			key:  mustGenerateKey(ecdsa.GenerateKey(elliptic.P256(), rand.Reader)),
		},
		{
			desc: "EC384",
			key:  mustGenerateKey(ecdsa.GenerateKey(elliptic.P384(), rand.Reader)),
		},
		{
			desc: "MLDSA44",
			key:  mustGenerateKey(mldsa.GenerateKey(mldsa.MLDSA44())),
		},
		{
			desc: "MLDSA65",
			key:  mustGenerateKey(mldsa.GenerateKey(mldsa.MLDSA65())),
		},
		{
			desc: "MLDSA87",
			key:  mustGenerateKey(mldsa.GenerateKey(mldsa.MLDSA87())),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			signer := NewSigner(test.key, "https://example.com")

			content, err := signer.SignContent(&MockNonceSource{}, "https://foo.example", []byte("{}"))
			require.NoError(t, err)

			check(t, content)
		})
	}
}
