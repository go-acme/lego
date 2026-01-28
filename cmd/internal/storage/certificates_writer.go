package storage

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/log"
	"software.sslmate.com/src/go-pkcs12"
)

const filePerm os.FileMode = 0o600

type CertificatesWriterConfig struct {
	BasePath string

	PEM         bool
	PFX         bool
	PFXFormat   string
	PFXPassword string
}

// CertificatesWriter a writer of certificate files.
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
type CertificatesWriter struct {
	rootPath    string
	archivePath string

	pem bool

	pfx         bool
	pfxFormat   string
	pfxPassword string
}

// NewCertificatesWriter create a new certificates storage writer.
func NewCertificatesWriter(config CertificatesWriterConfig) (*CertificatesWriter, error) {
	if config.PFX {
		switch config.PFXFormat {
		case "DES", "RC2", "SHA256":
		default:
			return nil, fmt.Errorf("invalid PFX format: %s", config.PFXFormat)
		}
	}

	return &CertificatesWriter{
		rootPath:    getCertificatesRootPath(config.BasePath),
		archivePath: getCertificatesArchivePath(config.BasePath),
		pem:         config.PEM,
		pfx:         config.PFX,
		pfxPassword: config.PFXPassword,
		pfxFormat:   config.PFXFormat,
	}, nil
}

func (s *CertificatesWriter) CreateRootFolder() {
	err := CreateNonExistingFolder(s.rootPath)
	if err != nil {
		log.Fatal("Could not check/create the root folder",
			slog.String("filepath", s.rootPath),
			log.ErrorAttr(err),
		)
	}
}

func (s *CertificatesWriter) CreateArchiveFolder() {
	err := CreateNonExistingFolder(s.archivePath)
	if err != nil {
		log.Fatal("Could not check/create the archive folder.",
			slog.String("filepath", s.archivePath),
			log.ErrorAttr(err),
		)
	}
}

func (s *CertificatesWriter) SaveResource(certRes *certificate.Resource) {
	domain := certRes.Domain

	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	err := s.writeFile(domain, ExtCert, certRes.Certificate)
	if err != nil {
		log.Fatal("Unable to save Certificate.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	if certRes.IssuerCertificate != nil {
		err = s.writeFile(domain, ExtIssuer, certRes.IssuerCertificate)
		if err != nil {
			log.Fatal("Unable to save IssuerCertificate.",
				log.DomainAttr(domain),
				log.ErrorAttr(err),
			)
		}
	}

	// if we were given a CSR, we don't know the private key
	if certRes.PrivateKey != nil {
		err = s.writeCertificateFiles(domain, certRes)
		if err != nil {
			log.Fatal("Unable to save PrivateKey.", log.DomainAttr(domain), log.ErrorAttr(err))
		}
	} else if s.pem || s.pfx {
		// we don't have the private key; can't write the .pem or .pfx file
		log.Fatal("Unable to save PEM or PFX without the private key. Are you using a CSR?", log.DomainAttr(domain))
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		log.Fatal("Unable to marshal CertResource.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	err = s.writeFile(domain, ExtResource, jsonBytes)
	if err != nil {
		log.Fatal("Unable to save CertResource.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}
}

func (s *CertificatesWriter) MoveToArchive(domain string) error {
	baseFilename := filepath.Join(s.rootPath, sanitizedDomain(domain))

	matches, err := filepath.Glob(baseFilename + ".*")
	if err != nil {
		return err
	}

	for _, oldFile := range matches {
		if strings.TrimSuffix(oldFile, filepath.Ext(oldFile)) != baseFilename && oldFile != baseFilename+ExtIssuer {
			continue
		}

		date := strconv.FormatInt(time.Now().Unix(), 10)
		filename := date + "." + filepath.Base(oldFile)
		newFile := filepath.Join(s.archivePath, filename)

		err = os.Rename(oldFile, newFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *CertificatesWriter) GetArchivePath() string {
	return s.archivePath
}

func (s *CertificatesWriter) IsPEM() bool {
	return s.pem
}

func (s *CertificatesWriter) IsPFX() bool {
	return s.pfx
}

func (s *CertificatesWriter) writeCertificateFiles(domain string, certRes *certificate.Resource) error {
	err := s.writeFile(domain, ExtKey, certRes.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to save key file: %w", err)
	}

	if s.pem {
		err = s.writeFile(domain, ExtPEM, bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil))
		if err != nil {
			return fmt.Errorf("unable to save PEM file: %w", err)
		}
	}

	if s.pfx {
		err = s.writePFXFile(domain, certRes)
		if err != nil {
			return fmt.Errorf("unable to save PFX file: %w", err)
		}
	}

	return nil
}

func (s *CertificatesWriter) writePFXFile(domain string, certRes *certificate.Resource) error {
	certPemBlock, _ := pem.Decode(certRes.Certificate)
	if certPemBlock == nil {
		return fmt.Errorf("unable to parse Certificate for domain %s", domain)
	}

	cert, err := x509.ParseCertificate(certPemBlock.Bytes)
	if err != nil {
		return fmt.Errorf("unable to load Certificate for domain %s: %w", domain, err)
	}

	certChain, err := getCertificateChain(certRes)
	if err != nil {
		return fmt.Errorf("unable to get certificate chain for domain %s: %w", domain, err)
	}

	privateKey, err := certcrypto.ParsePEMPrivateKey(certRes.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse PrivateKey for domain %s: %w", domain, err)
	}

	encoder, err := getPFXEncoder(s.pfxFormat)
	if err != nil {
		return fmt.Errorf("PFX encoder: %w", err)
	}

	pfxBytes, err := encoder.Encode(privateKey, cert, certChain, s.pfxPassword)
	if err != nil {
		return fmt.Errorf("unable to encode PFX data for domain %s: %w", domain, err)
	}

	return s.writeFile(domain, ExtPFX, pfxBytes)
}

func (s *CertificatesWriter) writeFile(domain, extension string, data []byte) error {
	filePath := filepath.Join(s.rootPath, sanitizedDomain(domain)+extension)

	log.Info("Writing file.",
		slog.String("filepath", filePath))

	return os.WriteFile(filePath, data, filePerm)
}

func getCertificateChain(certRes *certificate.Resource) ([]*x509.Certificate, error) {
	chainCertPemBlock, rest := pem.Decode(certRes.IssuerCertificate)
	if chainCertPemBlock == nil {
		return nil, errors.New("unable to parse Issuer Certificate")
	}

	var certChain []*x509.Certificate

	for chainCertPemBlock != nil {
		chainCert, err := x509.ParseCertificate(chainCertPemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unable to parse Chain Certificate: %w", err)
		}

		certChain = append(certChain, chainCert)
		chainCertPemBlock, rest = pem.Decode(rest) // Try decoding the next pem block
	}

	return certChain, nil
}

func getPFXEncoder(pfxFormat string) (*pkcs12.Encoder, error) {
	var encoder *pkcs12.Encoder

	switch pfxFormat {
	case "SHA256":
		encoder = pkcs12.Modern2023
	case "DES":
		encoder = pkcs12.LegacyDES
	case "RC2":
		encoder = pkcs12.LegacyRC2
	default:
		return nil, fmt.Errorf("invalid PFX format: %s", pfxFormat)
	}

	return encoder, nil
}
