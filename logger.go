package main

import (
	"io"
	"log/slog"
)

// TODO: Extract to common repo

type LogConfig struct {
	Disable bool `help:"Disable logging. Shorthand for handler=discard."`
	HandlerType string `name:"handler" enum:"json,discard,text" env:"HANDLER" default:"json" help:"Handler to use (${enum}) (env=$$${env})"`
	Level slog.Level `default:"debug"`
	AddSource bool `default:"false"`
}

func (config *LogConfig) AfterApply() error {
	if config.Disable {
		config.HandlerType = "discard"
	}
	return nil
}

func (config LogConfig) NewLogger(stdout io.Writer) *slog.Logger {
	return slog.New(config.Handler(stdout))
}

func (config LogConfig) Handler(stdout io.Writer) slog.Handler {
	switch config.HandlerType {
		case "discard":
			return slog.DiscardHandler
		case "json":
			return slog.NewJSONHandler(stdout, &slog.HandlerOptions {
				AddSource: config.AddSource,
				Level: config.Level,
			})
		case "text":
			fallthrough
		default:
			return slog.NewTextHandler(stdout, &slog.HandlerOptions {
				AddSource: config.AddSource,
				Level: config.Level,
			})
	}
}
