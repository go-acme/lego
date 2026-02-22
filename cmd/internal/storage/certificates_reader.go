package storage

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/log"
)

func (s *CertificatesStorage) ReadResource(certID string) (*certificate.Resource, error) {
	raw, err := s.ReadFile(certID, ExtResource)
	if err != nil {
		return nil, fmt.Errorf("unable to load resource for domain %q: %w", certID, err)
	}

	resource := new(certificate.Resource)
	if err = json.Unmarshal(raw, resource); err != nil {
		return nil, fmt.Errorf("unable to unmarshal resource for domain %q: %w", certID, err)
	}

	return resource, nil
}

func (s *CertificatesStorage) ReadCertificate(certID string) ([]*x509.Certificate, error) {
	// The input may be a bundle or a single certificate.
	return ReadCertificateFile(s.GetFileName(certID, ExtCert))
}

func (s *CertificatesStorage) ReadPrivateKey(certID string) (crypto.PrivateKey, error) {
	privateKey, err := ReadPrivateKeyFile(s.GetFileName(certID, ExtKey))
	if err != nil {
		return nil, fmt.Errorf("error while parsing the private key for %q: %w", certID, err)
	}

	return privateKey, nil
}

func (s *CertificatesStorage) ReadFile(certID, extension string) ([]byte, error) {
	return os.ReadFile(s.GetFileName(certID, extension))
}

func (s *CertificatesStorage) ExistsFile(certID, extension string) bool {
	filePath := s.GetFileName(certID, extension)

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

func (s *CertificatesStorage) GetFileName(certID, extension string) string {
	return filepath.Join(s.rootPath, SanitizedName(certID)+extension)
}
