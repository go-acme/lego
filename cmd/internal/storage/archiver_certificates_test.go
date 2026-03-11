package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiver_Certificates(t *testing.T) {
	if runtime.GOOS == "windows" {
		// The error is:
		// TempDir RemoveAll cleanup: unlinkat C:\Users\RUNNER~1\AppData\Local\Temp\xxx: The process cannot access the file because it is being used by another process.
		t.Skip("skipping test on Windows")
	}

	domain := "example.com"
	archiveDomain := "example.org"

	cfg := &configuration.Configuration{
		Storage: t.TempDir(),
		Certificates: map[string]*configuration.Certificate{
			domain: {},
		},
	}

	archiver := NewArchiver(cfg.Storage)
	archiver.maxTimeBeforeCleaning = 0

	domainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, domain)
	_ = generateFakeCertificateFiles(t, archiver.certificatesBasePath, archiveDomain)

	// archive

	err := archiver.Certificates(cfg.Certificates)
	require.NoError(t, err)

	root, err := os.ReadDir(archiver.certificatesBasePath)
	require.NoError(t, err)
	assert.Len(t, root, len(domainFiles))

	archive, err := os.ReadDir(archiver.certificatesArchivePath)
	require.NoError(t, err)

	require.Len(t, archive, 1)
	assert.Regexp(t, regexp.QuoteMeta(archiveDomain)+`_\d+\.zip`, archive[0].Name())

	// clean

	err = archiver.Certificates(cfg.Certificates)
	require.NoError(t, err)

	archive, err = os.ReadDir(archiver.certificatesArchivePath)
	require.NoError(t, err)

	require.Empty(t, archive)
}

func TestArchiver_archiveCertificate(t *testing.T) {
	if runtime.GOOS == "windows" {
		// The error is:
		// TempDir RemoveAll cleanup: unlinkat C:\Users\RUNNER~1\AppData\Local\Temp\xxx: The process cannot access the file because it is being used by another process.
		t.Skip("skipping test on Windows")
	}

	domain := "example.com"

	archiver := NewArchiver(t.TempDir())

	domainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, domain)

	err := archiver.archiveCertificate(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	root, err := os.ReadDir(archiver.certificatesBasePath)
	require.NoError(t, err)
	require.Empty(t, root)

	archive, err := os.ReadDir(archiver.certificatesArchivePath)
	require.NoError(t, err)

	require.Len(t, archive, 1)
	assert.Regexp(t, regexp.QuoteMeta(domain)+`_\d+\.zip`, archive[0].Name())
}

func TestArchiver_archiveCertificate_noFileRelatedToDomain(t *testing.T) {
	if runtime.GOOS == "windows" {
		// The error is:
		// TempDir RemoveAll cleanup: unlinkat C:\Users\RUNNER~1\AppData\Local\Temp\xxx: The process cannot access the file because it is being used by another process.
		t.Skip("skipping test on Windows")
	}

	domain := "example.com"

	archiver := NewArchiver(t.TempDir())

	domainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, "example.org")

	err := archiver.archiveCertificate(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(archiver.certificatesBasePath)
	require.NoError(t, err)
	assert.Len(t, root, len(domainFiles))

	assert.NoFileExists(t, archiver.certificatesArchivePath)
}

func TestArchiver_archiveCertificate_ambiguousDomain(t *testing.T) {
	if runtime.GOOS == "windows" {
		// The error is:
		// TempDir RemoveAll cleanup: unlinkat C:\Users\RUNNER~1\AppData\Local\Temp\xxx: The process cannot access the file because it is being used by another process.
		t.Skip("skipping test on Windows")
	}

	domain := "example.com"

	archiver := NewArchiver(t.TempDir())

	domainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, domain)
	otherDomainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, domain+".example.org")

	err := archiver.archiveCertificate(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	for _, file := range otherDomainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(archiver.certificatesBasePath)
	require.NoError(t, err)
	require.Len(t, root, len(otherDomainFiles))

	archive, err := os.ReadDir(archiver.certificatesArchivePath)
	require.NoError(t, err)

	require.Len(t, archive, 1)
	assert.Regexp(t, regexp.QuoteMeta(domain)+`_\d+\.zip`, archive[0].Name())
}

func assertCertificateFileContent(t *testing.T, basePath, filename string) {
	t.Helper()

	actual, err := os.ReadFile(filepath.Join(basePath, baseCertificatesFolderName, filename))
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", baseCertificatesFolderName, filename))
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}

func generateFakeCertificateFiles(t *testing.T, dir, domain string) []string {
	t.Helper()

	err := CreateNonExistingFolder(dir)
	require.NoError(t, err)

	var filenames []string

	for _, ext := range []string{ExtIssuer, ExtCert, ExtKey, ExtPEM, ExtPFX} {
		filename := filepath.Join(dir, domain+ext)

		err = os.WriteFile(filename, []byte("test"), 0o666)
		require.NoError(t, err)

		filenames = append(filenames, filename)
	}

	filename := filepath.Join(dir, domain+ExtResource)

	file, err := os.Create(filename)
	require.NoError(t, err)

	r := certificate.Resource{ID: domain}

	err = json.NewEncoder(file).Encode(r)
	require.NoError(t, err)

	filenames = append(filenames, filename)

	return filenames
}
