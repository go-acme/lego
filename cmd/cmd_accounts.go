package cmd

import (
	"github.com/urfave/cli/v3"
)

func createAccounts() *cli.Command {
	return &cli.Command{
		Name:  "accounts",
		Usage: "Accounts management.",
		Commands: []*cli.Command{
			createRegister(),
			createListAccounts(),
		},
	}
}
