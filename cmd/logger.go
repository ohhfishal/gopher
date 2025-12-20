package cmd

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// TODO: Extract to common repo

type LogConfig struct {
	Disable     bool       `help:"Disable logging. Shorthand for handler=discard."`
	Directory   string     `kong:"-"`
	File        string     `default:"gopher.log" help:"File to write logs to inside of --gopher-dir. Use \"-\" to write logs to stdout."`
	HandlerType string     `name:"handler" enum:"json,discard,text" env:"HANDLER" default:"json" help:"Handler to use (${enum}) (env=$$${env})"`
	Level       slog.Level `default:"INFO" enum:"WARN,ERROR,INFO,DEBUG" help:"Set logging level (${enum})."`
	AddSource   bool       `default:"false" help:"Sets AddSource in the slog handler."`
	SetDefault  bool       `default:"true" help:"Set the global slog logger to use this config."`
}

func (config LogConfig) NewLogger(stdout io.Writer) (*slog.Logger, error) {
	if config.Disable {
		return slog.New(slog.DiscardHandler), nil
	}

	var file io.Writer = stdout
	var err error
	if config.File != "-" {
		destination := filepath.Join(config.Directory, config.File)
		file, err = os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
	}
	logger := slog.New(config.Handler(file))
	if config.SetDefault {
		slog.SetDefault(logger)
	}
	return logger, nil
}

func (config LogConfig) Handler(stdout io.Writer) slog.Handler {
	switch config.HandlerType {
	case "discard":
		return slog.DiscardHandler
	case "json":
		return slog.NewJSONHandler(stdout, &slog.HandlerOptions{
			AddSource: config.AddSource,
			Level:     config.Level,
		})
	case "text":
		fallthrough
	default:
		return slog.NewTextHandler(stdout, &slog.HandlerOptions{
			AddSource: config.AddSource,
			Level:     config.Level,
		})
	}
}
