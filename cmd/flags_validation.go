package cmd

import (
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

// FIXME rename + remove fatal?
func validateFlags(cmd *cli.Command) {
	err := createNonExistingFolder(cmd.String(flgPath))
	if err != nil {
		log.Fatal("Could not check/create the path.",
			slog.String("flag", flgPath),
			slog.String("filepath", cmd.String(flgPath)),
			log.ErrorAttr(err),
		)
	}

	if cmd.String(flgServer) == "" {
		log.Fatal(fmt.Sprintf("Could not determine the current working server. Please pass --%s.", flgServer))
	}
}
