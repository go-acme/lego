package e2e

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/e2e/internal"
	"github.com/go-acme/lego/v5/e2e/loader"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeHTTP_Run(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"-d", testDomain1,
		"--http",
		"--http.port", ":5002",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeHTTP_Client_Obtain(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{testDomain1},
		Bundle:  true,
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

func TestChallengeHTTP_Client_Obtain_profile(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{testDomain1},
		Bundle:  true,
		Profile: "shortlived",
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

func TestChallengeHTTP_Client_Obtain_emails_csr(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains:        []string{testDomain1},
		Bundle:         true,
		EmailAddresses: []string{testEmail1},
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

func TestChallengeHTTP_Client_Obtain_notBefore_notAfter(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	now := time.Now().UTC()

	request := certificate.ObtainRequest{
		Domains:   []string{testDomain1},
		NotBefore: now.Add(1 * time.Hour),
		NotAfter:  now.Add(2 * time.Hour),
		Bundle:    true,
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

	cert, err := certcrypto.ParsePEMCertificate(resource.Certificate)
	require.NoError(t, err)
	assert.WithinDuration(t, now.Add(1*time.Hour), cert.NotBefore, 1*time.Second)
	assert.WithinDuration(t, now.Add(2*time.Hour), cert.NotAfter, 1*time.Second)
}

func TestChallengeHTTP_Client_Registration_QueryRegistration(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	resource, err := client.Registration.QueryRegistration(ctx)
	require.NoError(t, err)

	require.NotNil(t, resource)

	assert.Equal(t, "valid", resource.Status)
	assert.Regexp(t, `https://localhost:14000/list-orderz/[\w\d]+`, resource.Orders)
	assert.Regexp(t, `https://localhost:14000/my-account/[\w\d]+`, resource.Location)
}
