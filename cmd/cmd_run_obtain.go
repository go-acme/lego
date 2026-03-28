package cmd

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/urfave/cli/v3"
)

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

func obtainForDomains(ctx context.Context, cmd *cli.Command, client *lego.Client, certID string, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	domains := cmd.StringSlice(flags.FlgDomains)

	request, err := newObtainRequest(cmd, domains)
	if err != nil {
		return err
	}

	// TODO(ldez): factorize?
	if cmd.IsSet(flags.FlgPrivateKey) {
		request.PrivateKey, err = storage.ReadPrivateKeyFile(cmd.String(flags.FlgPrivateKey))
		if err != nil {
			return fmt.Errorf("load private key: %w", err)
		}
	}

	err = hookManager.PreForDomains(ctx, certID, request)
	if err != nil {
		return err
	}

	defer func() { _ = hookManager.Post(ctx) }()

	certRes, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		return err
	}

	if certID != "" {
		certRes.ID = certID
	}

	options := newSaveOptions(cmd)

	err = certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginCommand,
		},
		options,
	)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return hookManager.Deploy(ctx, certRes, options)
}

func obtainForCSR(ctx context.Context, cmd *cli.Command, client *lego.Client, certID string, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	// read the CSR
	csr, err := storage.ReadCSRFile(cmd.String(flags.FlgCSR))
	if err != nil {
		return err
	}

	// obtain a certificate for this CSR
	request := newObtainForCSRRequest(cmd, csr)

	// TODO(ldez): factorize?
	if cmd.IsSet(flags.FlgPrivateKey) {
		request.PrivateKey, err = storage.ReadPrivateKeyFile(cmd.String(flags.FlgPrivateKey))
		if err != nil {
			return fmt.Errorf("load private key: %w", err)
		}
	}

	err = hookManager.PreForCSR(ctx, certID, request)
	if err != nil {
		return err
	}

	defer func() { _ = hookManager.Post(ctx) }()

	certRes, err := client.Certificate.ObtainForCSR(ctx, request)
	if err != nil {
		return err
	}

	if certID != "" {
		certRes.ID = certID
	}

	options := newSaveOptions(cmd)

	err = certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginCommand,
		},
		options,
	)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return hookManager.Deploy(ctx, certRes, options)
}
