package storage

import (
	"os"
	"path/filepath"
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

func getCertificatesRootPath(basePath string) string {
	return filepath.Join(basePath, baseCertificatesFolderName)
}

func getCertificatesArchivePath(basePath string) string {
	return filepath.Join(basePath, baseArchivesFolderName)
}

func CreateNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0o700)
	} else if err != nil {
		return err
	}

	return nil
}
