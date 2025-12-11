package runner

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
)

var ErrOK = errors.New("OK")

// TargetFunc is the function the user implements
type TargetFunc func(context.Context, Gopher) error

// Runner is the interface wrapping tools the user may want to run
// Ex: go build or go fmt
type Runner interface {
	Run(context.Context, RunArgs) error
}

type RunArgs struct {
	GoBin  string
	Stdout io.Writer
}

type RunEvent iter.Seq[any]

func Register(name string, f TargetFunc, description string) {
	targets[name] = targetEntry{f, description}
}

var targets = map[string]targetEntry{}

type targetEntry struct {
	Func        TargetFunc
	Description string
}

type CMD struct {
	Target     string `arg:"" default:"default" help:"Recipe to run."`
	List       bool   `short:"l" help:"List all targets then exit."`
	Go         Gopher `embed:"" group:"Golang Flags"`
	GopherDir  string `default:".gopher" help:"Directory to cache files gopher creates."`
	GopherFile string `short:"C" default:"gopher.go" help:"File to read from. If gopher.go is not found, defaults to using examples/default.go. (See source code)"`
}

func (config *CMD) Run(ctx context.Context, logger *slog.Logger) error {
	config.Go.Logger = logger
	_, err := config.Go.Load(config.GopherFile, config.GopherDir)
	if err != nil {
		return err
	}
	// config.Gopher.logger = logger
	// if err := config.LoadGopherFile(logger); err != nil {
	// }
	// TODO: Run generated code instead of this
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
