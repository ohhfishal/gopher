// Package runtime provides methods for easily creating gopherfiles.
//
// Runtime has pre-configured runners that make it easy
// to run common Go tooling as well as methods for periodicaly running a target.
// (Such as for repeatably running go build while developing.)
//
// # Available Runners
//
// Standard Go Tooling: [GoTest], [GoVet], [GoBuild] [GoFormat]
//
// Quality of life: [Printer]
package runtime

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
)

// TODO: Hook interfaces? Let runners define an Init, Run and Close methods

// Sentinel error to notify the caller to stop break the run loop until the next [Event].
var ErrSkip = errors.New("stop and skip iteration")

type GoConfig struct {
	GoBin string `default:"go" help:"Go binary to use for commands."`
}

/*
Runners wrap a method to be called in a [Gopher.Run] event loop.
Ex: go build or go fmt
*/
type Runner interface {
	Run(context.Context, RunArgs) error
}

type runner struct {
	f func(context.Context, RunArgs) error
}

func (r *runner) Run(ctx context.Context, args RunArgs) error {
	return r.f(ctx, args)
}

/*
Converts a function to a [Runner]
*/
func RunnerFunc(f func(context.Context, RunArgs) error) Runner {
	return &runner{
		f: f,
	}
}

/*
Arguments provided to [Runner]s when called at runtime.
*/
type RunArgs struct {
	GoConfig GoConfig
	Stdout   io.Writer
}

/*
Calls [Gopher.Run] on the default [Gopher] instance.
*/
func Run(ctx context.Context, event Event, runners ...Runner) error {
	var gopher Gopher
	return gopher.Run(ctx, event, runners...)
}

type Gopher struct {
}

/*
Calls all runners sequentially when event triggers. Any runners that implement [Init] have the method called.
*/
func (gopher *Gopher) Run(ctx context.Context, event Event, runners ...Runner) error {
	for range event {
		if ctx.Err() != nil {
			return nil
		}
		runCtx, cancel := context.WithCancel(ctx)
		gopher.run(runCtx, runners...)
		// TODO: This may need to be canceled *at the start* of the next iteration
		cancel()
	}
	return nil
}

func (gopher *Gopher) run(ctx context.Context, runners ...Runner) {
	for _, runner := range runners {
		err := runner.Run(ctx, RunArgs{
			// GoBin: gopher.GoBin,
			// TODO: FIX HACK
			GoConfig: GoConfig{
				GoBin: "go",
			},
			Stdout: os.Stdout,
		})
		if errors.Is(ErrSkip, err) {
			return
		} else if err != nil {
			fmt.Fprintln(os.Stdout, err)
			return
		}
	}
}
