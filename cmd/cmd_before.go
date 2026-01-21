package cmd

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.String(flgPath) == "" {
		log.Fatal(fmt.Sprintf("Could not determine the current working directory. Please pass --%s.", flgPath))
	}

	err := createNonExistingFolder(cmd.String(flgPath))
	if err != nil {
		log.Fatal("Could not check/create the path.", "flag", flgPath, "filepath", cmd.String(flgPath), "error", err)
	}

	if cmd.String(flgServer) == "" {
		log.Fatal(fmt.Sprintf("Could not determine the current working server. Please pass --%s.", flgServer))
	}

	return ctx, nil
}
