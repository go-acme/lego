package certcrypto

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePrivateKey(t *testing.T) {
	key, err := GeneratePrivateKey(RSA2048)
	require.NoError(t, err, "Error generating private key")

	assert.NotNil(t, key)
}

func TestGenerateCSR(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Error generating private key")

	type expected struct {
		len   int
		error bool
	}

	testCases := []struct {
		desc       string
		privateKey crypto.PrivateKey
		domain     string
		san        []string
		mustStaple bool
		expected   expected
	}{
		{
			desc:       "without SAN",
			privateKey: privateKey,
			domain:     "lego.acme",
			mustStaple: true,
			expected:   expected{len: 245},
		},
		{
			desc:       "without SAN",
			privateKey: privateKey,
			domain:     "lego.acme",
			san:        []string{},
			mustStaple: true,
			expected:   expected{len: 245},
		},
		{
			desc:       "with SAN",
			privateKey: privateKey,
			domain:     "lego.acme",
			san:        []string{"a.lego.acme", "b.lego.acme", "c.lego.acme"},
			mustStaple: true,
			expected:   expected{len: 296},
		},
		{
			desc:       "no domain",
			privateKey: privateKey,
			domain:     "",
			mustStaple: true,
			expected:   expected{len: 225},
		},
		{
			desc:       "no domain with SAN",
			privateKey: privateKey,
			domain:     "",
			san:        []string{"a.lego.acme", "b.lego.acme", "c.lego.acme"},
			mustStaple: true,
			expected:   expected{len: 276},
		},
		{
			desc:       "private key nil",
			privateKey: nil,
			domain:     "fizz.buzz",
			mustStaple: true,
			expected:   expected{error: true},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			csr, err := GenerateCSR(test.privateKey, test.domain, test.san, test.mustStaple)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err, "Error generating CSR")

				assert.NotEmpty(t, csr)
				assert.Len(t, csr, test.expected.len)
			}
		})
	}
}

func TestPEMEncode(t *testing.T) {
	buf := bytes.NewBufferString("TestingRSAIsSoMuchFun")

	reader := MockRandReader{b: buf}
	key, err := rsa.GenerateKey(reader, 32)
	require.NoError(t, err, "Error generating private key")

	data := PEMEncode(key)
	require.NotNil(t, data)
	assert.Len(t, data, 127)
}

func TestParsePEMCertificate(t *testing.T) {
	privateKey, err := GeneratePrivateKey(RSA2048)
	require.NoError(t, err, "Error generating private key")

	expiration := time.Now().Add(365).Round(time.Second)
	certBytes, err := generateDerCert(privateKey.(*rsa.PrivateKey), expiration, "test.com", nil)
	require.NoError(t, err, "Error generating cert")

	buf := bytes.NewBufferString("TestingRSAIsSoMuchFun")

	// Some random string should return an error.
	cert, err := ParsePEMCertificate(buf.Bytes())
	require.Errorf(t, err, "returned %v", cert)

	// A DER encoded certificate should return an error.
	_, err = ParsePEMCertificate(certBytes)
	require.Error(t, err, "Expected to return an error for DER certificates")

	// A PEM encoded certificate should work ok.
	pemCert := PEMEncode(DERCertificateBytes(certBytes))
	cert, err = ParsePEMCertificate(pemCert)
	require.NoError(t, err)

	assert.Equal(t, expiration.UTC(), cert.NotAfter)
}

type MockRandReader struct {
	b *bytes.Buffer
}

func (r MockRandReader) Read(p []byte) (int, error) {
	return r.b.Read(p)
}
