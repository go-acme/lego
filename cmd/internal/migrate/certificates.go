package migrate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-zglob"
)

type oldResource struct {
	Domain        string `json:"domain"`
	CertURL       string `json:"certUrl"`
	CertStableURL string `json:"certStableUrl"`
}

func Certificates(root string) error {
	matches, err := zglob.Glob(filepath.Join(root, "certificates", "**", "*.json"))
	if err != nil {
		return err
	}

	for _, certResourceFilePath := range matches {
		data, err := os.ReadFile(certResourceFilePath)
		if err != nil {
			return fmt.Errorf("could not read the certificate file %q: %w", certResourceFilePath, err)
		}

		var oldCertRes oldResource

		err = json.Unmarshal(data, &oldCertRes)
		if err != nil {
			return fmt.Errorf("could not parse the certificate file %q: %w", certResourceFilePath, err)
		}

		if oldCertRes.Domain == "" {
			log.Error("Skip migration: the old certificate resource does not contain a domain.", slog.String("filepath", certResourceFilePath))

			continue
		}

		log.Info("Migrating a certificate.", slog.String("filepath", certResourceFilePath))

		err = migrateCertificate(oldCertRes, certResourceFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateCertificate(oldCertRes oldResource, certResourceFilePath string) error {
	certRes := certificate.Resource{
		CertURL:       oldCertRes.CertURL,
		CertStableURL: oldCertRes.CertStableURL,
	}

	baseName := strings.TrimSuffix(certResourceFilePath, filepath.Ext(certResourceFilePath))

	certPath := baseName + storage.ExtCert

	certs, err := storage.ReadCertificateFile(certPath)
	if err != nil {
		return fmt.Errorf("could not parse the certificate %q: %w", certPath, err)
	}

	cert := certs[0]

	certRes.ID, err = certcrypto.GetCertificateMainDomain(cert)
	if err != nil {
		return fmt.Errorf("could not get the certificate main domain: %w", err)
	}

	certRes.Domains = slices.Concat(cert.DNSNames, toStringSlice(cert.IPAddresses))

	f, err := os.Create(certResourceFilePath)
	if err != nil {
		return fmt.Errorf("could not open the certificate file %q: %w", certResourceFilePath, err)
	}

	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(certRes)
	if err != nil {
		return fmt.Errorf("could not encode the certificate file %q: %w", certResourceFilePath, err)
	}

	return nil
}

// TODO(ldez) factorize with storage?
func toStringSlice[T fmt.Stringer](values []T) []string {
	var s []string

	for _, value := range values {
		s = append(s, value.String())
	}

	return s
}
