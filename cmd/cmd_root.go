package cmd

import (
	"context"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/root"
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

			return ctx, nil
		},
		Action:   rootRun,
		Flags:    flags.CreateRootFlags(),
		Commands: CreateCommands(),
	}
}

// CreateCommands Creates all CLI commands.
func CreateCommands() []*cli.Command {
	return []*cli.Command{
		createRun(),
		createRevoke(),
		createAccounts(),
		createDNSHelp(),
		createList(),
		createMigrate(),
	}
}

func rootRun(ctx context.Context, cmd *cli.Command) error {
	filename, err := getConfigurationPath(cmd)
	if err != nil {
		return err
	}

	cfg, err := configuration.ReadConfiguration(filename)
	if err != nil {
		return err
	}

	setUpLogger(cmd, cfg.Log)

	configuration.ApplyDefaults(cfg)

	err = configuration.Validate(cfg)
	if err != nil {
		return err
	}

	// Set effective User Agent.
	cfg.UserAgent = getUserAgent(cmd, cfg.UserAgent)

	return root.Process(ctx, cfg)
}

func getConfigurationPath(cmd *cli.Command) (string, error) {
	configPath := cmd.String(flags.FlgConfig)

	if configPath != "" {
		return configPath, nil
	}

	return configuration.FindDefaultConfigurationFile()
}
