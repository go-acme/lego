//go:build go1.27

package e2e

import (
	"crypto/mldsa"
	"os"
	"testing"

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

func TestChallengeHTTP_Run_MLDSA(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"-d", testDomain1,
		"--key-type", certcrypto.MLDSA44.String(),
		"--http",
		"--http.address", ":5002",
	)
	require.NoError(t, err)
}

func TestChallengeHTTP_Client_Obtain_MLDSA(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := mldsa.GenerateKey(mldsa.MLDSA44())
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
		KeyType: certcrypto.MLDSA65,
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

	cert, err := certcrypto.ParsePEMCertificate(resource.Certificate)
	require.NoError(t, err)

	certKeyType, err := certcrypto.GetCertificateKeyType(cert)
	require.NoError(t, err)

	assert.Equal(t, certcrypto.MLDSA65, certKeyType)
}
