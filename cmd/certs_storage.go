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

	config := storage.CertificatesWriterConfig{
		BasePath:    basePath,
		PEM:         cmd.Bool(flgPEM),
		PFX:         cmd.Bool(flgPFX),
		PFXFormat:   cmd.String(flgPFXPass),
		PFXPassword: cmd.String(flgPFXFormat),
		Filename:    cmd.String(flgFilename),
	}

	writer, err := storage.NewCertificatesWriter(config)
	if err != nil {
		log.Fatal("Certificates storage initialization", log.ErrorAttr(err))
	}

	return &CertificatesStorage{
		CertificatesWriter: writer,
		CertificatesReader: storage.NewCertificatesReader(basePath),
	}
}
