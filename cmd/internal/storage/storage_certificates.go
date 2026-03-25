package storage

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
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
	rootPath string
}

// NewCertificatesStorage create a new certificates storage.
func NewCertificatesStorage(basePath string) *CertificatesStorage {
	return &CertificatesStorage{
		rootPath: filepath.Join(basePath, baseCertificatesFolderName),
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

// SanitizedName Make sure no funny chars are in the cert names (like wildcards ;)).
func SanitizedName(name string) string {
	safe, err := idna.ToASCII(strings.NewReplacer(":", "-", "*", "_").Replace(name))
	if err != nil {
		log.Fatal("Could not sanitize the name.",
			slog.String("name", name),
			log.ErrorAttr(err),
		)
	}

	return strings.Join(
		strings.FieldsFunc(safe, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' && r != '_' && r != '.' && r != '@'
		}),
		"",
	)
}

// ReadPrivateKeyFile reads a private key file.
func ReadPrivateKeyFile(filename string) (crypto.Signer, error) {
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

// ReadCertificateFile reads a certificate file.
func ReadCertificateFile(filename string) ([]*x509.Certificate, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading the certificate: %w", err)
	}

	certs, err := certcrypto.ParsePEMBundle(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing the certificate: %w", err)
	}

	return certs, nil
}

// ReadCSRFile reads a CSR file.
func ReadCSRFile(filename string) (*x509.CertificateRequest, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	raw := bytes

	// see if we can find a PEM-encoded CSR
	var p *pem.Block

	rest := bytes
	for {
		// decode a PEM block
		p, rest = pem.Decode(rest)

		// did we fail?
		if p == nil {
			break
		}

		// did we get a CSR?
		if p.Type == "CERTIFICATE REQUEST" || p.Type == "NEW CERTIFICATE REQUEST" {
			raw = p.Bytes
		}
	}

	// no PEM-encoded CSR
	// assume we were given a DER-encoded ASN.1 CSR
	// (if this assumption is wrong, parsing these bytes will fail)
	return x509.ParseCertificateRequest(raw)
}
