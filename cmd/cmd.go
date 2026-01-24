package cmd

import "github.com/urfave/cli/v3"

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
