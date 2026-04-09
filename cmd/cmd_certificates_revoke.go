package cmd

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/prompt"
	"github.com/go-acme/lego/v5/cmd/internal/root"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createRevoke() *cli.Command {
	return &cli.Command{
		Name:   "revoke",
		Usage:  "Revoke a certificate",
		Action: revokeFromConfig,
		Flags:  flags.CreateRevokeFlags(),
	}
}

func revokeFromConfig(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfiguration(cmd)
	if err == nil {
		if len(cmd.StringSlice(flags.FlgCertName)) == 0 && !prompt.Confirm("Are you sure you want to revoke all certificates defined in the configuration file?") {
			return nil
		}

		return root.Revoke(ctx, cmd, cfg)
	}

	nfErr := &configuration.FileNotFoundError{}
	if !errors.As(err, &nfErr) {
		return err
	}

	if len(cmd.StringSlice(flags.FlgCertName)) == 0 {
		return errors.New("no certificate names/IDs specified")
	}

	return revoke(ctx, cmd)
}

func revoke(ctx context.Context, cmd *cli.Command) error {
	keyType, err := certcrypto.ToKeyType(cmd.String(flags.FlgKeyType))
	if err != nil {
		return err
	}

	store := storage.New(cmd.String(flags.FlgPath))

	account, err := store.Account.Get(cmd.String(flags.FlgServer), keyType, cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	lazyClient := sync.OnceValues(func() (*lego.Client, error) {
		return newClient(cmd, account)
	})

	err = handleRegistration(ctx, cmd, lazyClient, store.Account, account, false)
	if err != nil {
		return fmt.Errorf("registration: %w", err)
	}

	client, err := lazyClient()
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	reason := cmd.Uint(flags.FlgReason)
	keep := cmd.Bool(flags.FlgKeep)

	for _, certID := range cmd.StringSlice(flags.FlgCertName) {
		err := revokeCertificate(ctx, client, store, certID, reason, keep)
		if err != nil {
			return err
		}
	}

	return nil
}

func revokeCertificate(ctx context.Context, client *lego.Client, store *storage.Storage, certID string, reason uint, keep bool) error {
	log.Info("Trying to revoke the certificate.", log.CertNameAttr(certID))

	certBytes, err := store.Certificate.ReadFile(certID, storage.ExtCert)
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

	err = store.Archiver.Certificate(certID)
	if err != nil {
		return err
	}

	log.Info("The certificate has been archived.", log.CertNameAttr(certID))

	return nil
}
