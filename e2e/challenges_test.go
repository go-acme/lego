package e2e

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/e2e/internal"
	"github.com/go-acme/lego/v5/e2e/loader"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/stretchr/testify/require"
)

const (
	testDomain1 = "acme.localhost"
	testDomain2 = "lego.localhost"
	testDomain3 = "acme.lego.localhost"
	testDomain4 = "légô.localhost"
)

const (
	testEmail1 = "lego@example.com"
	testEmail2 = "acme@example.com"
)

const caDirectory = "https://localhost:14000/dir"

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		HealthCheckURL: caDirectory,
		Args:           []string{"-strict", "-config", "fixtures/pebble-config.json"},
		Env:            []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=./fixtures/certs/pebble.minica.pem",
		"LEGO_DEBUG_ACME_HTTP_CLIENT=1",
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(context.Background(), m))
}

func TestHelp(t *testing.T) {
	output, err := load.RunLegoCombinedOutput(t.Context(), "-h")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
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

func createTestCSRFile(t *testing.T, raw bool) string {
	t.Helper()

	csr := createTestCSR(t)

	if raw {
		filename := filepath.Join(t.TempDir(), "csr.raw")

		fileRaw, err := os.Create(filename)
		require.NoError(t, err)

		defer fileRaw.Close()

		_, err = fileRaw.Write(csr)
		require.NoError(t, err)

		return filename
	}

	filename := filepath.Join(t.TempDir(), "csr.cert")

	file, err := os.Create(filename)
	require.NoError(t, err)

	defer file.Close()

	_, err = file.Write(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr}))
	require.NoError(t, err)

	return filename
}

func createTestCSR(t *testing.T) []byte {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	csr, err := certcrypto.CreateCSR(privateKey, certcrypto.CSROptions{
		Domain: testDomain1,
		SAN: []string{
			testDomain1,
			testDomain2,
		},
	})
	require.NoError(t, err)

	return csr
}
