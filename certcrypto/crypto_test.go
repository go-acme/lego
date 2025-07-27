package certcrypto

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDomain1 = "lego.example"
	testDomain2 = "a.lego.example"
	testDomain3 = "b.lego.example"
	testDomain4 = "c.lego.example"
)

func TestGeneratePrivateKey(t *testing.T) {
	key, err := GeneratePrivateKey(RSA2048)
	require.NoError(t, err, "Error generating private key")

	assert.NotNil(t, key)
}

func TestGenerateCSR(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err, "Error generating private key")

	type expected struct {
		len   int
		error bool
	}

	testCases := []struct {
		desc       string
		privateKey crypto.PrivateKey
		opts       CSROptions
		expected   expected
	}{
		{
			desc:       "without SAN (nil)",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     testDomain1,
				MustStaple: true,
			},
			expected: expected{len: 382},
		},
		{
			desc:       "without SAN (empty)",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     testDomain1,
				SAN:        []string{},
				MustStaple: true,
			},
			expected: expected{len: 382},
		},
		{
			desc:       "with SAN",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     testDomain1,
				SAN:        []string{testDomain2, testDomain3, testDomain4},
				MustStaple: true,
			},
			expected: expected{len: 442},
		},
		{
			desc:       "no domain",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     "",
				MustStaple: true,
			},
			expected: expected{len: 359},
		},
		{
			desc:       "no domain with SAN",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     "",
				SAN:        []string{testDomain2, testDomain3, testDomain4},
				MustStaple: true,
			},
			expected: expected{len: 419},
		},
		{
			desc:       "private key nil",
			privateKey: nil,
			opts: CSROptions{
				Domain:     testDomain1,
				MustStaple: true,
			},
			expected: expected{error: true},
		},
		{
			desc:       "with email addresses",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:         "example.com",
				SAN:            []string{"example.org"},
				EmailAddresses: []string{"foo@example.com", "bar@example.com"},
			},
			expected: expected{len: 421},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			csr, err := CreateCSR(test.privateKey, test.opts)

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
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err, "Error generating private key")

	data := PEMEncode(key)
	require.NotNil(t, data)

	p, rest := pem.Decode(data)

	assert.Equal(t, "RSA PRIVATE KEY", p.Type)
	assert.Empty(t, rest)
	assert.Empty(t, p.Headers)
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

func TestParsePEMPrivateKey(t *testing.T) {
	privateKey, err := GeneratePrivateKey(RSA2048)
	require.NoError(t, err, "Error generating private key")

	pemPrivateKey := PEMEncode(privateKey)

	// Decoding a key should work and create an identical RSA key to the original,
	// ignoring precomputed values.
	decoded, err := ParsePEMPrivateKey(pemPrivateKey)
	require.NoError(t, err)
	decodedRsaPrivateKey := decoded.(*rsa.PrivateKey)
	require.True(t, decodedRsaPrivateKey.Equal(privateKey))

	// Decoding a PEM block that doesn't contain a private key should error
	_, err = ParsePEMPrivateKey(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE"}))
	require.Errorf(t, err, "Expected to return an error for non-private key input")

	// Decoding a PEM block that doesn't actually contain a key should error
	_, err = ParsePEMPrivateKey(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY"}))
	require.Errorf(t, err, "Expected to return an error for empty input")

	// Decoding non-PEM input should return an error
	_, err = ParsePEMPrivateKey([]byte("This is not PEM"))
	require.Errorf(t, err, "Expected to return an error for non-PEM input")
}
