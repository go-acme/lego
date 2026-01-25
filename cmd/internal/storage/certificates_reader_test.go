package storage

import (
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificatesStorage_ReadResource(t *testing.T) {
	reader := NewCertificatesStorage("testdata")

	resource, err := reader.ReadResource("example.com")
	require.NoError(t, err)

	expected := certificate.Resource{
		Domain:        "example.com",
		CertURL:       "https://acme.example.org/cert/123",
		CertStableURL: "https://acme.example.org/cert/456",
	}

	assert.Equal(t, expected, resource)
}

func TestCertificatesStorage_ExistsFile(t *testing.T) {
	reader := NewCertificatesStorage("testdata")

	testCases := []struct {
		desc      string
		domain    string
		extension string
		assert    assert.BoolAssertionFunc
	}{
		{
			desc:      "exists",
			domain:    "example.com",
			extension: ExtResource,
			assert:    assert.True,
		},
		{
			desc:      "not exists",
			domain:    "example.org",
			extension: ExtResource,
			assert:    assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.assert(t, reader.ExistsFile(test.domain, test.extension))
		})
	}
}

func TestCertificatesStorage_ReadFile(t *testing.T) {
	reader := NewCertificatesStorage("testdata")

	data, err := reader.ReadFile("example.com", ExtResource)
	assert.NoError(t, err)

	assert.NotEmpty(t, data)
}

func TestCertificatesStorage_GetRootPath(t *testing.T) {
	basePath := t.TempDir()

	reader := NewCertificatesStorage(basePath)

	rootPath := reader.GetRootPath()

	expected := filepath.Join(basePath, baseCertificatesFolderName)

	assert.Equal(t, expected, rootPath)
}

func TestCertificatesStorage_GetArchivePath(t *testing.T) {
	basePath := t.TempDir()

	writer := NewCertificatesStorage(basePath)

	assert.Equal(t, filepath.Join(basePath, baseArchivesFolderName), writer.archivePath)
}

func TestCertificatesStorage_GetFileName(t *testing.T) {
	testCases := []struct {
		desc      string
		domain    string
		extension string
		expected  string
	}{
		{
			desc:      "simple",
			domain:    "example.com",
			extension: ExtCert,
			expected:  "example.com.crt",
		},
		{
			desc:      "hyphen",
			domain:    "test-acme.example.com",
			extension: ExtResource,
			expected:  "test-acme.example.com.json",
		},
		{
			desc:      "wildcard",
			domain:    "*.example.com",
			extension: ExtKey,
			expected:  "_.example.com.key",
		},
		{
			desc:      "colon",
			domain:    "acme:test.example.com",
			extension: ExtResource,
			expected:  "acme-test.example.com.json",
		},
		{
			desc:      "IDN",
			domain:    "测试.com",
			extension: ExtResource,
			expected:  "xn--0zwm56d.com.json",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()

			reader := NewCertificatesStorage(basePath)

			filename := reader.GetFileName(test.domain, test.extension)

			expected := filepath.Join(basePath, baseCertificatesFolderName, test.expected)

			assert.Equal(t, expected, filename)
		})
	}
}

func TestCertificatesStorage_ReadCertificate(t *testing.T) {
	reader := NewCertificatesStorage("testdata")

	cert, err := reader.ReadCertificate("example.org", ExtCert)
	assert.NoError(t, err)

	assert.NotEmpty(t, cert)
}
