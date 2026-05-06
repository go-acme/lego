package e2e

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/e2e/loader"
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
