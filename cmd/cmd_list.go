package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/urfave/cli/v2"
)

func createList() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "Display certificates and accounts information.",
		Action: list,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "accounts",
				Aliases: []string{"a"},
				Usage:   "Display accounts.",
			},
			&cli.BoolFlag{
				Name:    "names",
				Aliases: []string{"n"},
				Usage:   "Display certificate common names only.",
			},
			// fake email, needed by NewAccountsStorage
			&cli.StringFlag{
				Name:   "email",
				Value:  "unknown",
				Hidden: true,
			},
		},
	}
}

func list(ctx *cli.Context) error {
	if ctx.Bool("accounts") && !ctx.Bool("names") {
		if err := listAccount(ctx); err != nil {
			return err
		}
	}

	return listCertificates(ctx)
}

func listCertificates(ctx *cli.Context) error {
	certsStorage := NewCertificatesStorage(ctx)

	matches, err := filepath.Glob(filepath.Join(certsStorage.GetRootPath(), "*.crt"))
	if err != nil {
		return err
	}

	names := ctx.Bool("names")

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
		if strings.HasSuffix(filename, ".issuer.crt") {
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

func listAccount(ctx *cli.Context) error {
	accountsStorage := NewAccountsStorage(ctx)

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
