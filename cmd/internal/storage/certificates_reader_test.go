package storage

import (
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/stretchr/testify/assert"
)

func TestNewCertificatesWriter_ReadResource(t *testing.T) {
	reader := NewCertificatesReader("testdata")

	resource := reader.ReadResource("example.com")

	expected := certificate.Resource{
		Domain:        "example.com",
		CertURL:       "https://acme.example.org/cert/123",
		CertStableURL: "https://acme.example.org/cert/456",
	}

	assert.Equal(t, expected, resource)
}

func TestNewCertificatesWriter_ExistsFile(t *testing.T) {
	reader := NewCertificatesReader("testdata")

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

func TestNewCertificatesWriter_ReadFile(t *testing.T) {
	reader := NewCertificatesReader("testdata")

	data, err := reader.ReadFile("example.com", ExtResource)
	assert.NoError(t, err)

	assert.NotEmpty(t, data)
}

func TestNewCertificatesWriter_GetRootPath(t *testing.T) {
	basePath := t.TempDir()

	reader := NewCertificatesReader(basePath)

	rootPath := reader.GetRootPath()

	expected := filepath.Join(basePath, baseCertificatesFolderName)

	assert.Equal(t, expected, rootPath)
}

func TestNewCertificatesWriter_GetFileName(t *testing.T) {
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

			reader := NewCertificatesReader(basePath)

			filename := reader.GetFileName(test.domain, test.extension)

			expected := filepath.Join(basePath, baseCertificatesFolderName, test.expected)

			assert.Equal(t, expected, filename)
		})
	}
}

func TestNewCertificatesWriter_ReadCertificate(t *testing.T) {
	reader := NewCertificatesReader("testdata")

	cert, err := reader.ReadCertificate("example.org", ExtCert)
	assert.NoError(t, err)

	assert.NotEmpty(t, cert)
}
