package cmd

import (
	"context"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/urfave/cli/v3"
)

// CreateRootCommand Creates the root CLI command.
func CreateRootCommand() *cli.Command {
	return &cli.Command{
		Name:                  "lego",
		Usage:                 "ACME client written in Go",
		EnableShellCompletion: true,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			setUpLogger(cmd)

			return ctx, nil
		},
		Flags:    flags.CreateRootFlags(),
		Commands: CreateCommands(),
	}
}

// CreateCommands Creates all CLI commands.
func CreateCommands() []*cli.Command {
	return []*cli.Command{
		createRun(),
		createRevoke(),
		createRenew(),
		createRegister(),
		createDNSHelp(),
		createList(),
		createMigrate(),
	}
}
