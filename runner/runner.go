package runner

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"iter"
	"log/slog"
)

var ErrOK = errors.New("OK")

type Config struct {
	// TODO: I don't think this goes here? Gopher is the compiler?
	GoBin  string       `default:"go" help:"Go binary to use for commands."`
	Logger *slog.Logger `kong:"-"`
}

// TargetFunc is the function the user implements
type TargetFunc func(context.Context, Config) error

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
