package cmd

import (
	"github.com/urfave/cli/v3"
)

func createCertificates() *cli.Command {
	return &cli.Command{
		Name:  "certificates",
		Usage: "Certificates management.",
		Commands: []*cli.Command{
			createRevoke(),
			createListCertificates(),
		},
	}
}
