package storage

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-zglob"
)

const maxTimeBeforeCleaning = 30 * 24 * time.Hour

const (
	baseArchivesFolderName = "archives"
)

// Archiver manages the archiving of accounts and certificates.
type Archiver struct {
	basePath string

	maxTimeBeforeCleaning time.Duration

	accountsBasePath     string
	certificatesBasePath string

	accountsArchivePath     string
	certificatesArchivePath string
}

// NewArchiver creates a new Archiver.
func NewArchiver(basePath string) *Archiver {
	return &Archiver{
		basePath: basePath,

		maxTimeBeforeCleaning: maxTimeBeforeCleaning,

		accountsBasePath:     filepath.Join(basePath, baseAccountsRootFolderName),
		certificatesBasePath: filepath.Join(basePath, baseCertificatesFolderName),

		accountsArchivePath:     filepath.Join(basePath, baseArchivesFolderName, baseAccountsRootFolderName),
		certificatesArchivePath: filepath.Join(basePath, baseArchivesFolderName, baseCertificatesFolderName),
	}
}

func (m *Archiver) cleanArchives(pattern string) error {
	matches, err := zglob.Glob(pattern)
	if err != nil {
		return err
	}

	for _, filename := range matches {
		li := strings.LastIndex(filename, "_")

		v := strings.TrimSuffix(filename[li+1:], ".zip")

		s, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		if time.Unix(s, 0).Add(m.maxTimeBeforeCleaning).After(time.Now()) {
			log.Debug("The archive is not old enough to be cleaned.", slog.String("filename", filename))
			continue
		}

		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}

	return nil
}
