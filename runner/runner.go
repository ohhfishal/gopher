package runner

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"time"
)

var ErrOK = errors.New("OK")

type GoConfig struct {
	GoBin string `default:"go" help:"Go binary to use for commands."`
}

type RunEvent iter.Seq[any]
type RunFunc func(context.Context, RunArgs) error

// Runner is the interface wrapping tools the user may want to run
// Ex: go build or go fmt
type Runner interface {
	Run(context.Context, RunArgs) error
}

func RunnerFunc(f RunFunc) Runner {
	panic("not implemented: runnerfunc")
	return nil
}

type RunArgs struct {
	GoConfig GoConfig
	Stdout   io.Writer
}

func Run(ctx context.Context, event RunEvent, runners ...Runner) error {
	for range event {
		for _, runner := range runners {
			if ctx.Err() != nil {
				return nil
			}

			err := runner.Run(ctx, RunArgs{
				// GoBin: gopher.GoBin,
				// TODO: FIX HACK
				GoConfig: GoConfig{
					GoBin: "go",
				},
				Stdout: os.Stdout,
			})

			if errors.Is(ErrOK, err) {
				// TODO: ????
				// Eventually print Go Build: OK
				fmt.Println("OK")
			} else if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func NowAnd(when RunEvent) RunEvent {
	return func(yield func(any) bool) {
		for range Now() {
			if !yield(nil) {
				break
			}
		}
		for range when {
			if !yield(nil) {
				return
			}
		}
	}
}

func Now() RunEvent {
	return func(yield func(_ any) bool) {
		_ = yield(nil)
	}
}

func Every(duration time.Duration) RunEvent {
	ticker := time.NewTicker(duration)
	return func(yield func(_ any) bool) {
		defer ticker.Stop()
		for range ticker.C {
			if !yield(nil) {
				return
			}
		}
	}
}
