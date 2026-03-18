package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/migrate"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createMigrate() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Migrate certificates and accounts.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if !confirmMigration(cmd) {
				return nil
			}

			err := migrate.Accounts(cmd.String(flags.FlgPath))
			if err != nil {
				return err
			}

			if cmd.Bool(flags.FlgAccountOnly) {
				return nil
			}

			return migrate.Certificates(cmd.String(flags.FlgPath))
		},
		Flags: flags.CreateMigrateFlags(),
	}
}

func confirmMigration(cmd *cli.Command) bool {
	reader := bufio.NewReader(os.Stdin)

	log.Warnf(log.LazySprintf("The migration will not work if the certificates have been generated with the '--filename' flag."+
		" Use the flag '--%s' to only migrate accounts.", flags.FlgAccountOnly))
	log.Warnf(log.LazySprintf("Please create a backup of %q before the migration.", cmd.String(flags.FlgPath)))

	for {
		fmt.Println("Continue? Y/n")

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Could not read from the console", log.ErrorAttr(err))
		}

		text = strings.Trim(text, "\r\n")
		switch strings.ToUpper(text) {
		case "", "Y":
			return true
		case "N":
			return false
		default:
			log.Warn("Your input was invalid. Please answer with one of Y/y, n/N or by pressing enter.")
		}
	}
}
