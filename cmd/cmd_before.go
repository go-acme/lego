package cmd

import (
	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v2"
)

func Before(ctx *cli.Context) error {
	if ctx.String(flgPath) == "" {
		log.Fatalf("Could not determine current working directory. Please pass --%s.", flgPath)
	}

	err := createNonExistingFolder(ctx.String(flgPath))
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	if ctx.String(flgServer) == "" {
		log.Fatalf("Could not determine current working server. Please pass --%s.", flgServer)
	}

	return nil
}
