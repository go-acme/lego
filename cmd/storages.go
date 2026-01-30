package cmd

import (
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/urfave/cli/v3"
)

func newAccountsStorageConfig(cmd *cli.Command) storage.AccountsStorageConfig {
	return storage.AccountsStorageConfig{
		BasePath:  cmd.String(flgPath),
		Server:    cmd.String(flgServer),
		UserAgent: getUserAgent(cmd),
	}
}

func newSaveOptions(cmd *cli.Command) *storage.SaveOptions {
	return &storage.SaveOptions{
		PEM:         cmd.Bool(flgPEM),
		PFX:         cmd.Bool(flgPFX),
		PFXFormat:   cmd.String(flgPFXPass),
		PFXPassword: cmd.String(flgPFXFormat),
	}
}
