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
		Name:   "migrate",
		Usage:  "Migrate certificates and accounts.",
		Action: migration,
		Flags:  flags.CreateMigrateFlags(),
	}
}

func migration(_ context.Context, cmd *cli.Command) error {
	root := cmd.String(flags.FlgPath)

	log.Warnf(log.LazySprintf("The migration will not work if the certificates have been generated with the '--filename' flag."+
		" Use the flag '--%s' to only migrate accounts.", flags.FlgAccountOnly))
	log.Warnf(log.LazySprintf("Please create a backup of %q before the migration.", root))

	if !prompt.Confirm("Continue?") {
		return nil
	}

	cfg := &configuration.Configuration{
		Accounts:     map[string]*configuration.Account{},
		Certificates: map[string]*configuration.Certificate{},
	}

	err := migrate.Accounts(root, cfg)
	if err != nil {
		return err
	}

	if cmd.Bool(flags.FlgAccountOnly) {
		return createConfigurationFile(root, cfg)
	}

	err = migrate.Certificates(root, cfg)
	if err != nil {
		return err
	}

	return createConfigurationFile(root, cfg)
}

func createConfigurationFile(root string, cfg *configuration.Configuration) error {
	if len(cfg.Accounts) == 0 && len(cfg.Certificates) == 0 {
		return nil
	}

	wd, err := os.Getwd()
	if err == nil {
		if filepath.Join(wd, ".lego") != root {
			cfg.Storage = "FIXME: " + root
		}
	}

	date := strconv.FormatInt(time.Now().Unix(), 10)

	file, err := os.Create(fmt.Sprintf(".lego.migration.%s.yml", date))
	if err != nil {
		log.Debug("The suggested configuration file cannot be created.", log.ErrorAttr(err))

		return suggestedConfigurationFallback(cfg)
	}

	defer func() { _ = file.Close() }()

	filename, err := filepath.Abs(file.Name())
	if err != nil {
		filename = file.Name()
	}

	log.Debug("Creating the configuration file.", slog.String("filepath", filename))

	err = createSuggestedConfiguration(file, cfg)
	if err != nil {
		return err
	}

	log.Warn("If you want to use the configuration file, please rename and review the file to handle the FIXME.", slog.String("filepath", filename))

	return nil
}

func createSuggestedConfiguration(file *os.File, cfg *configuration.Configuration) error {
	_, err := file.WriteString(callToAction)
	if err != nil {
		return err
	}

	err = yaml.NewEncoder(file).Encode(cfg)
	if err != nil {
		return fmt.Errorf("could not encode the configuration file: %w", err)
	}

	return nil
}

func suggestedConfigurationFallback(cfg *configuration.Configuration) error {
	log.Info("Suggested configuration file content.")

	err := createSuggestedConfiguration(os.Stdout, cfg)
	if err != nil {
		return err
	}

	log.Warn("If you want to use the configuration file, please review the content to handle the FIXME and save it to a `.lego.yml` file.")

	return nil
}
