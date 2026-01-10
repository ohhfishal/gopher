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
	"strings"
	"syscall"
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

	// Never use this flag in production. It execs to ./gopher and is thus an attack vector
	Dev bool `hidden:"" env:"GOPHER_DEV" help:"Turn on gopher developer tools. Let's gopher rebuild and exec itself."`
}

func (config *RunCMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	if config.Dev {
		devLog(stdout, logger)
	}
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
		if config.Dev {
			devExit(stdout, err)
		}
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

func devExit(stdout io.Writer, err error) {
	/*
		Dev exit is here to allow hot swapping of the gopher binary using gopher while developing it.
		I can not think of a non-malicious reason to use this feature otherwise.
		Hence protecting it via an ENV/Flag and prohibiting it for any version of gopher not
		built via go run or in a dirty git repo.
		Should this be insufficient, open an issue but I only ever expect to remove this.
	*/
	if state, ok := err.(*exec.ExitError); ok {
		version := cache.Version()
		if version != "(devel)" && !strings.Contains(version, "dirty") {
			panic("attempted source code dev-exit in tagged version")
		} else if state.ProcessState.ExitCode() == 42 {
			msg := "Dev exit used by gopherfile. Execing into newer gopher build. Hope you know what you are doing"
			pretty.Fwarnln(stdout, msg)
			time.Sleep(2 * time.Second)
			panic(syscall.Exec("./gopher", os.Args, os.Environ()))
		}
	}
}

func devLog(stdout io.Writer, logger *slog.Logger) {
	devLogger := logger.With("$GOPHER_DEV", os.Getenv("GOPHER_DEV"), "args", os.Args)
	devLogger.Error("DEV MODE ENABLED, THIS SHOULD ONLY BE DONE WHILE DEVLOPING GOPHER SOURCE")
	pretty.Fwarnln(stdout, "----- DEV MODE ENABLED! -----")
	pretty.Fwarnln(stdout, "If that was not intentional, please kill me.")
	pretty.Fwarnln(stdout, "A flag/env was set designed specifcally to develop gopher source code.")
	pretty.Fwarnln(stdout, "I blindly exec ./gopher and thus *arbitrary code on your machine*.")
	pretty.Fwarnln(stdout, "-----------------------------")
	time.Sleep(2 * time.Second)
}
