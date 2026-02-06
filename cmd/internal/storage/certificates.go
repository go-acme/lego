package storage

import (
	"crypto"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/go-acme/lego/v5/certcrypto"
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
//
// rootPath:
//
//	./.lego/certificates/
//	     │      └── root certificates directory
//	     └── "path" option
//
// archivePath:
//
//	./.lego/archives/
//	     │      └── archived certificates directory
//	     └── "path" option
type CertificatesStorage struct {
	rootPath    string
	archivePath string
}

// NewCertificatesStorage create a new certificates storage.
func NewCertificatesStorage(basePath string) *CertificatesStorage {
	return &CertificatesStorage{
		rootPath:    getCertificatesRootPath(basePath),
		archivePath: getCertificatesArchivePath(basePath),
	}
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
// FIXME rename
func sanitizedDomain(domain string) string {
	safe, err := idna.ToASCII(strings.NewReplacer(":", "-", "*", "_").Replace(domain))
	if err != nil {
		log.Fatal("Could not sanitize the local certificate ID.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	return strings.Join(
		strings.FieldsFunc(safe, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' && r != '_' && r != '.'
		}),
		"",
	)
}

// ReadPrivateKeyFile reads a private key file.
func ReadPrivateKeyFile(filename string) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading the private key: %w", err)
	}

	privateKey, err := certcrypto.ParsePEMPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing the private key: %w", err)
	}

	return privateKey, nil
}
