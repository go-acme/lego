package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiver_Certificates(t *testing.T) {
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

	archives, err := archiver.ListArchivedCertificates()
	require.NoError(t, err)

	require.Len(t, archives, 1)
	assert.Regexp(t, regexp.QuoteMeta(archiveDomain)+`_\d+\.zip`, archives[0])

	// clean

	err = archiver.Certificates(cfg.Certificates)
	require.NoError(t, err)

	archives, err = archiver.ListArchivedCertificates()
	require.NoError(t, err)

	require.Empty(t, archives)
}

func TestArchiver_Restore_certificates(t *testing.T) {
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

	archives, err := archiver.ListArchivedCertificates()
	require.NoError(t, err)

	require.Len(t, archives, 1)
	assert.Regexp(t, regexp.QuoteMeta(archiveDomain)+`_\d+\.zip`, archives[0])

	// restore

	err = archiver.Restore(archives[0])
	require.NoError(t, err)

	root, err = os.ReadDir(archiver.certificatesBasePath)
	require.NoError(t, err)
	assert.Len(t, root, len(domainFiles)*2)

	archives, err = archiver.ListArchivedCertificates()
	require.NoError(t, err)

	require.Empty(t, archives)
}

func TestArchiver_Certificate(t *testing.T) {
	domain := "example.com"

	archiver := NewArchiver(t.TempDir())

	domainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, domain)

	err := archiver.Certificate(domain)
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
	domain := "example.com"

	archiver := NewArchiver(t.TempDir())

	domainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, "example.org")

	err := archiver.Certificate(domain)
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
	domain := "example.com"

	archiver := NewArchiver(t.TempDir())

	domainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, domain)
	otherDomainFiles := generateFakeCertificateFiles(t, archiver.certificatesBasePath, domain+".example.org")

	err := archiver.Certificate(domain)
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

	defer func() { _ = file.Close() }()

	r := Certificate{
		Resource: &certificate.Resource{ID: domain},
		Origin:   OriginConfiguration,
	}

	err = json.NewEncoder(file).Encode(r)
	require.NoError(t, err)

	filenames = append(filenames, filename)

	return filenames
}
