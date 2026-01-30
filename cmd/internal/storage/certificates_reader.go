package storage

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/log"
)

type CertificatesReader struct {
	rootPath string
}

func NewCertificatesReader(basePath string) *CertificatesReader {
	return &CertificatesReader{
		rootPath: getCertificatesRootPath(basePath),
	}
}

func (s *CertificatesReader) ReadResource(domain string) (certificate.Resource, error) {
	raw, err := s.ReadFile(domain, ExtResource)
	if err != nil {
		return certificate.Resource{}, fmt.Errorf("unable to load resource for domain %q: %w", domain, err)
	}

	var resource certificate.Resource
	if err = json.Unmarshal(raw, &resource); err != nil {
		return certificate.Resource{}, fmt.Errorf("unable to unmarshal resource for domain %q: %w", domain, err)
	}

	return resource, nil
}

func (s *CertificatesReader) ExistsFile(domain, extension string) bool {
	filePath := s.GetFileName(domain, extension)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatal("File stat", slog.String("filepath", filePath), log.ErrorAttr(err))
	}

	return true
}

func (s *CertificatesReader) ReadFile(domain, extension string) ([]byte, error) {
	return os.ReadFile(s.GetFileName(domain, extension))
}

func (s *CertificatesReader) GetRootPath() string {
	return s.rootPath
}

func (s *CertificatesReader) GetFileName(domain, extension string) string {
	filename := sanitizedDomain(domain) + extension
	return filepath.Join(s.rootPath, filename)
}

func (s *CertificatesReader) ReadCertificate(domain, extension string) ([]*x509.Certificate, error) {
	content, err := s.ReadFile(domain, extension)
	if err != nil {
		return nil, err
	}

	// The input may be a bundle or a single certificate.
	return certcrypto.ParsePEMBundle(content)
}
