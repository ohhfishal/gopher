package runner

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
)

// TODO: Hook interfaces? Let runners define an Init, Run and Close methods

var ErrSkip = errors.New("stop and skip iteration")

type GoConfig struct {
	GoBin string `default:"go" help:"Go binary to use for commands."`
}

type RunFunc func(context.Context, RunArgs) error

// Runner is the interface wrapping tools the user may want to run
// Ex: go build or go fmt
type Runner interface {
	Run(context.Context, RunArgs) error
}

type runner struct {
	f RunFunc
}

func (r *runner) Run(ctx context.Context, args RunArgs) error {
	return r.f(ctx, args)
}

func RunnerFunc(f RunFunc) Runner {
	return &runner{
		f: f,
	}
}

type RunArgs struct {
	GoConfig GoConfig
	Stdout   io.Writer
}

func Run(ctx context.Context, event RunEvent, runners ...Runner) error {
	var gopher Gopher
	return gopher.Run(ctx, event, runners...)
}

type Gopher struct {
}

func (gopher *Gopher) Run(ctx context.Context, event RunEvent, runners ...Runner) error {
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
