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

// Save saves the certificate and related files.
// - the resource file (JSON)
// - the certificate file
// - the private key file (if any)
// - the issuer certificate file (if any)
// - the PFX file (if needed)
// - the PEM file (if needed).
func (s *CertificatesStorage) Save(certRes *certificate.Resource, opts *SaveOptions) error {
	err := opts.Validate()
	if err != nil {
		return err
	}

	err = CreateNonExistingFolder(s.rootPath)
	if err != nil {
		return fmt.Errorf("root folder creation: %w", err)
	}

	err = s.writeFile(certRes.ID, ExtCert, certRes.Certificate)
	if err != nil {
		return fmt.Errorf("unable to save the certificate for %q: %w", certRes.ID, err)
	}

	if certRes.IssuerCertificate != nil {
		err = s.writeFile(certRes.ID, ExtIssuer, certRes.IssuerCertificate)
		if err != nil {
			return fmt.Errorf("unable to save the issuer certificate for %q: %w", certRes.ID, err)
		}
	}

	// if we were given a CSR, we don't know the private key
	if certRes.PrivateKey != nil {
		err = s.writeCertificateFiles(certRes, opts)
		if err != nil {
			return fmt.Errorf("unable to save the private key for %q: %w", certRes.ID, err)
		}
	} else if opts != nil && (opts.PEM || opts.PFX) {
		// we don't have the private key; can't write the .pem or .pfx file
		return fmt.Errorf("unable to save PEM or PFX without the private key for %q: probable usage of a CSR", certRes.ID)
	}

	return s.saveResource(certRes)
}

// Archive moves the certificate files to the archive folder.
func (s *CertificatesStorage) Archive(domain string) error {
	err := CreateNonExistingFolder(s.archivePath)
	if err != nil {
		return fmt.Errorf("could not check/create the archive folder %q: %w", s.archivePath, err)
	}

	baseFilename := filepath.Join(s.rootPath, SanitizedName(domain))

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

func (s *CertificatesStorage) saveResource(certRes *certificate.Resource) error {
	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal the resource for %q: %w", certRes.ID, err)
	}

	err = s.writeFile(certRes.ID, ExtResource, jsonBytes)
	if err != nil {
		return fmt.Errorf("unable to save the resource for %q: %w", certRes.ID, err)
	}

	return nil
}

func (s *CertificatesStorage) writeCertificateFiles(certRes *certificate.Resource, opts *SaveOptions) error {
	err := s.writeFile(certRes.ID, ExtKey, certRes.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to save the key file: %w", err)
	}

	if opts == nil {
		return nil
	}

	if opts.PEM {
		err = s.writeFile(certRes.ID, ExtPEM, bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil))
		if err != nil {
			return fmt.Errorf("unable to save the PEM file: %w", err)
		}
	}

	if opts.PFX {
		err = s.writePFXFile(certRes, opts.PFXPassword, opts.PFXFormat)
		if err != nil {
			return fmt.Errorf("unable to save the PFX file: %w", err)
		}
	}

	return nil
}

func (s *CertificatesStorage) writePFXFile(certRes *certificate.Resource, password, format string) error {
	certPemBlock, _ := pem.Decode(certRes.Certificate)
	if certPemBlock == nil {
		return fmt.Errorf("unable to parse certificate %q", certRes.ID)
	}

	cert, err := x509.ParseCertificate(certPemBlock.Bytes)
	if err != nil {
		return fmt.Errorf("unable to load certificate %q: %w", certRes.ID, err)
	}

	certChain, err := getCertificateChain(certRes)
	if err != nil {
		return fmt.Errorf("unable to get certificate chain %q: %w", certRes.ID, err)
	}

	privateKey, err := certcrypto.ParsePEMPrivateKey(certRes.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse private ky %q: %w", certRes.ID, err)
	}

	encoder, err := getPFXEncoder(format)
	if err != nil {
		return fmt.Errorf("PFX encoder: %w", err)
	}

	pfxBytes, err := encoder.Encode(privateKey, cert, certChain, password)
	if err != nil {
		return fmt.Errorf("unable to encode PFX data %q: %w", certRes.ID, err)
	}

	return s.writeFile(certRes.ID, ExtPFX, pfxBytes)
}

func (s *CertificatesStorage) writeFile(domain, extension string, data []byte) error {
	filePath := s.GetFileName(domain, extension)

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
			return nil, fmt.Errorf("unable to parse chain certificate: %w", err)
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
