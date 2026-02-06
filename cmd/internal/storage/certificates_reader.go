package storage

import (
	"crypto"
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

func (s *CertificatesStorage) ReadResource(domain string) (*certificate.Resource, error) {
	raw, err := s.ReadFile(domain, ExtResource)
	if err != nil {
		return nil, fmt.Errorf("unable to load resource for domain %q: %w", domain, err)
	}

	resource := new(certificate.Resource)
	if err = json.Unmarshal(raw, resource); err != nil {
		return nil, fmt.Errorf("unable to unmarshal resource for domain %q: %w", domain, err)
	}

	return resource, nil
}

func (s *CertificatesStorage) ReadCertificate(domain string) ([]*x509.Certificate, error) {
	content, err := s.ReadFile(domain, ExtCert)
	if err != nil {
		return nil, err
	}

	// The input may be a bundle or a single certificate.
	return certcrypto.ParsePEMBundle(content)
}

func (s *CertificatesStorage) ReadPrivateKey(domain string) (crypto.PrivateKey, error) {
	privateKey, err := ReadPrivateKeyFile(s.GetFileName(domain, ExtKey))
	if err != nil {
		return nil, fmt.Errorf("error while parsing the private key for %q: %w", domain, err)
	}

	return privateKey, nil
}

func (s *CertificatesStorage) ReadFile(domain, extension string) ([]byte, error) {
	return os.ReadFile(s.GetFileName(domain, extension))
}

func (s *CertificatesStorage) ExistsFile(domain, extension string) bool {
	filePath := s.GetFileName(domain, extension)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatal("File stat", slog.String("filepath", filePath), log.ErrorAttr(err))
	}

	return true
}

func (s *CertificatesStorage) GetRootPath() string {
	return s.rootPath
}

func (s *CertificatesStorage) GetFileName(domain, extension string) string {
	return filepath.Join(s.rootPath, sanitizedDomain(domain)+extension)
}
