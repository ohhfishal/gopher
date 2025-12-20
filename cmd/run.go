package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ohhfishal/gopher/cache"
	"github.com/ohhfishal/gopher/compile"
	"github.com/ohhfishal/gopher/example"
	"github.com/ohhfishal/gopher/pretty"
	"github.com/ohhfishal/gopher/runner"
)

type RunCMD struct {
	Target     string          `arg:"" default:"default" help:"Recipe to run."`
	List       bool            `short:"l" help:"List all targets then exit."`
	Compile    bool            `help:"Force the gopher compiler to run regardless of cache state."`
	GoConfig   runner.GoConfig `embed:"" group:"Golang Flags"`
	GopherFile string          `short:"C" default:"gopher.go" help:"File to read from. If gopher.go is not found, defaults to using examples/default.go. (See source code)"`
	GopherDir  string          `kong:"-"`
}

func (config *RunCMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	if err := buildGopherIfNeeded(stdout, config.GopherFile, config.GopherDir, config.GoConfig.GoBin, config.Compile); err != nil {
		return err
	}

	// Following convention of passing in invokation cmd
	args := []string{"./" + compile.BinaryName}
	if config.List {
		args = append(args, "-l")
	}
	args = append(args, config.Target)

	cmd := &exec.Cmd{
		Path:   filepath.Join(config.GopherDir, compile.BinaryName),
		Stdout: stdout,
		Stderr: stdout,
		Args:   args,
	}

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
		// TODO: Probably should print a better error like: No gopher file found run "gopher bootstrap" to get started
		if filepath == DefaultFilePath {
			slog.Info("could not find gopher.go, using default", "err", err)
			return io.NopCloser(strings.NewReader(example.DefaultGopherFile)), nil
		}
		return nil, fmt.Errorf("failed to open file: %s: %w", filepath, err)
	}
	return file, nil
}
