package eab

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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

const (
	testDomain1 = "acme.localhost"
)

const (
	testEmail1 = "lego@example.com"
)

const caDirectory = "https://localhost:16000/dir"

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		HealthCheckURL: caDirectory,
		Args:           []string{"-strict", "-config", "fixtures/pebble-config-eab.json"},
		Env:            []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
		Dir:            "../",
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=../fixtures/certs/pebble.minica.pem",
		"LEGO_DEBUG_ACME_HTTP_CLIENT=1",
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(context.Background(), m))
}

func TestChallengeHTTP_Run_EAB(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"run",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
		"-d", testDomain1,
		"--http",
		"--http.port", ":5006",
		"--eab",
		"--eab.kid", "kid-3",
		"--eab.hmac", "HjudV5qnbreN-n9WyFSH-t4HXuEx_XFen45zuxY-G1h6fr74V3cUM_dVlwQZBWmc",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeHTTP_Client_Obtain_EAB(t *testing.T) {
	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{PrivateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5006"))
	require.NoError(t, err)

	ctx := t.Context()

	options := registration.RegisterEABOptions{
		TermsOfServiceAgreed: true,
		Kid:                  "kid-3",
		HmacEncoded:          "HjudV5qnbreN-n9WyFSH-t4HXuEx_XFen45zuxY-G1h6fr74V3cUM_dVlwQZBWmc",
	}

	reg, err := client.Registration.RegisterWithExternalAccountBinding(ctx, options)
	require.NoError(t, err)

	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{testDomain1},
		KeyType: certcrypto.RSA2048,
		Bundle:  true,
	}

	resource, err := client.Certificate.Obtain(ctx, request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, testDomain1, resource.ID)
	assert.Equal(t, []string{testDomain1}, resource.Domains)
	assert.Regexp(t, `https://localhost:16000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:16000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.Empty(t, resource.CSR)
}
