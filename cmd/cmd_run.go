package cmd

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/urfave/cli/v3"
)

func createRun() *cli.Command {
	return &cli.Command{
		Name:   "run",
		Usage:  "Register an account, then create and install a certificate",
		Before: runFlagsValidation,
		Action: run,
		Flags:  createRunFlags(),
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	keyType, err := certcrypto.GetKeyType(cmd.String(flgKeyType))
	if err != nil {
		return fmt.Errorf("get the key type: %w", err)
	}

	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		return fmt.Errorf("accounts storage initialization: %w", err)
	}

	account, err := accountsStorage.Get(ctx, keyType, cmd.String(flgEmail), cmd.String(flgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flgPath))

	hookManager := newHookManager(cmd, certsStorage, account)

	client, err := newClient(cmd, account, keyType)
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

		fmt.Printf(rootPathWarningMessage, accountsStorage.GetRootPath())
	}

	setupChallenges(cmd, client)

	certRes, err := obtainCertificate(ctx, cmd, client, hookManager)
	if err != nil {
		// Make sure to return a non-zero exit code if ObtainSANCertificate returned at least one error.
		// Due to us not returning partial certificate we can just exit here instead of at the end.
		return fmt.Errorf("obtain certificate: %w", err)
	}

	certID := cmd.String(flgCertName)
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
	domains := cmd.StringSlice(flgDomains)

	if len(domains) > 0 {
		err := hookManager.Pre(ctx, cmd.String(flgCertName), domains)
		if err != nil {
			return nil, err
		}

		defer func() { _ = hookManager.Post(ctx) }()

		// obtain a certificate, generating a new private key
		request := newObtainRequest(cmd, domains)

		// TODO(ldez): factorize?
		if cmd.IsSet(flgPrivateKey) {
			var err error

			request.PrivateKey, err = storage.ReadPrivateKeyFile(cmd.String(flgPrivateKey))
			if err != nil {
				return nil, fmt.Errorf("load private key: %w", err)
			}
		}

		return client.Certificate.Obtain(ctx, request)
	}

	// read the CSR
	csr, err := storage.ReadCSRFile(cmd.String(flgCSR))
	if err != nil {
		return nil, err
	}

	err = hookManager.Pre(ctx, cmd.String(flgCertName), certcrypto.ExtractDomainsCSR(csr))
	if err != nil {
		return nil, err
	}

	defer func() { _ = hookManager.Post(ctx) }()

	// obtain a certificate for this CSR
	request := newObtainForCSRRequest(cmd, csr)

	// TODO(ldez): factorize?
	if cmd.IsSet(flgPrivateKey) {
		var err error

		request.PrivateKey, err = storage.ReadPrivateKeyFile(cmd.String(flgPrivateKey))
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
	}

	return client.Certificate.ObtainForCSR(ctx, request)
}
