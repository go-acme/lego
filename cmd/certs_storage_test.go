package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificatesStorage_MoveToArchive(t *testing.T) {
	domain := "example.com"

	storage := CertificatesStorage{
		rootPath:    t.TempDir(),
		archivePath: t.TempDir(),
	}

	domainFiles := generateTestFiles(t, storage.rootPath, domain)

	err := storage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	root, err := os.ReadDir(storage.rootPath)
	require.NoError(t, err)
	require.Empty(t, root)

	archive, err := os.ReadDir(storage.archivePath)
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func TestCertificatesStorage_MoveToArchive_noFileRelatedToDomain(t *testing.T) {
	domain := "example.com"

	storage := CertificatesStorage{
		rootPath:    t.TempDir(),
		archivePath: t.TempDir(),
	}

	domainFiles := generateTestFiles(t, storage.rootPath, "example.org")

	err := storage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(storage.rootPath)
	require.NoError(t, err)
	assert.Len(t, root, len(domainFiles))

	archive, err := os.ReadDir(storage.archivePath)
	require.NoError(t, err)

	assert.Empty(t, archive)
}

func TestCertificatesStorage_MoveToArchive_ambiguousDomain(t *testing.T) {
	domain := "example.com"

	storage := CertificatesStorage{
		rootPath:    t.TempDir(),
		archivePath: t.TempDir(),
	}

	domainFiles := generateTestFiles(t, storage.rootPath, domain)
	otherDomainFiles := generateTestFiles(t, storage.rootPath, domain+".example.org")

	err := storage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	for _, file := range otherDomainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(storage.rootPath)
	require.NoError(t, err)
	require.Len(t, root, len(otherDomainFiles))

	archive, err := os.ReadDir(storage.archivePath)
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func generateTestFiles(t *testing.T, dir, domain string) []string {
	t.Helper()

	var filenames []string

	for _, ext := range []string{issuerExt, certExt, keyExt, pemExt, pfxExt, resourceExt} {
		filename := filepath.Join(dir, domain+ext)
		err := os.WriteFile(filename, []byte("test"), 0o666)
		require.NoError(t, err)

		filenames = append(filenames, filename)
	}

	return filenames
}
