package cmd

import (
	"cmp"
	"log/slog"
	"os"
	"strings"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
	"gitlab.com/greyxor/slogor"
)

const rfc3339NanoNatural = "2006-01-02T15:04:05.000000000Z07:00"

func setUpLogger(cmd *cli.Command, logCfg *configuration.Log) {
	cfg := &configuration.Log{}

	if logCfg == nil {
		cfg.Level = cmd.String(flags.FlgLogLevel)
		cfg.Format = cmd.String(flags.FlgLogFormat)
	} else {
		cfg.Level = cmp.Or(logCfg.Level, cmd.String(flags.FlgLogLevel))
		cfg.Format = cmp.Or(logCfg.Format, cmd.String(flags.FlgLogFormat))
	}

	level := getLogLeveler(cfg.Level)

	var logger *slog.Logger

	switch cfg.Format {
	case configuration.LogFormatJSON:
		opts := &slog.HandlerOptions{
			Level: level,
		}

		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))

	case configuration.LogFormatText:
		opts := &slog.HandlerOptions{
			Level: level,
		}

		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))

	default:
		opts := []slogor.OptionFn{
			slogor.SetLevel(level),
			slogor.SetTimeFormat(rfc3339NanoNatural),
		}

		if !isatty.IsTerminal(os.Stdout.Fd()) {
			opts = append(opts, slogor.DisableColor())
		}

		logger = slog.New(slogor.NewHandler(os.Stdout, opts...))
	}

	log.SetDefault(logger)
}

func getLogLeveler(lvl string) slog.Leveler {
	switch strings.ToUpper(lvl) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
