package cmd

import (
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

// FIXME convert to flag requirement?
func validatePathFlag(cmd *cli.Command) {
	if cmd.String(flgPath) == "" {
		log.Fatal(fmt.Sprintf("Could not determine the current working directory. Please pass --%s.", flgPath))
	}

	// FIXME is command list need an existing path?
}

// FIXME rename + remove fatal?
func validateFlags(cmd *cli.Command) {
	validatePathFlag(cmd)

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
