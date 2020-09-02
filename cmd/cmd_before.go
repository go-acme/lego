package cmd

import (
	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli"
)

func Before(ctx *cli.Context) error {
	if len(ctx.GlobalString("path")) == 0 {
		log.Fatal("Could not determine current working directory. Please pass --path.")
	}

	err := createNonExistingFolder(ctx.GlobalString("path"))
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	if len(ctx.GlobalString("server")) == 0 {
		log.Fatal("Could not determine current working server. Please pass --server.")
	}

	return nil
}
