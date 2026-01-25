package cmd

import "github.com/urfave/cli/v3"

// CreateRootCommand Creates the root CLI command.
func CreateRootCommand() *cli.Command {
	return &cli.Command{
		Name:                  "lego",
		Usage:                 "ACME client written in Go",
		EnableShellCompletion: true,
		Commands:              CreateCommands(),
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
	}
}
