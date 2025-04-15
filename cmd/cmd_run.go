package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/registration"
	"github.com/urfave/cli/v2"
)

// Flag names.
const (
	flgNoBundle                       = "no-bundle"
	flgMustStaple                     = "must-staple"
	flgNotBefore                      = "not-before"
	flgNotAfter                       = "not-after"
	flgPrivateKey                     = "private-key"
	flgPreferredChain                 = "preferred-chain"
	flgProfile                        = "profile"
	flgAlwaysDeactivateAuthorizations = "always-deactivate-authorizations"
	flgRunHook                        = "run-hook"
	flgRunHookTimeout                 = "run-hook-timeout"
)

func createRun() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Register an account (if not registered), then create and install a certificate",
		Before: func(ctx *cli.Context) error {
			// we require either domains or csr, but not both
			hasDomains := len(ctx.StringSlice(flgDomains)) > 0
			hasCsr := ctx.String(flgCSR) != ""
			if hasDomains && hasCsr {
				log.Fatal("Please specify either --domains/-d or --csr/-c, but not both")
			}
			if !hasDomains && !hasCsr {
				log.Fatal("Please specify --domains/-d (or --csr/-c if you already have a CSR)")
			}
			return nil
		},
		Action: run,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  flgNoBundle,
				Usage: "Do not create a certificate bundle by adding the issuers certificate to the new certificate.",
			},
			&cli.BoolFlag{
				Name: flgMustStaple,
				Usage: "Include the OCSP must staple TLS extension in the CSR and generated certificate." +
					" Only works if the CSR is generated by lego.",
			},
			&cli.TimestampFlag{
				Name:   flgNotBefore,
				Usage:  "Set the notBefore field in the certificate (RFC3339 format)",
				Layout: time.RFC3339,
			},
			&cli.TimestampFlag{
				Name:   flgNotAfter,
				Usage:  "Set the notAfter field in the certificate (RFC3339 format)",
				Layout: time.RFC3339,
			},
			&cli.StringFlag{
				Name:  flgPrivateKey,
				Usage: "Path to private key (in PEM encoding) for the certificate. By default, the private key is generated.",
			},
			&cli.StringFlag{
				Name: flgPreferredChain,
				Usage: "If the CA offers multiple certificate chains, prefer the chain with an issuer matching this Subject Common Name." +
					" If no match, the default offered chain will be used.",
			},
			&cli.StringFlag{
				Name:  flgProfile,
				Usage: "If the CA offers multiple certificate profiles (draft-aaron-acme-profiles), choose this one.",
			},
			&cli.StringFlag{
				Name:  flgAlwaysDeactivateAuthorizations,
				Usage: "Force the authorizations to be relinquished even if the certificate request was successful.",
			},
			&cli.StringFlag{
				Name:  flgRunHook,
				Usage: "Define a hook. The hook is executed when the certificates are effectively created.",
			},
			&cli.DurationFlag{
				Name:  flgRunHookTimeout,
				Usage: "Define the timeout for the hook execution.",
				Value: 2 * time.Minute,
			},
		},
	}
}

const rootPathWarningMessage = `!!!! HEADS UP !!!!

Your account credentials have been saved in your
configuration directory at "%s".

You should make a secure backup of this folder now. This
configuration directory will also contain certificates and
private keys obtained from the ACME server so making regular
backups of this folder is ideal.
`

func run(ctx *cli.Context) error {
	accountsStorage := NewAccountsStorage(ctx)

	account, keyType := setupAccount(ctx, accountsStorage)

	client := setupClient(ctx, account, keyType)

	if account.Registration == nil {
		reg, err := register(ctx, client)
		if err != nil {
			log.Fatalf("Could not complete registration\n\t%v", err)
		}

		account.Registration = reg
		if err = accountsStorage.Save(account); err != nil {
			log.Fatal(err)
		}

		fmt.Printf(rootPathWarningMessage, accountsStorage.GetRootPath())
	}

	certsStorage := NewCertificatesStorage(ctx)
	certsStorage.CreateRootFolder()

	cert, err := obtainCertificate(ctx, client)
	if err != nil {
		// Make sure to return a non-zero exit code if ObtainSANCertificate returned at least one error.
		// Due to us not returning partial certificate we can just exit here instead of at the end.
		log.Fatalf("Could not obtain certificates:\n\t%v", err)
	}

	certsStorage.SaveResource(cert)

	meta := map[string]string{
		hookEnvAccountEmail: account.Email,
	}

	addPathToMetadata(meta, cert.Domain, cert, certsStorage)

	return launchHook(ctx.String(flgRunHook), ctx.Duration(flgRunHookTimeout), meta)
}

