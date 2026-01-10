package cmd

import (
	"fmt"

	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v2"
)

func Before(ctx *cli.Context) error {
	if ctx.String(flgPath) == "" {
		log.Fatal(fmt.Sprintf("Could not determine the current working directory. Please pass --%s.", flgPath))
	}

	err := createNonExistingFolder(ctx.String(flgPath))
	if err != nil {
		log.Fatal("Could not check/create the path.", "flag", flgPath, "filepath", ctx.String(flgPath), "error", err)
	}

	if ctx.String(flgServer) == "" {
		log.Fatal(fmt.Sprintf("Could not determine the current working server. Please pass --%s.", flgServer))
	}

	return nil
}
