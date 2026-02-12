package cmd

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/urfave/cli/v3"
)

func createRun() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Register an account, then create and install a certificate",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// we require either domains or csr, but not both
			hasDomains := len(cmd.StringSlice(flgDomains)) > 0

			hasCsr := cmd.String(flgCSR) != ""
			if hasDomains && hasCsr {
				log.Fatal("Please specify either --domains/-d or --csr, but not both")
			}

			if !hasDomains && !hasCsr {
				log.Fatal("Please specify --domains/-d (or --csr if you already have a CSR)")
			}

			return ctx, validateNetworkStack(cmd)
		},
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

	client, err := newClient(cmd, account, keyType)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	setupChallenges(cmd, client)

	if account.Registration == nil {
		var reg *registration.Resource

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

	certRes, err := obtainCertificate(ctx, cmd, client)
	if err != nil {
		// Make sure to return a non-zero exit code if ObtainSANCertificate returned at least one error.
		// Due to us not returning partial certificate we can just exit here instead of at the end.
		return fmt.Errorf("obtain certificate: %w", err)
	}

	certID := cmd.String(flgCertName)
	if certID != "" {
		certRes.ID = certID
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flgPath))

	options := newSaveOptions(cmd)

	err = certsStorage.Save(certRes, options)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	meta := map[string]string{
		// TODO(ldez) add account ID.
		hook.EnvAccountEmail: account.Email,
	}

	hook.AddPathToMetadata(meta, certRes, certsStorage, options)

	return hook.Launch(ctx, cmd.String(flgDeployHook), cmd.Duration(flgDeployHookTimeout), meta)
}

func obtainCertificate(ctx context.Context, cmd *cli.Command, client *lego.Client) (*certificate.Resource, error) {
	domains := cmd.StringSlice(flgDomains)

	if len(domains) > 0 {
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
	csr, err := readCSRFile(cmd.String(flgCSR))
	if err != nil {
		return nil, err
	}

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
