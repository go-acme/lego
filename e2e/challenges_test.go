package e2e

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/e2e/loader"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		HealthCheckURL: "https://localhost:14000/dir",
		Args:           []string{"-strict", "-config", "fixtures/pebble-config.json"},
		Env:            []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=./fixtures/certs/pebble.minica.pem",
		"LEGO_DEBUG_ACME_HTTP_CLIENT=1",
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(m))
}

func TestHelp(t *testing.T) {
	output, err := load.RunLego("-h")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}

func TestChallengeHTTP_Run(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--http",
		"--http.port", ":5002",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_Domains(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--tls",
		"--tls.port", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_IP(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-d", "127.0.0.1",
		"--tls",
		"--tls.port", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_CSR(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-csr", "./fixtures/csr.raw",
		"--tls",
		"--tls.port", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_CSR_PEM(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-csr", "./fixtures/csr.cert",
		"--tls",
		"--tls.port", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_Revoke(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-d", "lego.wtf",
		"-d", "acme.lego.wtf",
		"--tls",
		"--tls.port", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}

	output, err = load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-d", "lego.wtf",
		"--tls",
		"--tls.port", ":5001",
		"revoke")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_Revoke_Non_ASCII(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-d", "légô.wtf",
		"--tls",
		"--tls.port", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}

	output, err = load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-s", "https://localhost:14000/dir",
		"-d", "légô.wtf",
		"--tls",
		"--tls.port", ":5001",
		"revoke")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
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

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	user.registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"acme.wtf"},
		Bundle:  true,
	}
	resource, err := client.Certificate.Obtain(request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
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

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	user.registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"acme.wtf"},
		Bundle:  true,
		Profile: "shortlived",
	}
	resource, err := client.Certificate.Obtain(request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
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

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	user.registration = reg

	now := time.Now().UTC()

	request := certificate.ObtainRequest{
		Domains:   []string{"acme.wtf"},
		NotBefore: now.Add(1 * time.Hour),
		NotAfter:  now.Add(2 * time.Hour),
		Bundle:    true,
	}
	resource, err := client.Certificate.Obtain(request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
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

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	require.NoError(t, err)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	user.registration = reg

	resource, err := client.Registration.QueryRegistration()
	require.NoError(t, err)

	require.NotNil(t, resource)

	assert.Equal(t, "valid", resource.Body.Status)
	assert.Regexp(t, `https://localhost:14000/list-orderz/[\w\d]+`, resource.Body.Orders)
	assert.Regexp(t, `https://localhost:14000/my-account/[\w\d]+`, resource.URI)
}

func TestChallengeTLS_Client_Obtain(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)
	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	require.NoError(t, err)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	user.registration = reg

	// https://github.com/letsencrypt/pebble/issues/285
	privateKeyCSR, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	request := certificate.ObtainRequest{
		Domains:    []string{"acme.wtf"},
		Bundle:     true,
		PrivateKey: privateKeyCSR,
	}
	resource, err := client.Certificate.Obtain(request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
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

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	require.NoError(t, err)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	user.registration = reg

	csrRaw, err := os.ReadFile("./fixtures/csr.raw")
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrRaw)
	require.NoError(t, err)

	resource, err := client.Certificate.ObtainForCSR(certificate.ObtainForCSRRequest{
		CSR:    csr,
		Bundle: true,
	})
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
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

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	require.NoError(t, err)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	user.registration = reg

	csrRaw, err := os.ReadFile("./fixtures/csr.raw")
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrRaw)
	require.NoError(t, err)

	resource, err := client.Certificate.ObtainForCSR(certificate.ObtainForCSRRequest{
		CSR:     csr,
		Bundle:  true,
		Profile: "shortlived",
	})
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.NotEmpty(t, resource.CSR)
}

func TestRegistrar_UpdateAccount(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)
	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{
		privateKey: privateKey,
		email:      "foo@example.com",
	}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	regOptions := registration.RegisterOptions{TermsOfServiceAgreed: true}
	reg, err := client.Registration.Register(regOptions)
	require.NoError(t, err)
	require.Equal(t, []string{"mailto:foo@example.com"}, reg.Body.Contact)
	user.registration = reg

	user.email = "bar@example.com"
	resource, err := client.Registration.UpdateRegistration(regOptions)
	require.NoError(t, err)
	require.Equal(t, []string{"mailto:bar@example.com"}, resource.Body.Contact)
	require.Equal(t, reg.URI, resource.URI)
}

type fakeUser struct {
	email        string
	privateKey   crypto.PrivateKey
	registration *registration.Resource
}

func (f *fakeUser) GetEmail() string                        { return f.email }
func (f *fakeUser) GetRegistration() *registration.Resource { return f.registration }
func (f *fakeUser) GetPrivateKey() crypto.PrivateKey        { return f.privateKey }
