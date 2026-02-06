package storage

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificatesStorage_CreateRootFolder(t *testing.T) {
	writer := NewCertificatesStorage(t.TempDir())

	require.NoDirExists(t, writer.rootPath)

	err := writer.CreateRootFolder()
	require.NoError(t, err)

	require.DirExists(t, writer.rootPath)
}

func TestCertificatesStorage_CreateArchiveFolder(t *testing.T) {
	writer := NewCertificatesStorage(t.TempDir())

	require.NoDirExists(t, writer.archivePath)

	err := writer.CreateArchiveFolder()
	require.NoError(t, err)

	require.DirExists(t, writer.archivePath)
}

func TestCertificatesStorage_SaveResource(t *testing.T) {
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
		CertURL:           "https://acme.example.org/cert/123",
		CertStableURL:     "https://acme.example.org/cert/456",
		PrivateKey:        []byte("PrivateKey"),
		Certificate:       []byte("Certificate"),
		IssuerCertificate: []byte("IssuerCertificate"),
		CSR:               []byte("CSR"),
	}

	err = writer.SaveResource(resource, nil)
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

func TestCertificatesStorage_SaveResource_pem(t *testing.T) {
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
		CertURL:           "https://acme.example.org/cert/123",
		CertStableURL:     "https://acme.example.org/cert/456",
		PrivateKey:        []byte("PrivateKey"),
		Certificate:       []byte("Certificate"),
		IssuerCertificate: []byte("IssuerCertificate"),
		CSR:               []byte("CSR"),
	}

	err = writer.SaveResource(resource, &SaveOptions{
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

func TestCertificatesStorage_MoveToArchive(t *testing.T) {
	domain := "example.com"

	certStorage := setupCertificatesStorage(t)

	domainFiles := generateTestFiles(t, certStorage.rootPath, domain)

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.rootPath)
	require.NoError(t, err)
	require.Empty(t, root)

	archive, err := os.ReadDir(certStorage.archivePath)
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func TestCertificatesStorage_MoveToArchive_noFileRelatedToDomain(t *testing.T) {
	domain := "example.com"

	certStorage := setupCertificatesStorage(t)

	domainFiles := generateTestFiles(t, certStorage.rootPath, "example.org")

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.rootPath)
	require.NoError(t, err)
	assert.Len(t, root, len(domainFiles))

	archive, err := os.ReadDir(certStorage.archivePath)
	require.NoError(t, err)

	assert.Empty(t, archive)
}

func TestCertificatesStorage_MoveToArchive_ambiguousDomain(t *testing.T) {
	domain := "example.com"

	certStorage := setupCertificatesStorage(t)

	domainFiles := generateTestFiles(t, certStorage.rootPath, domain)
	otherDomainFiles := generateTestFiles(t, certStorage.rootPath, domain+".example.org")

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	for _, file := range otherDomainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.rootPath)
	require.NoError(t, err)
	require.Len(t, root, len(otherDomainFiles))

	archive, err := os.ReadDir(certStorage.archivePath)
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func assertCertificateFileContent(t *testing.T, basePath, filename string) {
	t.Helper()

	actual, err := os.ReadFile(filepath.Join(basePath, baseCertificatesFolderName, filename))
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", baseCertificatesFolderName, filename))
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}

func setupCertificatesStorage(t *testing.T) *CertificatesStorage {
	t.Helper()

	basePath := t.TempDir()

	writer := NewCertificatesStorage(basePath)

	err := writer.CreateRootFolder()
	require.NoError(t, err)

	err = writer.CreateArchiveFolder()
	require.NoError(t, err)

	return writer
}

func generateTestFiles(t *testing.T, dir, domain string) []string {
	t.Helper()

	var filenames []string

	for _, ext := range []string{ExtIssuer, ExtCert, ExtKey, ExtPEM, ExtPFX, ExtResource} {
		filename := filepath.Join(dir, domain+ext)
		err := os.WriteFile(filename, []byte("test"), 0o666)
		require.NoError(t, err)

		filenames = append(filenames, filename)
	}

	return filenames
}