func handleTOS(ctx *cli.Context, client *lego.Client) bool {
	// Check for a global accept override
	if ctx.Bool(flgAcceptTOS) {
		return true
	}

	reader := bufio.NewReader(os.Stdin)
	log.Printf("Please review the TOS at %s", client.GetToSURL())

	for {
		fmt.Println("Do you accept the TOS? Y/n")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Could not read from console: %v", err)
		}

		text = strings.Trim(text, "\r\n")
		switch text {
		case "", "y", "Y":
			return true
		case "n", "N":
			return false
		default:
			fmt.Println("Your input was invalid. Please answer with one of Y/y, n/N or by pressing enter.")
		}
	}
}

func register(ctx *cli.Context, client *lego.Client) (*registration.Resource, error) {
	accepted := handleTOS(ctx, client)
	if !accepted {
		log.Fatal("You did not accept the TOS. Unable to proceed.")
	}

	if ctx.Bool(flgEAB) {
		kid := ctx.String(flgKID)
		hmacEncoded := ctx.String(flgHMAC)

		if kid == "" || hmacEncoded == "" {
			log.Fatalf("Requires arguments --%s and --%s.", flgKID, flgHMAC)
		}

		return client.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
			TermsOfServiceAgreed: accepted,
			Kid:                  kid,
			HmacEncoded:          hmacEncoded,
		})
	}

	return client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
}

func obtainCertificate(ctx *cli.Context, client *lego.Client) (*certificate.Resource, error) {
	bundle := !ctx.Bool(flgNoBundle)

	domains := ctx.StringSlice(flgDomains)
	if len(domains) > 0 {
		// obtain a certificate, generating a new private key
		request := certificate.ObtainRequest{
			Domains:                        domains,
			MustStaple:                     ctx.Bool(flgMustStaple),
			NotBefore:                      getTime(ctx, flgNotBefore),
			NotAfter:                       getTime(ctx, flgNotAfter),
			Bundle:                         bundle,
			PreferredChain:                 ctx.String(flgPreferredChain),
			Profile:                        ctx.String(flgProfile),
			AlwaysDeactivateAuthorizations: ctx.Bool(flgAlwaysDeactivateAuthorizations),
		}

		if ctx.IsSet(flgPrivateKey) {
			var err error
			request.PrivateKey, err = loadPrivateKey(ctx.String(flgPrivateKey))
			if err != nil {
				return nil, fmt.Errorf("load private key: %w", err)
			}
		}

		return client.Certificate.Obtain(request)
	}

	// read the CSR
	csr, err := readCSRFile(ctx.String(flgCSR))
	if err != nil {
		return nil, err
	}

	// obtain a certificate for this CSR
	request := certificate.ObtainForCSRRequest{
		CSR:                            csr,
		NotBefore:                      getTime(ctx, flgNotBefore),
		NotAfter:                       getTime(ctx, flgNotAfter),
		Bundle:                         bundle,
		PreferredChain:                 ctx.String(flgPreferredChain),
		Profile:                        ctx.String(flgProfile),
		AlwaysDeactivateAuthorizations: ctx.Bool(flgAlwaysDeactivateAuthorizations),
	}

	if ctx.IsSet(flgPrivateKey) {
		var err error
		request.PrivateKey, err = loadPrivateKey(ctx.String(flgPrivateKey))
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
	}

	return client.Certificate.ObtainForCSR(request)
}
