package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-zglob"
)

func (m *Archiver) Certificates(certificates map[string]*configuration.Certificate) error {
	err := m.cleanArchivedCertificates()
	if err != nil {
		return fmt.Errorf("clean archived certificates: %w", err)
	}

	err = m.archiveRemovedCertificates(certificates)
	if err != nil {
		return fmt.Errorf("archive removed certificates: %w", err)
	}

	return nil
}

func (m *Archiver) archiveRemovedCertificates(certificates map[string]*configuration.Certificate) error {
	// Only archive the certificates that are not in the configuration.
	return m.archiveCertificates(func(resourceID string) bool {
		_, ok := certificates[resourceID]

		return ok
	})
}

func (m *Archiver) archiveCertificate(certID string) error {
	return m.archiveCertificates(func(resourceID string) bool {
		return certID != resourceID
	})
}

func (m *Archiver) archiveCertificates(skip func(resourceID string) bool) error {
	_, err := os.Stat(m.certificatesBasePath)
	if os.IsNotExist(err) {
		return nil
	}

	matches, err := zglob.Glob(filepath.Join(m.certificatesBasePath, "*.json"))
	if err != nil {
		return fmt.Errorf("search certificate files: %w", err)
	}

	date := strconv.FormatInt(time.Now().Unix(), 10)

	for _, filename := range matches {
		file, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("reading certificate file %q: %w", filename, err)
		}

		resource := new(certificate.Resource)

		err = json.Unmarshal(file, resource)
		if err != nil {
			return fmt.Errorf("unmarshalling certificate file %q: %w", filename, err)
		}

		if skip(resource.ID) {
			continue
		}

		err = m.archiveOneCertificate(filename, date, resource)
		if err != nil {
			return fmt.Errorf("archive certificate %q: %w", resource.ID, err)
		}
	}

	return nil
}

func (m *Archiver) archiveOneCertificate(filename, date string, resource *certificate.Resource) error {
	dest := filepath.Join(m.certificatesArchivePath, strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))+"_"+date+".zip")

	log.Info("Archiving certificate", log.CertNameAttr(resource.ID), slog.String("archive", dest))

	err := CreateNonExistingFolder(filepath.Dir(dest))
	if err != nil {
		return fmt.Errorf("could not check/create the certificates archive folder %q: %w", filepath.Dir(dest), err)
	}

	files, err := getCertificateFiles(filename, resource.ID)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(m.basePath, filepath.Dir(filename))
	if err != nil {
		return err
	}

	err = compressFiles(dest, files, rel)
	if err != nil {
		return fmt.Errorf("compress certificate files: %w", err)
	}

	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			return fmt.Errorf("remove certificate file %q: %w", file, err)
		}
	}

	return nil
}

func (m *Archiver) cleanArchivedCertificates() error {
	_, err := os.Stat(m.certificatesArchivePath)
	if os.IsNotExist(err) {
		return nil
	}

	return m.cleanArchives(filepath.Join(m.certificatesArchivePath, "*.zip"))
}

func getCertificateFiles(filename, resourceID string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(filepath.Dir(filename), SanitizedName(resourceID)+".*"))
	if err != nil {
		return nil, err
	}

	var restrictedFiles []string

	baseFilename := filepath.Join(filepath.Dir(filename), SanitizedName(resourceID))

	// Filter files to avoid ambiguous names (ex: foo.com and foo.com.uk)
	for _, file := range files {
		if strings.TrimSuffix(file, filepath.Ext(file)) != baseFilename && file != baseFilename+ExtIssuer {
			continue
		}

		restrictedFiles = append(restrictedFiles, file)
	}

	return restrictedFiles, nil
}
