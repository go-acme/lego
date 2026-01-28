package storage

import (
	"crypto/x509"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/log"
	"golang.org/x/net/idna"
)

type CertificatesReader struct {
	rootPath string
}

func NewCertificatesReader(basePath string) *CertificatesReader {
	return &CertificatesReader{
		rootPath: getCertificatesRootPath(basePath),
	}
}

func (s *CertificatesReader) ReadResource(domain string) certificate.Resource {
	raw, err := s.ReadFile(domain, ExtResource)
	if err != nil {
		log.Fatal("Error while loading the metadata.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	var resource certificate.Resource
	if err = json.Unmarshal(raw, &resource); err != nil {
		log.Fatal("Error while marshaling the metadata.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	return resource
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

// sanitizedDomain Make sure no funny chars are in the cert names (like wildcards ;)).
func sanitizedDomain(domain string) string {
	safe, err := idna.ToASCII(strings.NewReplacer(":", "-", "*", "_").Replace(domain))
	if err != nil {
		log.Fatal("Could not sanitize the domain.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	return safe
}
