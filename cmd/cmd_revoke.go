package cmd

import (
	"context"
	"fmt"

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
		Flags:  createRevokeFlags(),
	}
}

func revoke(ctx context.Context, cmd *cli.Command) error {
	keyType, err := getKeyType(cmd.String(flgKeyType))
	if err != nil {
		return fmt.Errorf("get the key type: %w", err)
	}

	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		return fmt.Errorf("accounts storage initialization: %w", err)
	}

	account, err := accountsStorage.Get(ctx, keyType)
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	if account.Registration == nil {
		return fmt.Errorf("the account %s is not registered", account.Email)
	}

	client, err := newClient(cmd, account, keyType)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flgPath))

	err = certsStorage.CreateRootFolder()
	if err != nil {
		return fmt.Errorf("root folder creation: %w", err)
	}

	reason := cmd.Uint(flgReason)
	keep := cmd.Bool(flgKeep)

	for _, domain := range cmd.StringSlice(flgDomains) {
		err := revokeCertificate(ctx, client, certsStorage, domain, reason, keep)
		if err != nil {
			return err
		}
	}

	return nil
}

func revokeCertificate(ctx context.Context, client *lego.Client, certsStorage *storage.CertificatesStorage, domain string, reason uint, keep bool) error {
	log.Info("Trying to revoke the certificate.", log.DomainAttr(domain))

	certBytes, err := certsStorage.ReadFile(domain, storage.ExtCert)
	if err != nil {
		return fmt.Errorf("certificate reading for domain %s: %w", domain, err)
	}

	err = client.Certificate.RevokeWithReason(ctx, certBytes, &reason)
	if err != nil {
		return fmt.Errorf("certificate revocation for domain %s: %w", domain, err)
	}

	log.Info("The certificate has been revoked.", log.DomainAttr(domain))

	if keep {
		return nil
	}

	err = certsStorage.CreateArchiveFolder()
	if err != nil {
		return fmt.Errorf("archive folder creation: %w", err)
	}

	err = certsStorage.MoveToArchive(domain)
	if err != nil {
		return err
	}

	log.Info("The certificate has been archived.", log.DomainAttr(domain))

	return nil
}
