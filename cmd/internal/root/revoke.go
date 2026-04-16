package root

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func Revoke(ctx context.Context, cmd *cli.Command, cfg *configuration.Configuration) error {
	reason := cmd.Uint(flags.FlgReason)
	keep := cmd.Bool(flags.FlgKeep)

	store := storage.New(cfg.Storage)

	accountNodes := configuration.LookupCertificates(cfg, cmd.StringSlice(flags.FlgCertName))

	for _, accountNode := range accountNodes {
		account, err := store.Account.Get(accountNode.ServerConfig.URL, accountNode.KeyType, accountNode.Email, accountNode.ID)
		if err != nil {
			return fmt.Errorf("set up account: %w", err)
		}

		lazyClient := sync.OnceValues(func() (*lego.Client, error) {
			client, errC := lego.NewClient(newClientConfig(accountNode.ServerConfig, account, cfg.UserAgent))
			if errC != nil {
				return nil, errC
			}

			return client, nil
		})

		err = handleRegistration(ctx, lazyClient, accountNode.Account, store.Account, account, false)
		if err != nil {
			return fmt.Errorf("registration: %w", err)
		}

		client, err := lazyClient()
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}

		for _, cert := range accountNode.Children {
			err = revokeCertificate(ctx, client, store, cert.ID, reason, keep)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO(ldez): duplication of `cmd/cmd_certificates_revoke.go`. I need to think about that.
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
