package cmd

import (
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
)

type RunFunc func(context.Context, Gopher) error

func Register(name string, f RunFunc, description string) {
	targets[name] = targetEntry{f, description}
}

var targets = map[string]targetEntry{}

type targetEntry struct {
	Func        RunFunc
	Description string
}

type CMD struct {
	Target string `arg:"" default:"default" help:"Recipe to run."`
	List   bool   `short:"l" help:"List all targets then exit."`
	Go     Gopher `embed:"" group:"Golang Flags"`
}

func (config *CMD) Run(ctx context.Context, logger *slog.Logger) error {
	if config.List {
		for name, target := range targets {
			fmt.Printf("%s: %s\n", name, target.Description)
		}
		return nil
	}
	target, ok := targets[config.Target]
	if !ok {
		return fmt.Errorf("unknown target: %s", config.Target)
	}
	return target.Func(ctx, config.Go)
}

type Gopher struct {
	GoBin string `default:"go" help:"Go binary to use for commands."`
}

type RunArgs struct {
	GoBin  string
	Stdout io.Writer
}

type RunEvent iter.Seq[any]

type Runner interface {
	Run(context.Context, RunArgs) error
}

var _ Runner = &GoBuild{}

func (gopher *Gopher) Run(ctx context.Context, event RunEvent, runners ...Runner) error {
	for range event {
		for _, runner := range runners {
			if ctx.Err() != nil {
				return nil
			}
			if err := runner.Run(ctx, RunArgs{
				GoBin:  gopher.GoBin,
				Stdout: os.Stdout,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
