package storage

import (
	"archive/zip"
	"fmt"
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

func (m *Archiver) Restore(archivePath string) error {
	err := m.restore(archivePath)
	if err != nil {
		return err
	}

	err = os.RemoveAll(archivePath)
	if err != nil {
		return err
	}

	return nil
}

func (m *Archiver) restore(archivePath string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("could not open archive %q: %w", archivePath, err)
	}

	defer func() { _ = reader.Close() }()

	baseDir := filepath.Join(m.basePath, reader.Comment)

	for _, file := range reader.File {
		err = extractZipFile(baseDir, file)
		if err != nil {
			return fmt.Errorf("extract file %q: %w", file.Name, err)
		}
	}

	return nil
}

func (m *Archiver) cleanArchives(pattern string) error {
	matches, err := zglob.Glob(pattern)
	if err != nil {
		return err
	}

	for _, filename := range matches {
		date, err := parseArchiveDate(filename)
		if err != nil {
			log.Error("The date of the archive cannot be parsed: the file is ignored.",
				slog.String("filename", filename),
				log.ErrorAttr(err),
			)

			continue
		}

		if date.Add(m.maxTimeBeforeCleaning).After(time.Now()) {
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

func listArchives(archivePath string) ([]string, error) {
	_, err := os.Stat(archivePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	matches, err := zglob.Glob(filepath.Join(archivePath, "**", "*.zip"))
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func parseArchiveDate(filename string) (time.Time, error) {
	lastIndex := strings.LastIndex(filename, "_")

	unixRaw, err := strconv.ParseInt(strings.TrimSuffix(filename[lastIndex+1:], ".zip"), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(unixRaw, 0), nil
}
