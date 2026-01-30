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

func TestCertificatesWriter_CreateRootFolder(t *testing.T) {
	writer, err := NewCertificatesWriter(CertificatesWriterConfig{
		BasePath: t.TempDir(),
	})
	require.NoError(t, err)

	require.NoDirExists(t, writer.rootPath)

	err = writer.CreateRootFolder()
	require.NoError(t, err)

	require.DirExists(t, writer.rootPath)
}

func TestCertificatesWriter_CreateArchiveFolder(t *testing.T) {
	writer, err := NewCertificatesWriter(CertificatesWriterConfig{
		BasePath: t.TempDir(),
	})
	require.NoError(t, err)

	require.NoDirExists(t, writer.GetArchivePath())

	err = writer.CreateArchiveFolder()
	require.NoError(t, err)

	require.DirExists(t, writer.GetArchivePath())
}

func TestCertificatesWriter_SaveResource(t *testing.T) {
	basePath := t.TempDir()

	writer, err := NewCertificatesWriter(CertificatesWriterConfig{
		BasePath: basePath,
	})
	require.NoError(t, err)

	err = os.MkdirAll(writer.rootPath, 0o755)
	require.NoError(t, err)

	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.crt"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.issuer"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.key"))
	require.NoFileExists(t, filepath.Join(basePath, baseCertificatesFolderName, "example.com.json"))

	err = writer.SaveResource(&certificate.Resource{
		Domain:            "example.com",
		CertURL:           "https://acme.example.org/cert/123",
		CertStableURL:     "https://acme.example.org/cert/456",
		PrivateKey:        []byte("PrivateKey"),
		Certificate:       []byte("Certificate"),
		IssuerCertificate: []byte("IssuerCertificate"),
		CSR:               []byte("CSR"),
	})
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

func TestCertificatesWriter_MoveToArchive(t *testing.T) {
	domain := "example.com"

	certStorage := setupCertificatesWriter(t)

	domainFiles := generateTestFiles(t, certStorage.rootPath, domain)

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.NoFileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.rootPath)
	require.NoError(t, err)
	require.Empty(t, root)

	archive, err := os.ReadDir(certStorage.GetArchivePath())
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func TestCertificatesWriter_MoveToArchive_noFileRelatedToDomain(t *testing.T) {
	domain := "example.com"

	certStorage := setupCertificatesWriter(t)

	domainFiles := generateTestFiles(t, certStorage.rootPath, "example.org")

	err := certStorage.MoveToArchive(domain)
	require.NoError(t, err)

	for _, file := range domainFiles {
		assert.FileExists(t, file)
	}

	root, err := os.ReadDir(certStorage.rootPath)
	require.NoError(t, err)
	assert.Len(t, root, len(domainFiles))

	archive, err := os.ReadDir(certStorage.GetArchivePath())
	require.NoError(t, err)

	assert.Empty(t, archive)
}

func TestCertificatesWriter_MoveToArchive_ambiguousDomain(t *testing.T) {
	domain := "example.com"

	certStorage := setupCertificatesWriter(t)

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

	archive, err := os.ReadDir(certStorage.GetArchivePath())
	require.NoError(t, err)

	require.Len(t, archive, len(domainFiles))
	assert.Regexp(t, `\d+\.`+regexp.QuoteMeta(domain), archive[0].Name())
}

func TestCertificatesWriter_GetArchivePath(t *testing.T) {
	basePath := t.TempDir()

	writer, err := NewCertificatesWriter(CertificatesWriterConfig{
		BasePath: basePath,
	})
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(basePath, baseArchivesFolderName), writer.GetArchivePath())
}

func TestCertificatesWriter_IsPEM(t *testing.T) {
	testCases := []struct {
		desc   string
		pem    bool
		assert assert.BoolAssertionFunc
	}{
		{
			desc:   "PEM enable",
			pem:    true,
			assert: assert.True,
		},
		{
			desc:   "PEM disable",
			pem:    false,
			assert: assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			writer, err := NewCertificatesWriter(CertificatesWriterConfig{
				BasePath: t.TempDir(),
				PEM:      test.pem,
			})
			require.NoError(t, err)

			test.assert(t, writer.IsPEM())
		})
	}
}

func TestCertificatesWriter_IsPFX(t *testing.T) {
	testCases := []struct {
		desc   string
		pfx    bool
		assert assert.BoolAssertionFunc
	}{
		{
			desc:   "PFX enable",
			pfx:    true,
			assert: assert.True,
		},
		{
			desc:   "PFX disable",
			pfx:    false,
			assert: assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			writer, err := NewCertificatesWriter(CertificatesWriterConfig{
				BasePath:  t.TempDir(),
				PFX:       test.pfx,
				PFXFormat: "DES",
			})
			require.NoError(t, err)

			test.assert(t, writer.IsPFX())
		})
	}
}

func assertCertificateFileContent(t *testing.T, basePath, filename string) {
	t.Helper()

	actual, err := os.ReadFile(filepath.Join(basePath, baseCertificatesFolderName, filename))
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", baseCertificatesFolderName, filename))
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}

func setupCertificatesWriter(t *testing.T) *CertificatesWriter {
	t.Helper()

	basePath := t.TempDir()

	writer, err := NewCertificatesWriter(
		CertificatesWriterConfig{
			BasePath: basePath,
		},
	)
	require.NoError(t, err)

	err = writer.CreateRootFolder()
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
