package certcrypto

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"regexp"
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
		opts       CSROptions
		expected   expected
	}{
		{
			desc:       "without SAN (nil)",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     "lego.acme",
				MustStaple: true,
			},
			expected: expected{len: 245},
		},
		{
			desc:       "without SAN (empty)",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     "lego.acme",
				SAN:        []string{},
				MustStaple: true,
			},
			expected: expected{len: 245},
		},
		{
			desc:       "with SAN",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     "lego.acme",
				SAN:        []string{"a.lego.acme", "b.lego.acme", "c.lego.acme"},
				MustStaple: true,
			},
			expected: expected{len: 296},
		},
		{
			desc:       "no domain",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     "",
				MustStaple: true,
			},
			expected: expected{len: 225},
		},
		{
			desc:       "no domain with SAN",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     "",
				SAN:        []string{"a.lego.acme", "b.lego.acme", "c.lego.acme"},
				MustStaple: true,
			},
			expected: expected{len: 276},
		},
		{
			desc:       "private key nil",
			privateKey: nil,
			opts: CSROptions{
				Domain:     "fizz.buzz",
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
			expected: expected{len: 287},
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
	buf := bytes.NewBufferString("TestingRSAIsSoMuchFun")

	reader := MockRandReader{b: buf}
	key, err := rsa.GenerateKey(reader, 32)
	require.NoError(t, err, "Error generating private key")

	data := PEMEncode(key)
	require.NotNil(t, data)

	exp := regexp.MustCompile(`^-----BEGIN RSA PRIVATE KEY-----\s+\S{60,}\s+-----END RSA PRIVATE KEY-----\s+`)
	assert.Regexp(t, exp, string(data))
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

	// Decoding a key should work and create an identical key to the original
	decoded, err := ParsePEMPrivateKey(pemPrivateKey)
	require.NoError(t, err)
	assert.Equal(t, decoded, privateKey)

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

type MockRandReader struct {
	b *bytes.Buffer
}

func (r MockRandReader) Read(p []byte) (int, error) {
	return r.b.Read(p)
}
