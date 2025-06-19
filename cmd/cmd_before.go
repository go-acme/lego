package cmd

import (
	"context"

	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.String(flgPath) == "" {
		log.Fatalf("Could not determine current working directory. Please pass --%s.", flgPath)
	}

	err := createNonExistingFolder(cmd.String(flgPath))
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	if cmd.String(flgServer) == "" {
		log.Fatalf("Could not determine current working server. Please pass --%s.", flgServer)
	}

	return ctx, nil
}
