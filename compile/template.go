//go:build gopher

/*
The gopher build tag should only be used by the gopher compiler. Otherwise it will break.
This file should only be modified in its source in the gopher source code.

If you are reading this in a .gopher directory,
DO NOT EDIT THIS IT WILL GET OVERWRITTEN.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	. "github.com/ohhfishal/gopher/runtime"
	"io"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
)

type Target struct {
	Name        string
	Description string
	Func        func(context.Context, *Gopher) error
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	if err := Main(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func Main(ctx context.Context, stdout io.Writer, args []string) error {
	if len(args) < 1 {
		return errors.New(`missing argument: "<target>"`)
	}
	if args[0] == "-l" {
		PrintTargets()
		return nil
	}
	if target, ok := targets[args[0]]; ok {
		return target.Func(ctx, &Gopher{
			GoConfig: GoConfig{
				GoBin: "go",
			},
			Stdout: os.Stdout,
			Target: target.Name,
		})
	}
	return fmt.Errorf("unknown target: %s", args[0])
}

func PrintTargets() {
	fmt.Println("Targets:")
	keys := slices.Collect(maps.Keys(targets))
	slices.Sort(keys)
	for _, name := range keys {
		target := targets[name]
		name = strings.ToLower(name)
		fmt.Printf("  %8s: %s\n", name, strings.ReplaceAll(target.Description, "\n", "\n"+strings.Repeat(" ", max(len(name), 8)+4)))
	}
}
