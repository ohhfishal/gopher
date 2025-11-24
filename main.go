package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/ohhfishal/gopher/report"
	konghelp "github.com/ohhfishal/kong-help"
	// "github.com/ohhfishal/gopher/watch"
)

//go:embed version.txt
var version string

var ErrDone = errors.New("program ready to exit")

type Cmd struct {
	Version   VersionFlag `short:"v" help:"Print out version and exit."`
	LogConfig LogConfig   `embed:"" group:"Logging Flags:"`
	Report    report.CMD  `cmd:"" group:"" default:"withargs" help:"Output build report. Pipe in output from go build -json."`
	Debug     bool        `help:"Turn on debugging features."`
	// Watch     watch.CMD  `cmd:"" help:"Watch for changes and rebuild"`
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
		konghelp.Help(),
		kong.BindTo(ctx, new(context.Context)),
		kong.BindTo(stdout, new(io.Writer)),
		kong.Vars{
			"version": version,
		},
	)
	if err != nil {
		return err
	}

	parser.Stdout = stdout
	parser.Stderr = stdout

	context, err := parser.Parse(
		os.Args[1:],
	)
	if errors.Is(err, ErrDone) {
		return nil
	} else if err != nil || exit {
		return err
	}

	if cmd.Debug {
		cmd.LogConfig.Level = slog.LevelDebug
	}

	logger := cmd.LogConfig.NewLogger(stdout)
	if err := context.Run(logger); err != nil {
		// TODO: Handle some of the run options
		logger.Error("failed to run", "error", err)
		return nil
	}
	return nil
}

type VersionFlag bool

func (v VersionFlag) BeforeReset(app *kong.Kong, vars kong.Vars) error {
	if _, err := fmt.Fprint(app.Stdout, vars["version"]); err != nil {
		return err
	}
	return fmt.Errorf("%s: %w", "printed version", ErrDone)
}
