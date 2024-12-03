package cmd

import (
	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v2"
)

// Flag names.
const (
	flgKeep   = "keep"
	flgReason = "reason"
)

func createRevoke() *cli.Command {
	return &cli.Command{
		Name:   "revoke",
		Usage:  "Revoke a certificate",
		Action: revoke,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    flgKeep,
				Aliases: []string{"k"},
				Usage:   "Keep the certificates after the revocation instead of archiving them.",
			},
			&cli.UintFlag{
				Name: flgReason,
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
	account, keyType := setupAccount(ctx, NewAccountsStorage(ctx))

	if account.Registration == nil {
		log.Fatalf("Account %s is not registered. Use 'run' to register a new account.\n", account.Email)
	}

	client := newClient(ctx, account, keyType)

	certsStorage := NewCertificatesStorage(ctx)
	certsStorage.CreateRootFolder()

	for _, domain := range ctx.StringSlice(flgDomains) {
		log.Printf("Trying to revoke certificate for domain %s", domain)

		certBytes, err := certsStorage.ReadFile(domain, certExt)
		if err != nil {
			log.Fatalf("Error while revoking the certificate for domain %s\n\t%v", domain, err)
		}

		reason := ctx.Uint(flgReason)

		err = client.Certificate.RevokeWithReason(certBytes, &reason)
		if err != nil {
			log.Fatalf("Error while revoking the certificate for domain %s\n\t%v", domain, err)
		}

		log.Println("Certificate was revoked.")

		if ctx.Bool(flgKeep) {
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
