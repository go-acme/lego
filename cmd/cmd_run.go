package cmd

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
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
		Usage:  "Register an account, then create and install a certificate",
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

	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		return fmt.Errorf("accounts storage initialization: %w", err)
	}

	account, err := accountsStorage.Get(ctx, keyType, cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flags.FlgPath))

	hookManager := newHookManager(cmd, certsStorage, account)

	client, err := newClient(cmd, account)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	if account.Registration == nil {
		var reg *acme.ExtendedAccount

		reg, err = registerAccount(ctx, cmd, client)
		if err != nil {
			return fmt.Errorf("could not complete registration: %w", err)
		}

		account.Registration = reg
		if err = accountsStorage.Save(keyType, account); err != nil {
			return fmt.Errorf("could not save the account file: %w", err)
		}

		log.Warnf(log.LazySprintf(storage.RootPathWarningMessage, accountsStorage.GetRootPath()))
	}

	err = setupChallenges(cmd, client)
	if err != nil {
		return fmt.Errorf("setup challenges: %w", err)
	}

	certRes, err := obtainCertificate(ctx, cmd, client, hookManager)
	if err != nil {
		// Make sure to return a non-zero exit code if ObtainSANCertificate returned at least one error.
		// Due to us not returning partial certificate we can just exit here instead of at the end.
		return fmt.Errorf("obtain certificate: %w", err)
	}

	certID := cmd.String(flags.FlgCertName)
	if certID != "" {
		certRes.ID = certID
	}

	options := newSaveOptions(cmd)

	err = certsStorage.Save(certRes, options)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return hookManager.Deploy(ctx, certRes, options)
}

func obtainCertificate(ctx context.Context, cmd *cli.Command, client *lego.Client, hookManager *hook.Manager) (*certificate.Resource, error) {
	domains := cmd.StringSlice(flags.FlgDomains)

	if len(domains) > 0 {
		return obtainForDomains(ctx, cmd, client, hookManager)
	}

	return obtainForCSR(ctx, cmd, client, hookManager)
}

func obtainForDomains(ctx context.Context, cmd *cli.Command, client *lego.Client, hookManager *hook.Manager) (*certificate.Resource, error) {
	domains := cmd.StringSlice(flags.FlgDomains)

	err := hookManager.Pre(ctx, cmd.String(flags.FlgCertName), domains)
	if err != nil {
		return nil, err
	}

	defer func() { _ = hookManager.Post(ctx) }()

	request, err := newObtainRequest(cmd, domains)
	if err != nil {
		return nil, err
	}

	// TODO(ldez): factorize?
	if cmd.IsSet(flags.FlgPrivateKey) {
		request.PrivateKey, err = storage.ReadPrivateKeyFile(cmd.String(flags.FlgPrivateKey))
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
	}

	return client.Certificate.Obtain(ctx, request)
}

func obtainForCSR(ctx context.Context, cmd *cli.Command, client *lego.Client, hookManager *hook.Manager) (*certificate.Resource, error) {
	// read the CSR
	csr, err := storage.ReadCSRFile(cmd.String(flags.FlgCSR))
	if err != nil {
		return nil, err
	}

	err = hookManager.Pre(ctx, cmd.String(flags.FlgCertName), certcrypto.ExtractDomainsCSR(csr))
	if err != nil {
		return nil, err
	}

	defer func() { _ = hookManager.Post(ctx) }()

	// obtain a certificate for this CSR
	request := newObtainForCSRRequest(cmd, csr)

	// TODO(ldez): factorize?
	if cmd.IsSet(flags.FlgPrivateKey) {
		request.PrivateKey, err = storage.ReadPrivateKeyFile(cmd.String(flags.FlgPrivateKey))
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
	}

	return client.Certificate.ObtainForCSR(ctx, request)
}
