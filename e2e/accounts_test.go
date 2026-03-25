package e2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"

	"github.com/go-acme/lego/v5/e2e/internal"
	"github.com/go-acme/lego/v5/e2e/loader"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/stretchr/testify/require"
)

func TestAccount_Register(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"accounts", "register",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
	)
	require.NoError(t, err)
}

func TestRegistrar_UpdateAccount(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{
		PrivateKey: privateKey,
		Email:      testEmail1,
	}

	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	ctx := t.Context()

	regOptions := registration.RegisterOptions{TermsOfServiceAgreed: true}
	reg, err := client.Registration.Register(ctx, regOptions)
	require.NoError(t, err)
	require.Equal(t, []string{"mailto:" + testEmail1}, reg.Contact)
	user.Registration = reg

	user.Email = testEmail2
	resource, err := client.Registration.UpdateRegistration(ctx, regOptions)
	require.NoError(t, err)
	require.Equal(t, []string{"mailto:" + testEmail2}, resource.Contact)
	require.Equal(t, reg.Location, resource.Location)
}

func TestRegistrar_KeyRollover(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	oldKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &internal.FakeUser{
		PrivateKey: oldKey,
		Email:      testEmail1,
	}

	config := lego.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	ctx := t.Context()

	regOptions := registration.RegisterOptions{TermsOfServiceAgreed: true}
	_, err = client.Registration.Register(ctx, regOptions)
	require.NoError(t, err)

	newKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	err = client.Registration.KeyRollover(ctx, newKey)
	require.NoError(t, err)
}
