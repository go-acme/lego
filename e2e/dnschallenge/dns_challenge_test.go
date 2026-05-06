package dnschallenge

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/e2e/internal"
	"github.com/go-acme/lego/v5/e2e/loader"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDomain1 = "légo.localhost"
	testDomain2 = "*.légo.localhost"
)

func TestChallengeDNS_Run(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"--accept-tos",
		"--dns", "exec",
		"--dns.resolvers", ":8553",
		"--dns.propagation.wait", "0",
		"--env-file", "./fixtures/.env",
		"-s", caDirectory,
		"-d", testDomain2,
		"-d", testDomain1,
	)
	require.NoError(t, err)
}

func TestChallengeDNS_Client_Obtain(t *testing.T) {
	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")
	t.Setenv("EXEC_PATH", "../fixtures/update-dns.sh")
	t.Setenv("EXEC_SEQUENCE_INTERVAL", "5")

	defer func() { _ = os.Unsetenv("EXEC_PATH") }()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = caDirectory

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	provider, err := dns.NewDNSChallengeProviderByName("exec")
	require.NoError(t, err)

	mockDefault(t)

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.DisableAuthoritativeNssPropagationRequirement())
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	domains := []string{testDomain2, testDomain1}

	// https://github.com/letsencrypt/pebble/issues/285
	privateKeyCSR, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	request := certificate.ObtainRequest{
		Domains:    domains,
		KeyType:    certcrypto.RSA2048,
		Bundle:     true,
		PrivateKey: privateKeyCSR,
	}
	resource, err := client.Certificate.Obtain(ctx, request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "*.xn--lgo-bma.localhost", resource.ID)
	assert.Equal(t, []string{"*.xn--lgo-bma.localhost", "xn--lgo-bma.localhost"}, resource.Domains)
	assert.Regexp(t, `https://localhost:15000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:15000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.Empty(t, resource.CSR)
}

func TestChallengeDNS_Client_Obtain_profile(t *testing.T) {
	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")
	t.Setenv("EXEC_PATH", "../fixtures/update-dns.sh")
	t.Setenv("EXEC_SEQUENCE_INTERVAL", "5")

	defer func() { _ = os.Unsetenv("EXEC_PATH") }()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = caDirectory

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	provider, err := dns.NewDNSChallengeProviderByName("exec")
	require.NoError(t, err)

	mockDefault(t)

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.DisableAuthoritativeNssPropagationRequirement())
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.Registration = reg

	domains := []string{testDomain2, testDomain1}

	// https://github.com/letsencrypt/pebble/issues/285
	privateKeyCSR, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	request := certificate.ObtainRequest{
		Domains:    domains,
		KeyType:    certcrypto.RSA2048,
		Bundle:     true,
		PrivateKey: privateKeyCSR,
		Profile:    "shortlived",
	}
	resource, err := client.Certificate.Obtain(ctx, request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "*.xn--lgo-bma.localhost", resource.ID)
	assert.Equal(t, []string{"*.xn--lgo-bma.localhost", "xn--lgo-bma.localhost"}, resource.Domains)
	assert.Equal(t, "shortlived", resource.Profile)
	assert.Regexp(t, `https://localhost:15000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:15000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.Empty(t, resource.CSR)
}

func mockDefault(t *testing.T) {
	t.Helper()

	backup := dns01.DefaultClient()

	t.Cleanup(func() {
		dns01.SetDefaultClient(backup)
	})

	dns01.SetDefaultClient(dns01.NewClient(&dns01.Options{RecursiveNameservers: []string{":8553"}}))
}
