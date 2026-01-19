package cmd

import (
	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/log"
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

func revoke(cliCtx *cli.Context) error {
	account, keyType := setupAccount(cliCtx, NewAccountsStorage(cliCtx))

	if account.Registration == nil {
		log.Fatal("Account is not registered. Use 'run' to register a new account.", "email", account.Email)
	}

	client := newClient(cliCtx, account, keyType)

	certsStorage := NewCertificatesStorage(cliCtx)
	certsStorage.CreateRootFolder()

	for _, domain := range cliCtx.StringSlice(flgDomains) {
		log.Info("Trying to revoke the certificate.", "domain", domain)

		certBytes, err := certsStorage.ReadFile(domain, certExt)
		if err != nil {
			log.Fatal("Error while revoking the certificate.", "domain", domain, "error", err)
		}

		reason := cliCtx.Uint(flgReason)

		err = client.Certificate.RevokeWithReason(cliCtx.Context, certBytes, &reason)
		if err != nil {
			log.Fatal("Error while revoking the certificate.", "domain", domain, "error", err)
		}

		log.Info("Certificate was revoked.", "domain", domain)

		if cliCtx.Bool(flgKeep) {
			return nil
		}

		certsStorage.CreateArchiveFolder()

		err = certsStorage.MoveToArchive(domain)
		if err != nil {
			return err
		}

		log.Info("Certificate was archived", "domain", domain)
	}

	return nil
}
