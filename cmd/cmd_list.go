package cmd

import (
	"github.com/urfave/cli/v3"
)

func createList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Display certificates and accounts information.",
		Commands: []*cli.Command{
			createListCertificates(),
			createListAccounts(),
		},
	}
}
