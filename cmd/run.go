package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ohhfishal/gopher/cache"
	"github.com/ohhfishal/gopher/compile"
	"github.com/ohhfishal/gopher/pretty"
	"github.com/ohhfishal/gopher/runtime"
)

type RunCMD struct {
	Target         string           `arg:"" default:"default" help:"Recipe to run."`
	List           bool             `short:"l" help:"List all targets then exit."`
	Compile        bool             `help:"Only run the gopher compile then exit without running the target."`
	DisableHotswap bool             `help:"Disable restarting if the gopherfile changes while running."`
	GoConfig       runtime.GoConfig `embed:"" group:"Golang Flags"`
	GopherFile     string           `short:"C" default:"gopher.go" help:"File to read from. If gopher.go is not found, defaults to using examples/default.go. (See source code)"`
	GopherDir      string           `kong:"-"`
}

func (config *RunCMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if config.DisableHotswap {
		return config.run(ctx, stdout, logger)
	}

	for {
		runCtx := cache.WithFileCancel(ctx, config.GopherFile)

		err := config.run(runCtx, stdout, logger)

		// Check if we got an error because gopherfile changed
		cause := context.Cause(runCtx)
		if ctx.Err() == nil && errors.Is(cause, cache.ErrFileChanged) {
			time.Sleep(125 * time.Millisecond)
			pretty.Fwarnf(stdout, "Need to restart: %s\n", cause.Error())
			logger.Info("gopherfile changed, restarting",
				slog.Any("cause", cause),
				slog.String("gopherfile", config.GopherFile),
			)
			time.Sleep(125 * time.Millisecond)
			continue
		}
		return err
	}
}

func (config *RunCMD) run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	if err := buildGopherIfNeeded(stdout, config.GopherFile, config.GopherDir, config.GoConfig.GoBin, config.Compile); err != nil {
		return err
	}

	if config.Compile {
		return nil
	}

	var args []string
	if config.List {
		args = append(args, "-l")
	}
	args = append(args, config.Target)

	path := filepath.Join(config.GopherDir, compile.BinaryName)

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	slog.Debug("running target", "path", path, "args", args)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func buildGopherIfNeeded(stdout io.Writer, file string, directory string, goBin string, force bool) error {
	reader, err := GopherFile(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	if !force {
		ok, err := cache.Valid(file, directory, goBin)
		if err != nil {
			return fmt.Errorf("determining if cached: %w", err)
		}

		if ok {
			slog.Debug("don't need to compile, using cached")
			return nil
		}
	}
	slog.Debug("needs to compile, compiling")

	// TODO: Don't hardcode this
	printer := pretty.New(stdout, "Gopher compiler")
	if err := printer.Start(); err != nil {
		return fmt.Errorf("printing start message: %w", err)
	}
	if err := compile.Compile(printer, reader, directory, goBin); err != nil {
		printer.Done(err)
		return fmt.Errorf("compiling: %w", err)
	}
	return printer.Done(nil)
}

func GopherFile(filepath string) (io.ReadCloser, error) {
	file, err := os.Open(filepath)
	if err != nil {
		msg := `You can run "gopher bootstrap" to get started`
		return nil, fmt.Errorf("could not open %s: %w\n%s", filepath, err, msg)
	}
	return file, nil
}
