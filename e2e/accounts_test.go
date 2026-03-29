package e2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"os"
	"path/filepath"
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

func TestAccount_Recover(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"accounts", "register",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
	)
	require.NoError(t, err)

	file, err := os.ReadFile(filepath.FromSlash(".lego/accounts/localhost_14000/lego@example.com/lego@example.com.key"))
	require.NoError(t, err)

	privateKeyPath := filepath.Join(t.TempDir(), "foo.key")

	err = os.WriteFile(privateKeyPath, file, 0o600)
	require.NoError(t, err)

	// Delete the account directory
	err = os.RemoveAll(filepath.FromSlash(".lego/accounts/"))
	require.NoError(t, err)

	stdinReader, stdinWriter := io.Pipe()

	defer func() { _ = stdinReader.Close() }()

	go func() {
		defer func() { _ = stdinWriter.Close() }()

		_, err = io.WriteString(stdinWriter, "Y\n")
	}()

	err = load.RunLegoWithInput(t.Context(),
		stdinReader,
		"accounts", "recover",
		"-m", testEmail1,
		"-s", caDirectory,
		"--private-key", privateKeyPath,
	)
	require.NoError(t, err)
}

func TestAccount_KeyRollover(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"accounts", "register",
		"-m", testEmail1,
		"--accept-tos",
		"-s", caDirectory,
	)
	require.NoError(t, err)

	stdinReader, stdinWriter := io.Pipe()

	defer func() { _ = stdinReader.Close() }()

	go func() {
		defer func() { _ = stdinWriter.Close() }()

		_, err = io.WriteString(stdinWriter, "Y\n")
	}()

	err = load.RunLegoWithInput(t.Context(),
		stdinReader,
		"accounts", "keyrollover",
		"-m", testEmail1,
		"-s", caDirectory,
		"--key-type", "rsa2048",
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
