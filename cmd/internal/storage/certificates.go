package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v5/log"
	"golang.org/x/net/idna"
)

const (
	ExtIssuer   = ".issuer.crt"
	ExtCert     = ".crt"
	ExtKey      = ".key"
	ExtPEM      = ".pem"
	ExtPFX      = ".pfx"
	ExtResource = ".json"
)

const (
	baseCertificatesFolderName = "certificates"
	baseArchivesFolderName     = "archives"
)

// CertificatesStorage a certificates' storage.
type CertificatesStorage struct {
	*CertificatesWriter
	*CertificatesReader
}

// NewCertificatesStorage create a new certificates storage.
func NewCertificatesStorage(config CertificatesWriterConfig) (*CertificatesStorage, error) {
	writer, err := NewCertificatesWriter(config)
	if err != nil {
		return nil, fmt.Errorf("certificates storage writer: %w", err)
	}

	return &CertificatesStorage{
		CertificatesWriter: writer,
		CertificatesReader: NewCertificatesReader(config.BasePath),
	}, nil
}

func CreateNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0o700)
	} else if err != nil {
		return err
	}

	return nil
}

func getCertificatesRootPath(basePath string) string {
	return filepath.Join(basePath, baseCertificatesFolderName)
}

func getCertificatesArchivePath(basePath string) string {
	return filepath.Join(basePath, baseArchivesFolderName)
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
