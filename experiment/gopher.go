package main

import (
	"context"
	"fmt"
	"iter"
	"io"
	"os"
	"time"
)

type Gopher struct {
	GoBin string
}

type RunArgs struct {
	GoBin string
	Stdout io.Writer
}

type RunEvent iter.Seq[any]


type Runner interface {
	Run(context.Context, RunArgs) error
}
var _ Runner = &GoBuild{}

func main() {
	gopher := Gopher{
		GoBin: "go",
	}
	Devel(context.TODO(), gopher)
}


func (gopher *Gopher) Run(ctx context.Context, event RunEvent, runners ...Runner) error {
	for range event {
		for _, runner := range runners {
			if err := runner.Run(ctx, RunArgs{
				GoBin: gopher.GoBin,
				Stdout: os.Stdout,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

type GoBuild struct {
	Output string
	Flags []string
	Packages []string
}

func (build *GoBuild) Run(ctx context.Context, args RunArgs) error {
	fmt.Fprintln(args.Stdout, "building")
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
