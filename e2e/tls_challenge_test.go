package e2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/tlsalpn01"
	"github.com/go-acme/lego/v5/e2e/internal"
	"github.com/go-acme/lego/v5/e2e/loader"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeTLS_Run_Domains(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"-d", testDomain1,
		"--tls",
		"--tls.address", ":5001",
	)
	require.NoError(t, err)
}

func TestChallengeTLS_Run_IP(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"-d", "127.0.0.1",
		"--tls",
		"--tls.address", ":5001",
	)
	require.NoError(t, err)
}

func TestChallengeTLS_Run_CSR(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	csrPath := createTestCSRFile(t, true)

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"--csr", csrPath,
		"--tls",
		"--tls.address", ":5001",
	)
	require.NoError(t, err)
}

func TestChallengeTLS_Run_CSR_PEM(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	csrPath := createTestCSRFile(t, false)

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"--csr", csrPath,
		"--tls",
		"--tls.address", ":5001",
	)
	require.NoError(t, err)
}

func TestChallengeTLS_Run_Revoke(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"-d", testDomain2,
		"-d", testDomain3,
		"--tls",
		"--tls.address", ":5001",
	)
	require.NoError(t, err)

	err = load.RunLego(t.Context(),
		"certificates", "revoke",
		"-m", testEmail1,
		"-s", caDirectory,
		"-c", testDomain2,
	)
	require.NoError(t, err)
}

func TestChallengeTLS_Run_Revoke_Non_ASCII(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"-d", testDomain4,
		"--tls",
		"--tls.address", ":5001",
	)
	require.NoError(t, err)

	err = load.RunLego(t.Context(),
		"certificates", "revoke",
		"-m", testEmail1,
		"-s", caDirectory,
		"-c", testDomain4,
	)
	require.NoError(t, err)
}

func TestChallengeTLS_Client_Obtain(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	// https://github.com/letsencrypt/pebble/issues/285
	privateKeyCSR, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	request := certificate.ObtainRequest{
		Domains:    []string{testDomain1},
		KeyType:    certcrypto.RSA2048,
		Bundle:     true,
		PrivateKey: privateKeyCSR,
	}
	resource, err := client.Certificate.Obtain(ctx, request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, testDomain1, resource.ID)
	assert.Equal(t, []string{testDomain1}, resource.Domains)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.Empty(t, resource.CSR)
}

func TestChallengeTLS_Client_ObtainForCSR(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	csr, err := x509.ParseCertificateRequest(createTestCSR(t))
	require.NoError(t, err)

	resource, err := client.Certificate.ObtainForCSR(ctx, certificate.ObtainForCSRRequest{
		CSR:    csr,
		Bundle: true,
	})
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, testDomain1, resource.ID)
	assert.Equal(t, []string{testDomain1, testDomain2}, resource.Domains)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.NotEmpty(t, resource.CSR)
}

func TestChallengeTLS_Client_ObtainForCSR_profile(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	csr, err := x509.ParseCertificateRequest(createTestCSR(t))
	require.NoError(t, err)

	resource, err := client.Certificate.ObtainForCSR(ctx, certificate.ObtainForCSRRequest{
		CSR:     csr,
		Bundle:  true,
		Profile: "shortlived",
	})
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, testDomain1, resource.ID)
	assert.Equal(t, []string{testDomain1, testDomain2}, resource.Domains)
	assert.Equal(t, "shortlived", resource.Profile)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.NotEmpty(t, resource.CSR)
}
