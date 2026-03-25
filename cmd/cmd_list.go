package cmd

import (
	"github.com/urfave/cli/v3"
)

func createList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Display information about certificates.",
		Commands: []*cli.Command{
			createListCertificates(),
		},
	}
}
