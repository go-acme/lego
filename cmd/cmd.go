package cmd

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
	"gitlab.com/greyxor/slogor"
)

const rfc3339NanoNatural = "2006-01-02T15:04:05.000000000Z07:00"

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
		Flags:    CreateLogFlags(),
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
	}
}

func setUpLogger(cmd *cli.Command) {
	var logger *slog.Logger

	switch cmd.String(flgLogFormat) {
	case "json":
		opts := &slog.HandlerOptions{
			Level: getLogLeveler(cmd.String(flgLogLevel)),
		}

		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))

	case "text":
		opts := &slog.HandlerOptions{
			Level: getLogLeveler(cmd.String(flgLogLevel)),
		}

		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))

	default:
		logger = slog.New(slogor.NewHandler(os.Stderr,
			slogor.SetLevel(getLogLeveler(cmd.String(flgLogLevel))),
			slogor.SetTimeFormat(rfc3339NanoNatural)),
		)
	}

	log.SetDefault(logger)
}

func getLogLeveler(lvl string) slog.Leveler {
	switch strings.ToUpper(lvl) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
