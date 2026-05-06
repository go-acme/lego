package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/migrate"
	"github.com/go-acme/lego/v5/cmd/internal/prompt"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

const callToAction = `#######
#
# lego is an independent, free, and open-source project, if you value it, consider supporting it! ❤️
#
# https://donate.ldez.dev
#
#######

`

func createMigrate() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Migrate certificates and accounts.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			log.Warnf(log.LazySprintf("The migration will not work if the certificates have been generated with the '--filename' flag."+
				" Use the flag '--%s' to only migrate accounts.", flags.FlgAccountOnly))
			log.Warnf(log.LazySprintf("Please create a backup of %q before the migration.", cmd.String(flags.FlgPath)))

			if !prompt.Confirm("Continue?") {
				return nil
			}

			cfg := &configuration.Configuration{
				Accounts:     map[string]*configuration.Account{},
				Certificates: map[string]*configuration.Certificate{},
			}

			err := migrate.Accounts(cmd.String(flags.FlgPath), cfg)
			if err != nil {
				return err
			}

			if cmd.Bool(flags.FlgAccountOnly) {
				return createConfigurationFile(cfg)
			}

			err = migrate.Certificates(cmd.String(flags.FlgPath), cfg)
			if err != nil {
				return err
			}

			return createConfigurationFile(cfg)
		},
		Flags: flags.CreateMigrateFlags(),
	}
}

func createConfigurationFile(cfg *configuration.Configuration) error {
	date := strconv.FormatInt(time.Now().Unix(), 10)

	file, err := os.Create(fmt.Sprintf(".lego.migration.%s.yml", date))
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	filename, err := filepath.Abs(file.Name())
	if err != nil {
		filename = file.Name()
	}

	log.Debug("Creating the configuration file.", slog.String("filepath", filename))

	_, err = file.WriteString(callToAction)
	if err != nil {
		return err
	}

	err = yaml.NewEncoder(file).Encode(cfg)
	if err != nil {
		return fmt.Errorf("could not encode the configuration file: %w", err)
	}

	log.Warn("If you want to use the configuration file, please rename and review the file to handle the FIXME.", slog.String("filepath", filename))

	return nil
}
