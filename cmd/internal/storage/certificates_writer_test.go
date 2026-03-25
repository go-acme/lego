package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificatesStorage_Save(t *testing.T) {
	basePath := t.TempDir()

	writer := NewCertificatesStorage(basePath)

	err := os.MkdirAll(writer.rootPath, 0o755)
	require.NoError(t, err)

	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.crt"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.issuer"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.key"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.json"))

	resource := &certificate.Resource{
		ID:                "example.com",
		Domains:           []string{"example.com"},
		KeyType:           "EC256",
		CertURL:           "https://acme.example.org/cert/123",
		CertStableURL:     "https://acme.example.org/cert/456",
		PrivateKey:        []byte("PrivateKey"),
		Certificate:       []byte("Certificate"),
		IssuerCertificate: []byte("IssuerCertificate"),
		CSR:               []byte("CSR"),
	}

	err = writer.Save(resource, nil)
	require.NoError(t, err)

	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.crt"))
	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.issuer.crt"))
	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.key"))
	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.json"))

	assertCertificateFileContent(t, basePath, "example.com.crt")
	assertCertificateFileContent(t, basePath, "example.com.issuer.crt")
	assertCertificateFileContent(t, basePath, "example.com.key")

	actual, err := os.ReadFile(filepath.Join(basePath, baseCertificatesFolderName, "example.com.json"))
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", baseCertificatesFolderName, "example.com.json"))
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
}

func TestCertificatesStorage_Save_pem(t *testing.T) {
	basePath := t.TempDir()

	writer := NewCertificatesStorage(basePath)

	err := os.MkdirAll(writer.rootPath, 0o755)
	require.NoError(t, err)

	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.crt"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.issuer"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.key"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.json"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.pem"))

	resource := &certificate.Resource{
		ID:                "example.com",
		Domains:           []string{"example.com"},
		KeyType:           "EC256",
		CertURL:           "https://acme.example.org/cert/123",
		CertStableURL:     "https://acme.example.org/cert/456",
		PrivateKey:        []byte("PrivateKey"),
		Certificate:       []byte("Certificate"),
		IssuerCertificate: []byte("IssuerCertificate"),
		CSR:               []byte("CSR"),
	}

	err = writer.Save(resource, &SaveOptions{
		PEM: true,
	})
	require.NoError(t, err)

	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.crt"))
	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.issuer.crt"))
	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.key"))
	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.json"))
	require.FileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.pem"))

	assertCertificateFileContent(t, basePath, "example.com.crt")
	assertCertificateFileContent(t, basePath, "example.com.issuer.crt")
	assertCertificateFileContent(t, basePath, "example.com.key")

	actual, err := os.ReadFile(filepath.Join(basePath, baseCertificatesFolderName, "example.com.json"))
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", baseCertificatesFolderName, "example.com.json"))
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
}
