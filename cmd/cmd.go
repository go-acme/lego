package cmd // import "github.com/xenolf/lego/cmd"

import "github.com/urfave/cli"

// CreateCommands Creates all CLI commands
func CreateCommands() []cli.Command {
	return []cli.Command{
		createRun(),
		createRevoke(),
		createRenew(),
		createDNSHelp(),
		createList(),
	}
}
