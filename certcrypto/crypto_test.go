package certcrypto

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
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

func TestGenerateCSR(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err, "Error generating private key")

	type expected struct {
		len   int
		error bool
	}

	testCases := []struct {
		desc       string
		privateKey crypto.Signer
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
			desc:       "with IP as SAN",
			privateKey: privateKey,
			opts: CSROptions{
				Domain:     testDomain1,
				SAN:        []string{"127.0.0.1"},
				MustStaple: true,
			},
			expected: expected{len: 399},
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

func TestGeneratePrivateKey_error(t *testing.T) {
	_, err := GeneratePrivateKey("foo")

	require.EqualError(t, err, "invalid KeyType: foo")
}

func TestGetCertificateMainDomain(t *testing.T) {
	testCases := []struct {
		desc     string
		cert     *x509.Certificate
		expected string
	}{
		{
			desc: "CommonName",
			cert: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
				DNSNames: []string{
					"example.org",
					"example.net",
				},
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
			},
			expected: "example.com",
		},
		{
			desc: "DNSNames",
			cert: &x509.Certificate{
				DNSNames: []string{
					"example.org",
					"example.net",
				},
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
			},
			expected: "example.org",
		},
		{
			desc: "IPAddresses",
			cert: &x509.Certificate{
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
					net.ParseIP("10.0.0.1"),
				},
			},
			expected: "127.0.0.1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domain, err := GetCertificateMainDomain(test.cert)
			require.NoError(t, err)

			assert.Equal(t, test.expected, domain)
		})
	}
}

func TestGetCertificateMainDomain_error(t *testing.T) {
	cert := &x509.Certificate{}

	_, err := GetCertificateMainDomain(cert)
	require.EqualError(t, err, "missing domain")
}

func TestExtractDomains(t *testing.T) {
	testCases := []struct {
		desc     string
		cert     *x509.Certificate
		expected []string
	}{
		{
			desc: "all",
			cert: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
				DNSNames: []string{
					"example.org",
					"example.net",
				},
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
			},
			expected: []string{
				"example.com",
				"example.org",
				"example.net",
				"127.0.0.1",
			},
		},
		{
			desc: "CommonName",
			cert: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
			},
			expected: []string{
				"example.com",
			},
		},
		{
			desc: "DNSNames",
			cert: &x509.Certificate{
				DNSNames: []string{
					"example.org",
					"example.net",
				},
			},
			expected: []string{
				"example.org",
				"example.net",
			},
		},
		{
			desc: "IPAddresses",
			cert: &x509.Certificate{
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
					net.ParseIP("10.0.0.1"),
				},
			},
			expected: []string{
				"127.0.0.1",
				"10.0.0.1",
			},
		},
		{
			desc: "empty",
			cert: &x509.Certificate{},
		},
		{
			desc: "same CommonName and DNSNames",
			cert: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
				DNSNames: []string{
					"example.com",
					"example.org",
					"example.net",
				},
			},
			expected: []string{
				"example.com",
				"example.org",
				"example.net",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domains := ExtractDomains(test.cert)

			assert.Equal(t, test.expected, domains)
		})
	}
}

func TestGetCSRMainDomain(t *testing.T) {
	testCases := []struct {
		desc     string
		cert     *x509.CertificateRequest
		expected string
	}{
		{
			desc: "CommonName",
			cert: &x509.CertificateRequest{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
				DNSNames: []string{
					"example.org",
					"example.net",
				},
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
			},
			expected: "example.com",
		},
		{
			desc: "DNSNames",
			cert: &x509.CertificateRequest{
				DNSNames: []string{
					"example.org",
					"example.net",
				},
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
			},
			expected: "example.org",
		},
		{
			desc: "IPAddresses",
			cert: &x509.CertificateRequest{
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
					net.ParseIP("10.0.0.1"),
				},
			},
			expected: "127.0.0.1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domain, err := GetCSRMainDomain(test.cert)
			require.NoError(t, err)

			assert.Equal(t, test.expected, domain)
		})
	}
}

func TestGetCSRMainDomain_error(t *testing.T) {
	cert := &x509.CertificateRequest{}

	_, err := GetCSRMainDomain(cert)
	require.EqualError(t, err, "missing domain")
}

func TestExtractDomainsCSR(t *testing.T) {
	testCases := []struct {
		desc     string
		cert     *x509.CertificateRequest
		expected []string
	}{
		{
			desc: "all",
			cert: &x509.CertificateRequest{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
				DNSNames: []string{
					"example.org",
					"example.net",
				},
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
			},
			expected: []string{
				"example.com",
				"example.org",
				"example.net",
				"127.0.0.1",
			},
		},
		{
			desc: "CommonName",
			cert: &x509.CertificateRequest{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
			},
			expected: []string{
				"example.com",
			},
		},
		{
			desc: "DNSNames",
			cert: &x509.CertificateRequest{
				DNSNames: []string{
					"example.org",
					"example.net",
				},
			},
			expected: []string{
				"example.org",
				"example.net",
			},
		},
		{
			desc: "IPAddresses",
			cert: &x509.CertificateRequest{
				IPAddresses: []net.IP{
					net.ParseIP("127.0.0.1"),
					net.ParseIP("10.0.0.1"),
				},
			},
			expected: []string{
				"127.0.0.1",
				"10.0.0.1",
			},
		},
		{
			desc: "empty",
			cert: &x509.CertificateRequest{},
		},
		{
			desc: "same CommonName and DNSNames",
			cert: &x509.CertificateRequest{
				Subject: pkix.Name{
					CommonName: "example.com",
				},
				DNSNames: []string{
					"example.com",
					"example.org",
					"example.net",
				},
			},
			expected: []string{
				"example.com",
				"example.org",
				"example.net",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domains := ExtractDomainsCSR(test.cert)

			assert.Equal(t, test.expected, domains)
		})
	}
}
