package cmd

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/ohhfishal/gopher/cache"
	konghelp "github.com/ohhfishal/kong-help"
)

var ErrDone = errors.New("program ready to exit")
var ErrNeedsCompile = errors.New("needs to compile gopherfile")

const DefaultFilePath = "gopher.go"

type Executable func() ([]byte, error)

func Run(ctx context.Context, stdout io.Writer, args []string) error {
	var exit bool
	var cmd CMD
	parser, err := kong.New(
		&cmd,
		kong.Exit(func(_ int) { exit = true }),
		konghelp.Help(),
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
	if errors.Is(err, ErrDone) {
		return nil
	} else if err != nil || exit {
		return err
	}

	if cmd.Debug {
		cmd.LogConfig.Level = slog.LevelDebug
		// TODO: Have this maybe go to a gopher_debug.log??
	}

	logger := cmd.LogConfig.NewLogger(stdout)
	if err := context.Run(logger); err != nil {
		return err
		// logger.Error("failed to run", "error", err)
		// return nil
	}
	return nil
}

type CMD struct {
	LogConfig LogConfig    `embed:"" group:"Logging Flags:"`
	Debug     bool         `help:"Turn on debugging features."`
	Version   cache.CMD    `cmd:"" help:"Print gopher veresion then exit."`
	Run       RunCMD       `cmd:"" default:"withargs" help:"Run a given target from a gopher.go file."`
	Bootstrap BootstrapCMD `cmd:"" help:"Bootstrap a project to use gopher."`
}
