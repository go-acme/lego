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

// SaveOptions contains the options for saving a certificate.
type SaveOptions struct {
	PEM         bool
	PFX         bool
	PFXFormat   string
	PFXPassword string
}

// Validate validates the options.
func (o *SaveOptions) Validate() error {
	if o == nil {
		return nil
	}

	if o.PFX {
		switch o.PFXFormat {
		case "DES", "RC2", "SHA256":
		default:
			return fmt.Errorf("invalid PFX format: %s", o.PFXFormat)
		}
	}

	return nil
}

func (s *CertificatesStorage) CreateRootFolder() error {
	err := CreateNonExistingFolder(s.rootPath)
	if err != nil {
		return fmt.Errorf("could not check/create the root folder %q: %w", s.rootPath, err)
	}

	return nil
}

func (s *CertificatesStorage) CreateArchiveFolder() error {
	err := CreateNonExistingFolder(s.archivePath)
	if err != nil {
		return fmt.Errorf("could not check/create the archive folder %q: %w", s.archivePath, err)
	}

	return nil
}

func (s *CertificatesStorage) SaveResource(certRes *certificate.Resource, opts *SaveOptions) error {
	err := opts.Validate()
	if err != nil {
		return err
	}

	domain := certRes.Domain

	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	err = s.writeFile(domain, ExtCert, certRes.Certificate)
	if err != nil {
		return fmt.Errorf("unable to save the certificate for the domain %q: %w", domain, err)
	}

	if certRes.IssuerCertificate != nil {
		err = s.writeFile(domain, ExtIssuer, certRes.IssuerCertificate)
		if err != nil {
			return fmt.Errorf("unable to save the issuer certificate for the domain %q: %w", domain, err)
		}
	}

	// if we were given a CSR, we don't know the private key
	if certRes.PrivateKey != nil {
		err = s.writeCertificateFiles(domain, certRes, opts)
		if err != nil {
			return fmt.Errorf("unable to save the private key for the domain %q: %w", domain, err)
		}
	} else if opts != nil && (opts.PEM || opts.PFX) {
		// we don't have the private key; can't write the .pem or .pfx file
		return fmt.Errorf("unable to save PEM or PFX without the private key for the domain %q: probable usage of a CSR", domain)
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal the resource for domain %q: %w", domain, err)
	}

	err = s.writeFile(domain, ExtResource, jsonBytes)
	if err != nil {
		return fmt.Errorf("unable to save the resource for the domain %q: %w", domain, err)
	}

	return nil
}

func (s *CertificatesStorage) MoveToArchive(domain string) error {
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

func (s *CertificatesStorage) writeCertificateFiles(domain string, certRes *certificate.Resource, opts *SaveOptions) error {
	err := s.writeFile(domain, ExtKey, certRes.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to save key file: %w", err)
	}

	if opts == nil {
		return nil
	}

	if opts.PEM {
		err = s.writeFile(domain, ExtPEM, bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil))
		if err != nil {
			return fmt.Errorf("unable to save PEM file: %w", err)
		}
	}

	if opts.PFX {
		err = s.writePFXFile(domain, certRes, opts.PFXPassword, opts.PFXFormat)
		if err != nil {
			return fmt.Errorf("unable to save PFX file: %w", err)
		}
	}

	return nil
}

func (s *CertificatesStorage) writePFXFile(domain string, certRes *certificate.Resource, password, format string) error {
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

	encoder, err := getPFXEncoder(format)
	if err != nil {
		return fmt.Errorf("PFX encoder: %w", err)
	}

	pfxBytes, err := encoder.Encode(privateKey, cert, certChain, password)
	if err != nil {
		return fmt.Errorf("unable to encode PFX data for domain %s: %w", domain, err)
	}

	return s.writeFile(domain, ExtPFX, pfxBytes)
}

func (s *CertificatesStorage) writeFile(domain, extension string, data []byte) error {
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
