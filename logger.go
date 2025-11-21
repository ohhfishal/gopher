package main

import (
	"io"
	"log/slog"
)

// TODO: Extract to common repo

type LogConfig struct {
	Disable     bool       `help:"Disable logging. Shorthand for handler=discard."`
	HandlerType string     `name:"handler" enum:"json,discard,text" env:"HANDLER" default:"json" help:"Handler to use (${enum}) (env=$$${env})"`
	Level       slog.Level `default:"INFO" enum:"WARN,ERROR,INFO,DEBUG" help:"Set logging level (${enum})."`
	AddSource   bool       `default:"false" help:"Sets AddSource in the slog handler."`
	SetDefault  bool       `default:"true" help:"Set the global slog logger to use this config."`
}

func (config *LogConfig) AfterApply() error {
	if config.Disable {
		config.HandlerType = "discard"
	}
	return nil
}

func (config LogConfig) NewLogger(stdout io.Writer) *slog.Logger {
	logger := slog.New(config.Handler(stdout))
	if config.SetDefault {
		slog.SetDefault(logger)
	}
	return logger
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
