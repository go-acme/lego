package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sync"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createRun() *cli.Command {
	return &cli.Command{
		Name:   "run",
		Usage:  "Get or renew a certificate",
		Before: flags.RunFlagsValidation,
		Action: run,
		Flags:  flags.CreateRunFlags(),
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
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
		client, errC := newClient(cmd, account)
		if errC != nil {
			return nil, fmt.Errorf("new client: %w", errC)
		}

		errC = setupChallenges(cmd, client)
		if errC != nil {
			return nil, fmt.Errorf("setup challenges: %w", errC)
		}

		return client, nil
	})

	hookManager := newHookManager(cmd, store.Certificate, account)

	certID, err := getCertID(cmd)
	if err != nil {
		return err
	}

	resource, err := store.Certificate.ReadResource(certID)
	if err != nil {
		pe := new(fs.PathError)
		if !errors.As(err, &pe) {
			return fmt.Errorf("reading certificate resource file for %q: %w", certID, err)
		}
	}

	err = handleRegistration(ctx, cmd, lazyClient, store.Account, account, resource == nil)
	if err != nil {
		return fmt.Errorf("renew: registration: %w", err)
	}

	if resource == nil {
		// RUN
		err = obtain(ctx, cmd, certID, lazyClient, store.Certificate, hookManager)
		if err != nil {
			return fmt.Errorf("obtain certificate: %w", err)
		}

		return nil
	}

	log.Info("Renewing certificate", log.CertNameAttr(certID))

	// RENEW
	err = renew(ctx, cmd, certID, resource, lazyClient, store.Certificate, hookManager)
	if err != nil {
		return fmt.Errorf("renew certificate: %w", err)
	}

	return nil
}

func obtain(ctx context.Context, cmd *cli.Command, certID string, lazyClient lzSetUp, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	client, err := lazyClient()
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	if cmd.IsSet(flags.FlgCSR) {
		return obtainForCSR(ctx, cmd, client, certID, certsStorage, hookManager)
	}

	return obtainForDomains(ctx, cmd, client, certID, certsStorage, hookManager)
}

func renew(ctx context.Context, cmd *cli.Command, certID string, resource *storage.Certificate, lazyClient lzSetUp, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	if cmd.IsSet(flags.FlgCSR) {
		return renewForCSR(ctx, cmd, lazyClient, certID, certsStorage, hookManager)
	}

	domains := cmd.StringSlice(flags.FlgDomains)
	if len(domains) == 0 {
		domains = resource.Domains
	}

	return renewForDomains(ctx, cmd, lazyClient, certID, domains, certsStorage, hookManager)
}

func getCertID(cmd *cli.Command) (string, error) {
	domains := cmd.StringSlice(flags.FlgDomains)

	certID := cmd.String(flags.FlgCertName)

	switch {
	case certID != "":
		return certID, nil

	case cmd.IsSet(flags.FlgCSR):
		csr, err := storage.ReadCSRFile(cmd.String(flags.FlgCSR))
		if err != nil {
			return "", fmt.Errorf("could not read CSR file %q: %w", cmd.String(flags.FlgCSR), err)
		}

		return certcrypto.GetCSRMainDomain(csr)

	case len(domains) > 0:
		return domains[0], nil

	default:
		return "", errors.New("no domains, CSR, or certificate ID/name provided")
	}
}
