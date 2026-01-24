package cmd

import (
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

// newCertificatesStorage create a new certificates storage.
func newCertificatesStorage(cmd *cli.Command) *storage.CertificatesStorage {
	certsStorage, err := storage.NewCertificatesStorage(newCertificatesWriterConfig(cmd))
	if err != nil {
		log.Fatal("Certificates storage", log.ErrorAttr(err))
	}

	return certsStorage
}

func newCertificatesWriterConfig(cmd *cli.Command) storage.CertificatesWriterConfig {
	return storage.CertificatesWriterConfig{
		BasePath:    cmd.String(flgPath),
		PEM:         cmd.Bool(flgPEM),
		PFX:         cmd.Bool(flgPFX),
		PFXFormat:   cmd.String(flgPFXPass),
		PFXPassword: cmd.String(flgPFXFormat),
	}
}

func newAccountsStorageConfig(cmd *cli.Command) storage.AccountsStorageConfig {
	return storage.AccountsStorageConfig{
		Email:     cmd.String(flgEmail),
		BasePath:  cmd.String(flgPath),
		Server:    cmd.String(flgServer),
		UserAgent: getUserAgent(cmd),
	}
}
