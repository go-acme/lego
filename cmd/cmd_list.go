package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/urfave/cli/v3"
)

const (
	flgAccounts = "accounts"
	flgNames    = "names"
)

func createList() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "Display certificates and accounts information.",
		Action: list,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    flgAccounts,
				Aliases: []string{"a"},
				Usage:   "Display accounts.",
			},
			&cli.BoolFlag{
				Name:    flgNames,
				Aliases: []string{"n"},
				Usage:   "Display certificate common names only.",
			},
			// fake email, needed by NewAccountsStorage
			&cli.StringFlag{
				Name:   flgEmail,
				Value:  "unknown",
				Hidden: true,
			},
		},
	}
}

func list(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool(flgAccounts) && !cmd.Bool(flgNames) {
		if err := listAccount(ctx, cmd); err != nil {
			return err
		}
	}

	return listCertificates(ctx, cmd)
}

func listCertificates(_ context.Context, cmd *cli.Command) error {
	certsStorage := NewCertificatesStorage(cmd)

	matches, err := filepath.Glob(filepath.Join(certsStorage.GetRootPath(), "*.crt"))
	if err != nil {
		return err
	}

	names := cmd.Bool(flgNames)

	if len(matches) == 0 {
		if !names {
			fmt.Println("No certificates found.")
		}
		return nil
	}

	if !names {
		fmt.Println("Found the following certs:")
	}

	for _, filename := range matches {
		if strings.HasSuffix(filename, issuerExt) {
			continue
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		pCert, err := certcrypto.ParsePEMCertificate(data)
		if err != nil {
			return err
		}

		name, err := certcrypto.GetCertificateMainDomain(pCert)
		if err != nil {
			return err
		}

		if names {
			fmt.Println(name)
		} else {
			fmt.Println("  Certificate Name:", name)
			fmt.Println("    Domains:", strings.Join(pCert.DNSNames, ", "))
			fmt.Println("    Expiry Date:", pCert.NotAfter)
			fmt.Println("    Certificate Path:", filename)
			fmt.Println()
		}
	}

	return nil
}

func listAccount(_ context.Context, cmd *cli.Command) error {
	accountsStorage := NewAccountsStorage(cmd)

	matches, err := filepath.Glob(filepath.Join(accountsStorage.GetRootPath(), "*", "*", "*.json"))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Println("No accounts found.")
		return nil
	}

	fmt.Println("Found the following accounts:")
	for _, filename := range matches {
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		var account Account
		err = json.Unmarshal(data, &account)
		if err != nil {
			return err
		}

		uri, err := url.Parse(account.Registration.URI)
		if err != nil {
			return err
		}

		fmt.Println("  Email:", account.Email)
		fmt.Println("  Server:", uri.Host)
		fmt.Println("  Path:", filepath.Dir(filename))
		fmt.Println()
	}

	return nil
}
