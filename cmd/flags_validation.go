package cmd

import (
	"fmt"

	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

// FIXME rename + remove fatal?
func validateFlags(cmd *cli.Command) {
	if cmd.String(flgServer) == "" {
		log.Fatal(fmt.Sprintf("Could not determine the current working server. Please pass --%s.", flgServer))
	}
}
