package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/ohhfishal/gopher/watch"
)

type Cmd struct {
	LogConfig LogConfig       `embed:""`
	Report    watch.ReportCMD `cmd:"" default:"withargs" help:"Output build report."`
	Watch     watch.CMD       `cmd:"" help:"Watch for changes and rebuild"`
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	if err := Run(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func Run(ctx context.Context, stdout io.Writer, args []string) error {
	var exit bool
	var cmd Cmd
	parser, err := kong.New(
		&cmd,
		kong.Exit(func(_ int) { exit = true }),
		kong.BindTo(ctx, new(context.Context)),
		kong.BindTo(stdout, new(io.Writer)),
	)
	if err != nil {
		return err
	}

	parser.Stdout = stdout
	parser.Stderr = stdout

	context, err := parser.Parse(
		os.Args[1:],
	)
	if err != nil || exit {
		return err
	}

	logger := cmd.LogConfig.NewLogger(stdout)
	if err := context.Run(logger); err != nil {
		// TODO: Handle some of the run options
		logger.Error("failed to run", "error", err)
		return nil
	}
	return nil
}
