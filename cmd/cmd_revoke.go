package cmd

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createRevoke() *cli.Command {
	return &cli.Command{
		Name:   "revoke",
		Usage:  "Revoke a certificate",
		Action: revoke,
		Flags:  flags.CreateRevokeFlags(),
	}
}

func revoke(ctx context.Context, cmd *cli.Command) error {
	keyType, err := certcrypto.ToKeyType(cmd.String(flags.FlgKeyType))
	if err != nil {
		return err
	}

	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		return fmt.Errorf("accounts storage initialization: %w", err)
	}

	account, err := accountsStorage.Get(ctx, keyType, cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	if account.Registration == nil {
		return fmt.Errorf("the account %s is not registered", account.GetID())
	}

	client, err := newClient(cmd, account)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flags.FlgPath))

	reason := cmd.Uint(flags.FlgReason)
	keep := cmd.Bool(flags.FlgKeep)

	for _, certID := range cmd.StringSlice(flags.FlgCertName) {
		err := revokeCertificate(ctx, client, certsStorage, certID, reason, keep)
		if err != nil {
			return err
		}
	}

	return nil
}

func revokeCertificate(ctx context.Context, client *lego.Client, certsStorage *storage.CertificatesStorage, certID string, reason uint, keep bool) error {
	log.Info("Trying to revoke the certificate.", log.CertNameAttr(certID))

	certBytes, err := certsStorage.ReadFile(certID, storage.ExtCert)
	if err != nil {
		return fmt.Errorf("certificate reading for domain %s: %w", certID, err)
	}

	err = client.Certificate.RevokeWithReason(ctx, certBytes, &reason)
	if err != nil {
		return fmt.Errorf("certificate revocation for domain %s: %w", certID, err)
	}

	log.Info("The certificate has been revoked.", log.CertNameAttr(certID))

	if keep {
		return nil
	}

	err = certsStorage.Archive(certID)
	if err != nil {
		return err
	}

	log.Info("The certificate has been archived.", log.CertNameAttr(certID))

	return nil
}
