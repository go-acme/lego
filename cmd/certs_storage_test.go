package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockCertificateStorage(t *testing.T) *CertificatesStorage {
	t.Helper()

	basePath := t.TempDir()

	writer, err := storage.NewCertificatesWriter(
		storage.CertificatesWriterConfig{
			BasePath: basePath,
		},
	)
	require.NoError(t, err)

	certStorage := &CertificatesStorage{
		CertificatesWriter: writer,
		CertificatesReader: storage.NewCertificatesReader(basePath),
	}

	certStorage.CreateRootFolder()
	certStorage.CreateArchiveFolder()

	return certStorage
}

func TestCertificatesStorage_MoveToArchive(t *testing.T) {
	domain := "example.com"

	certStorage := mockCertificateStorage(t)

	domainFiles := generateTestFiles(t, certStorage.GetRootPath(), domain)

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.GetRootPath())
	require.NoError(t, err)
	require.Empty(t, root)

	archive, err := os.ReadDir(certStorage.GetArchivePath())
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func TestCertificatesStorage_MoveToArchive_noFileRelatedToDomain(t *testing.T) {
	domain := "example.com"

	certStorage := mockCertificateStorage(t)

	domainFiles := generateTestFiles(t, certStorage.GetRootPath(), "example.org")

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.GetRootPath())
	require.NoError(t, err)
	assert.Len(t, root, len(domainFiles))

	archive, err := os.ReadDir(certStorage.GetArchivePath())
	require.NoError(t, err)

	assert.Empty(t, archive)
}

func TestCertificatesStorage_MoveToArchive_ambiguousDomain(t *testing.T) {
	domain := "example.com"

	certStorage := mockCertificateStorage(t)

	domainFiles := generateTestFiles(t, certStorage.GetRootPath(), domain)
	otherDomainFiles := generateTestFiles(t, certStorage.GetRootPath(), domain+".example.org")

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	for _, file := range otherDomainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.GetRootPath())
	require.NoError(t, err)
	require.Len(t, root, len(otherDomainFiles))

	archive, err := os.ReadDir(certStorage.GetArchivePath())
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func generateTestFiles(t *testing.T, dir, domain string) []string {
	t.Helper()

	var filenames []string

	for _, ext := range []string{storage.IssuerExt, storage.CertExt, storage.KeyExt, storage.PEMExt, storage.PFXExt, storage.ResourceExt} {
		filename := filepath.Join(dir, domain+ext)
		err := os.WriteFile(filename, []byte("test"), 0o666)
		require.NoError(t, err)

		filenames = append(filenames, filename)
	}

	return filenames
}
