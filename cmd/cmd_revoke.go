package cmd

import (
	"context"
	"log/slog"

	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createRevoke() *cli.Command {
	return &cli.Command{
		Name:   "revoke",
		Usage:  "Revoke a certificate",
		Action: revoke,
		Flags:  createRevokeFlags(),
	}
}

func revoke(ctx context.Context, cmd *cli.Command) error {
	account, keyType := setupAccount(ctx, cmd, newAccountsStorage(cmd))

	if account.Registration == nil {
		log.Fatal("Account is not registered. Use 'run' to register a new account.", slog.String("email", account.Email))
	}

	client := newClient(cmd, account, keyType)

	certsStorage := newCertificatesStorage(cmd)
	certsStorage.CreateRootFolder()

	for _, domain := range cmd.StringSlice(flgDomains) {
		log.Info("Trying to revoke the certificate.", log.DomainAttr(domain))

		certBytes, err := certsStorage.ReadFile(domain, storage.ExtCert)
		if err != nil {
			log.Fatal("Error while revoking the certificate.", log.DomainAttr(domain), log.ErrorAttr(err))
		}

		reason := cmd.Uint(flgReason)

		err = client.Certificate.RevokeWithReason(ctx, certBytes, &reason)
		if err != nil {
			log.Fatal("Error while revoking the certificate.", log.DomainAttr(domain), log.ErrorAttr(err))
		}

		log.Info("Certificate was revoked.", log.DomainAttr(domain))

		if cmd.Bool(flgKeep) {
			return nil
		}

		certsStorage.CreateArchiveFolder()

		err = certsStorage.MoveToArchive(domain)
		if err != nil {
			return err
		}

		log.Info("Certificate was archived", log.DomainAttr(domain))
	}

	return nil
}
