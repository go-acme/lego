package cmd

import (
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

// CertificatesStorage a certificates' storage.
type CertificatesStorage struct {
	*storage.CertificatesWriter
	*storage.CertificatesReader
}

// newCertificatesStorage create a new certificates storage.
func newCertificatesStorage(cmd *cli.Command) *CertificatesStorage {
	basePath := cmd.String(flgPath)

	writer, err := storage.NewCertificatesWriter(newCertificatesWriterConfig(cmd, basePath))
	if err != nil {
		log.Fatal("Certificates storage initialization", log.ErrorAttr(err))
	}

	return &CertificatesStorage{
		CertificatesWriter: writer,
		CertificatesReader: storage.NewCertificatesReader(basePath),
	}
}

func newCertificatesWriterConfig(cmd *cli.Command, basePath string) storage.CertificatesWriterConfig {
	return storage.CertificatesWriterConfig{
		BasePath:    basePath,
		PEM:         cmd.Bool(flgPEM),
		PFX:         cmd.Bool(flgPFX),
		PFXFormat:   cmd.String(flgPFXPass),
		PFXPassword: cmd.String(flgPFXFormat),
		Filename:    cmd.String(flgFilename),
	}
}

// newAccountsStorage Creates a new AccountsStorage.
func newAccountsStorage(cmd *cli.Command) *storage.AccountsStorage {
	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		log.Fatal("Accounts storage initialization", log.ErrorAttr(err))
	}

	return accountsStorage
}

func newAccountsStorageConfig(cmd *cli.Command) storage.AccountsStorageConfig {
	return storage.AccountsStorageConfig{
		Email:     cmd.String(flgEmail),
		BasePath:  cmd.String(flgPath),
		Server:    cmd.String(flgServer),
		UserAgent: getUserAgent(cmd),
	}
}
