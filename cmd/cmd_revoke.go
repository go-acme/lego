package cmd

import (
	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v2"
)

func createRevoke() *cli.Command {
	return &cli.Command{
		Name:   "revoke",
		Usage:  "Revoke a certificate",
		Action: revoke,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "keep",
				Aliases: []string{"k"},
				Usage:   "Keep the certificates after the revocation instead of archiving them.",
			},
			&cli.UintFlag{
				Name: "reason",
				Usage: "Identifies the reason for the certificate revocation." +
					" See https://www.rfc-editor.org/rfc/rfc5280.html#section-5.3.1." +
					" Valid values are:" +
					" 0 (unspecified), 1 (keyCompromise), 2 (cACompromise), 3 (affiliationChanged)," +
					" 4 (superseded), 5 (cessationOfOperation), 6 (certificateHold), 8 (removeFromCRL)," +
					" 9 (privilegeWithdrawn), or 10 (aACompromise).",
				Value: acme.CRLReasonUnspecified,
			},
		},
	}
}

func revoke(ctx *cli.Context) error {
	acc, client := setup(ctx, NewAccountsStorage(ctx))

	if acc.Registration == nil {
		log.Fatalf("Account %s is not registered. Use 'run' to register a new account.\n", acc.Email)
	}

	certsStorage := NewCertificatesStorage(ctx)
	certsStorage.CreateRootFolder()

	for _, domain := range ctx.StringSlice("domains") {
		log.Printf("Trying to revoke certificate for domain %s", domain)

		certBytes, err := certsStorage.ReadFile(domain, ".crt")
		if err != nil {
			log.Fatalf("Error while revoking the certificate for domain %s\n\t%v", domain, err)
		}

		reason := ctx.Uint("reason")

		err = client.Certificate.RevokeWithReason(certBytes, &reason)
		if err != nil {
			log.Fatalf("Error while revoking the certificate for domain %s\n\t%v", domain, err)
		}

		log.Println("Certificate was revoked.")

		if ctx.Bool("keep") {
			return nil
		}

		certsStorage.CreateArchiveFolder()

		err = certsStorage.MoveToArchive(domain)
		if err != nil {
			return err
		}

		log.Println("Certificate was archived for domain:", domain)
	}

	return nil
}
