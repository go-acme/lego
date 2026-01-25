package cmd

import (
	"context"
	"crypto/x509"
	"fmt"

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
				log.Fatal("Please specify either --domains/-d or --csr/-c, but not both")
			}

			if !hasDomains && !hasCsr {
				log.Fatal("Please specify --domains/-d (or --csr/-c if you already have a CSR)")
			}

			return ctx, validateNetworkStack(cmd)
		},
		Action: run,
		Flags:  createRunFlags(),
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
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

	client, err := setupClient(cmd, account, keyType)
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	if account.Registration == nil {
		var reg *registration.Resource

		reg, err = registerAccount(ctx, cmd, client)
		if err != nil {
			return fmt.Errorf("could not complete registration: %w", err)
		}

		account.Registration = reg
		if err = accountsStorage.Save(account); err != nil {
			return fmt.Errorf("could not save the account file: %w", err)
		}

		fmt.Printf(rootPathWarningMessage, accountsStorage.GetRootPath())
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flgPath))

	err = certsStorage.CreateRootFolder()
	if err != nil {
		return fmt.Errorf("root folder creation: %w", err)
	}

	cert, err := obtainCertificate(ctx, cmd, client)
	if err != nil {
		// Make sure to return a non-zero exit code if ObtainSANCertificate returned at least one error.
		// Due to us not returning partial certificate we can just exit here instead of at the end.
		return fmt.Errorf("obtain certificate: %w", err)
	}

	options := newSaveOptions(cmd)

	err = certsStorage.SaveResource(cert, options)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	meta := map[string]string{
		hook.EnvAccountEmail: account.Email,
	}

	hook.AddPathToMetadata(meta, cert.Domain, cert, certsStorage, options)

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

			request.PrivateKey, err = storage.LoadPrivateKey(cmd.String(flgPrivateKey))
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

		request.PrivateKey, err = storage.LoadPrivateKey(cmd.String(flgPrivateKey))
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
	}

	return client.Certificate.ObtainForCSR(ctx, request)
}

func newObtainRequest(cmd *cli.Command, domains []string) certificate.ObtainRequest {
	return certificate.ObtainRequest{
		Domains:                        domains,
		MustStaple:                     cmd.Bool(flgMustStaple),
		NotBefore:                      cmd.Timestamp(flgNotBefore),
		NotAfter:                       cmd.Timestamp(flgNotAfter),
		Bundle:                         !cmd.Bool(flgNoBundle),
		PreferredChain:                 cmd.String(flgPreferredChain),
		Profile:                        cmd.String(flgProfile),
		AlwaysDeactivateAuthorizations: cmd.Bool(flgAlwaysDeactivateAuthorizations),
	}
}

func newObtainForCSRRequest(cmd *cli.Command, csr *x509.CertificateRequest) certificate.ObtainForCSRRequest {
	return certificate.ObtainForCSRRequest{
		CSR:                            csr,
		NotBefore:                      cmd.Timestamp(flgNotBefore),
		NotAfter:                       cmd.Timestamp(flgNotAfter),
		Bundle:                         !cmd.Bool(flgNoBundle),
		PreferredChain:                 cmd.String(flgPreferredChain),
		Profile:                        cmd.String(flgProfile),
		AlwaysDeactivateAuthorizations: cmd.Bool(flgAlwaysDeactivateAuthorizations),
	}
}

func validateNetworkStack(cmd *cli.Command) error {
	if cmd.Bool(flgIPv4Only) && cmd.Bool(flgIPv6Only) {
		return fmt.Errorf("cannot specify both --%s and --%s", flgIPv4Only, flgIPv6Only)
	}

	return nil
}
