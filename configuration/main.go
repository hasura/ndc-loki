package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/hasura/ndc-loki/configuration/version"
	"github.com/lmittmann/tint"
)

var cli struct {
	LogLevel string          `help:"Log level." enum:"debug,info,warn,error,DEBUG,INFO,WARN,ERROR" env:"HASURA_PLUGIN_LOG_LEVEL" default:"info"`
	Update   UpdateArguments `cmd:"" help:"Introspect metric metadata and update configuration."`
	Version  struct{}        `cmd:"" help:"Print the CLI version."`
}

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.TODO(), os.Interrupt)

	cmd := kong.Parse(&cli, kong.UsageOnError())
	logger, err := initLogger(cli.LogLevel)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize: %s", err))
		os.Exit(1)
	}

	switch cmd.Command() {
	case "update":
		if err := introspectSchema(ctx, &cli.Update); err != nil {
			logger.Error(fmt.Sprintf("failed to update configuration: %s", err))
			os.Exit(1)
		}
	case "version":
		_, _ = fmt.Fprint(os.Stderr, version.BuildVersion)
	default:
		logger.Error(fmt.Sprintf("unknown command <%s>", cmd.Command()))
		os.Exit(1)
	}

	stop()
}

func initLogger(logLevel string) (*slog.Logger, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(strings.ToUpper(logLevel)))
	if err != nil {
		return nil, err
	}

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: "15:04",
	}))
	slog.SetDefault(logger)

	return logger, nil
}
