package cmd

import "github.com/urfave/cli/v2"

// CreateCommands Creates all CLI commands.
func CreateCommands() []*cli.Command {
	return []*cli.Command{
		createRegister(),
		createRun(),
		createRevoke(),
		createRenew(),
		createDNSHelp(),
		createList(),
	}
}
