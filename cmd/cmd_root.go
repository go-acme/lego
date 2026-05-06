package cmd

import (
	"context"
	"errors"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/root"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/urfave/cli/v3"
)

// CreateRootCommand Creates the root CLI command.
func CreateRootCommand() *cli.Command {
	return &cli.Command{
		Name:                  "lego",
		Usage:                 "ACME client written in Go",
		EnableShellCompletion: true,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			setUpLogger(cmd, nil)

			if cmd.NArg() > 0 && cmd.Command(cmd.Args().First()) == nil {
				return ctx, errors.New("unknown command")
			}

			return ctx, nil
		},
		Action:   rootRun,
		Flags:    flags.CreateRootFlags(),
		Commands: createCommands(),
	}
}

// createCommands Creates all CLI commands.
func createCommands() []*cli.Command {
	return []*cli.Command{
		createRun(),
		createCertificates(),
		createAccounts(),
		createArchives(),
		createDNSHelp(),
		createMigrate(),
	}
}

func rootRun(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfiguration(cmd)
	if err != nil {
		return err
	}

	err = root.Process(ctx, cfg)
	if err != nil {
		return err
	}

	store := storage.NewConfigurationStorage(cfg.Storage)

	return store.Backup(cfg)
}

func loadConfiguration(cmd *cli.Command) (*configuration.Configuration, error) {
	filename, err := getConfigurationPath(cmd)
	if err != nil {
		return nil, err
	}

	cfg, err := configuration.ReadConfiguration(filename)
	if err != nil {
		return nil, err
	}

	setUpLogger(cmd, cfg.Log)

	configuration.ApplyDefaults(cfg)

	err = configuration.Validate(cfg)
	if err != nil {
		return nil, err
	}

	// Set effective User Agent.
	cfg.UserAgent = getUserAgent(cmd, cfg.UserAgent)

	return cfg, nil
}

func getConfigurationPath(cmd *cli.Command) (string, error) {
	configPath := cmd.String(flags.FlgConfig)

	if configPath != "" {
		return configPath, nil
	}

	return configuration.FindDefaultConfigurationFile()
}
