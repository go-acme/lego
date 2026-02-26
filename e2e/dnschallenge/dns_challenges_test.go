package dnschallenge

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/dns01"
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

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		HealthCheckURL: "https://localhost:15000/dir",
		Args:           []string{"-strict", "-config", "fixtures/pebble-config-dns.json", "-dnsserver", "localhost:8053"},
		Env:            []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
		Dir:            "../",
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=../fixtures/certs/pebble.minica.pem",
		"EXEC_PATH=../fixtures/update-dns.sh",
		"LEGO_DEBUG_ACME_HTTP_CLIENT=1",
	},
	ChallSrv: &loader.CmdOption{
		Args: []string{"-http01", ":5012", "-tlsalpn01", ":5011"},
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(context.Background(), m))
}

func TestDNSHelp(t *testing.T) {
	output, err := load.RunLegoCombinedOutput(t.Context(), "dnshelp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}

func TestChallengeDNS_Run(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"--accept-tos",
		"--dns", "exec",
		"--dns.resolvers", ":8053",
		"--dns.propagation.wait", "0",
		"-s", "https://localhost:15000/dir",
		"-d", testDomain2,
		"-d", testDomain1,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeDNS_Client_Obtain(t *testing.T) {
	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")
	t.Setenv("EXEC_PATH", "../fixtures/update-dns.sh")

	defer func() { _ = os.Unsetenv("EXEC_PATH") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = "https://localhost:15000/dir"

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	t.Setenv("EXEC_SEQUENCE_INTERVAL", "5")

	provider, err := dns.NewDNSChallengeProviderByName("exec")
	require.NoError(t, err)

	mockDefault(t)

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.DisableAuthoritativeNssPropagationRequirement())
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.registration = reg

	domains := []string{testDomain2, testDomain1}

	// https://github.com/letsencrypt/pebble/issues/285
	privateKeyCSR, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	request := certificate.ObtainRequest{
		Domains:    domains,
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

	defer func() { _ = os.Unsetenv("EXEC_PATH") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = "https://localhost:15000/dir"

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	t.Setenv("EXEC_SEQUENCE_INTERVAL", "5")

	provider, err := dns.NewDNSChallengeProviderByName("exec")
	require.NoError(t, err)

	mockDefault(t)

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.DisableAuthoritativeNssPropagationRequirement())
	require.NoError(t, err)

	ctx := t.Context()

	reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)

	user.registration = reg

	domains := []string{testDomain2, testDomain1}

	// https://github.com/letsencrypt/pebble/issues/285
	privateKeyCSR, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	request := certificate.ObtainRequest{
		Domains:    domains,
		Bundle:     true,
		PrivateKey: privateKeyCSR,
		Profile:    "shortlived",
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

type fakeUser struct {
	email        string
	privateKey   crypto.PrivateKey
	registration *registration.Resource
}

func (f *fakeUser) GetEmail() string                        { return f.email }
func (f *fakeUser) GetRegistration() *registration.Resource { return f.registration }
func (f *fakeUser) GetPrivateKey() crypto.PrivateKey        { return f.privateKey }

func mockDefault(t *testing.T) {
	t.Helper()

	backup := dns01.DefaultClient()

	t.Cleanup(func() {
		dns01.SetDefaultClient(backup)
	})

	dns01.SetDefaultClient(dns01.NewClient(&dns01.Options{RecursiveNameservers: []string{":8053"}}))
}
